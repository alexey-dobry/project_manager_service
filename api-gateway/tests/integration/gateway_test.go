package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	gjwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	httpdelivery "github.com/student-pm/api-gateway/internal/delivery/http"
	jwtpkg "github.com/student-pm/api-gateway/internal/pkg/jwt"
	"github.com/student-pm/api-gateway/internal/proxy"
)

const testSecret = "test-secret-test-secret-1234567890"

func issueAccessToken(t *testing.T, userID uuid.UUID, role string) string {
	t.Helper()
	now := time.Now().UTC()
	tok := gjwt.NewWithClaims(gjwt.SigningMethodHS256, gjwt.MapClaims{
		"sub":  userID.String(),
		"role": role,
		"iss":  "test-auth",
		"iat":  now.Unix(),
		"exp":  now.Add(time.Minute).Unix(),
	})
	signed, err := tok.SignedString([]byte(testSecret))
	require.NoError(t, err)
	return signed
}

// fakeUpstream — простой http.ServeMux, имитирующий backend-сервис.
// Возвращает свой адрес.
func fakeUpstream(t *testing.T, hits *int, expectAuth bool, headersSeen *http.Header) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		*hits++
		if headersSeen != nil {
			*headersSeen = r.Header.Clone()
		}
		if expectAuth && r.Header.Get("Authorization") == "" {
			http.Error(w, `{"error":{"code":"no_auth","message":"x"}}`, http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true,"path":"` + r.URL.Path + `"}`))
	})
	return httptest.NewServer(mux)
}

func buildGateway(t *testing.T, authURL, groupsURL, projectsURL string) *fiber.App {
	t.Helper()
	log := zerolog.Nop()

	authP, err := proxy.New(authURL, 5*time.Second, log)
	require.NoError(t, err)
	groupsP, err := proxy.New(groupsURL, 5*time.Second, log)
	require.NoError(t, err)
	projectsP, err := proxy.New(projectsURL, 5*time.Second, log)
	require.NoError(t, err)

	v, err := jwtpkg.New(testSecret)
	require.NoError(t, err)

	app := fiber.New()
	app.Use(httpdelivery.RequestID())
	httpdelivery.RegisterRoutes(app, httpdelivery.Upstreams{
		Auth: authP, Groups: groupsP, Projects: projectsP,
	}, v, httpdelivery.SwaggerHandler())
	return app
}

func TestGateway_PublicAuthEndpointsPassThrough(t *testing.T) {
	hits := 0
	auth := fakeUpstream(t, &hits, false, nil)
	defer auth.Close()
	groups := fakeUpstream(t, new(int), false, nil)
	defer groups.Close()
	projects := fakeUpstream(t, new(int), false, nil)
	defer projects.Close()

	app := buildGateway(t, auth.URL, groups.URL, projects.URL)

	// /auth/register без токена должен пройти на upstream.
	req := httptest.NewRequest(http.MethodPost, "/auth/register", nil)
	resp, err := app.Test(req, 5000)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, 1, hits)
}

func TestGateway_ProtectedEndpoints_NoToken_401(t *testing.T) {
	hits := 0
	auth := fakeUpstream(t, &hits, false, nil)
	defer auth.Close()
	groups := fakeUpstream(t, &hits, false, nil)
	defer groups.Close()
	projects := fakeUpstream(t, &hits, false, nil)
	defer projects.Close()

	app := buildGateway(t, auth.URL, groups.URL, projects.URL)

	for _, path := range []string{"/users/me", "/groups", "/projects"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			resp, err := app.Test(req, 5000)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		})
	}
	require.Equal(t, 0, hits, "upstream не должен быть вызван без токена")
}

func TestGateway_PassesUserHeadersDownstream(t *testing.T) {
	hits := 0
	var seen http.Header
	auth := fakeUpstream(t, &hits, true, &seen)
	defer auth.Close()
	groups := fakeUpstream(t, new(int), true, nil)
	defer groups.Close()
	projects := fakeUpstream(t, new(int), true, nil)
	defer projects.Close()

	app := buildGateway(t, auth.URL, groups.URL, projects.URL)

	uid := uuid.New()
	token := issueAccessToken(t, uid, "student")

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req, 5000)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	require.Equal(t, "Bearer "+token, seen.Get("Authorization"))
	require.Equal(t, uid.String(), seen.Get("X-User-ID"), "gateway должен пробрасывать X-User-ID")
	require.Equal(t, "student", seen.Get("X-User-Role"))
	require.NotEmpty(t, seen.Get("X-Request-ID"))
}

func TestGateway_RoutesByPrefix(t *testing.T) {
	authHits, groupsHits, projectsHits := 0, 0, 0
	auth := fakeUpstream(t, &authHits, false, nil)
	defer auth.Close()
	groups := fakeUpstream(t, &groupsHits, false, nil)
	defer groups.Close()
	projects := fakeUpstream(t, &projectsHits, false, nil)
	defer projects.Close()

	app := buildGateway(t, auth.URL, groups.URL, projects.URL)
	tok := issueAccessToken(t, uuid.New(), "admin")

	hit := func(path string) {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		resp, err := app.Test(req, 5000)
		require.NoError(t, err)
		resp.Body.Close()
	}
	hit("/users/me")
	hit("/groups")
	hit("/groups/abc")
	hit("/projects/abc")
	hit("/tasks/abc/comments")

	require.Equal(t, 1, authHits)
	require.Equal(t, 2, groupsHits)
	require.Equal(t, 2, projectsHits) // /projects/abc + /tasks/abc/comments
}

func TestGateway_RootAndHealth(t *testing.T) {
	auth := fakeUpstream(t, new(int), false, nil)
	defer auth.Close()
	groups := fakeUpstream(t, new(int), false, nil)
	defer groups.Close()
	projects := fakeUpstream(t, new(int), false, nil)
	defer projects.Close()

	app := buildGateway(t, auth.URL, groups.URL, projects.URL)

	for _, path := range []string{"/", "/health", "/ready"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		resp, err := app.Test(req, 5000)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode, path)
	}
}
