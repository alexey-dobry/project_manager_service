package proxy

import (
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/rs/zerolog"

	httperr "github.com/student-pm/api-gateway/internal/pkg/errors"
)

// Proxy — обёртка над httputil.ReverseProxy.
type Proxy struct {
	target *url.URL
	rp     *httputil.ReverseProxy
	log    zerolog.Logger
}

// New — создаёт proxy на upstream-URL вида "http://auth-service:8081".
func New(upstream string, timeout time.Duration, log zerolog.Logger) (*Proxy, error) {
	u, err := url.Parse(upstream)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, errors.New("invalid upstream URL: " + upstream)
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   20,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   3 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: timeout,
	}

	rp := httputil.NewSingleHostReverseProxy(u)
	rp.Transport = transport

	origDirector := rp.Director
	rp.Director = func(req *http.Request) {
		origDirector(req)
		req.Host = u.Host
	}

	rp.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Error().Err(err).
			Str("upstream", u.Host).
			Str("path", r.URL.Path).
			Msg("upstream error")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"error":{"code":"upstream_error","message":"upstream is unreachable"}}`))
	}

	return &Proxy{target: u, rp: rp, log: log}, nil
}

// Handler возвращает Fiber-handler, проксирующий запрос на upstream
// и пробрасывающий X-Request-ID, X-User-ID, X-User-Role из контекста.
func (p *Proxy) Handler() fiber.Handler {
	stdHandler := adaptor.HTTPHandler(p.rp)
	return func(c *fiber.Ctx) error {
		if id, ok := c.Locals("request_id").(string); ok && id != "" {
			c.Request().Header.Set("X-Request-ID", id)
		}
		if uid, ok := c.Locals("user_id").(string); ok && uid != "" {
			c.Request().Header.Set("X-User-ID", uid)
		}
		if role, ok := c.Locals("user_role").(string); ok && role != "" {
			c.Request().Header.Set("X-User-Role", role)
		}
		return stdHandler(c)
	}
}

// Mount регистрирует префикс и все его подпути на проксирование.
func Mount(app *fiber.App, prefix string, p *Proxy) {
	app.All(prefix, p.Handler())
	app.All(prefix+"/*", p.Handler())
}

// MountUnreachable монтирует префикс, отвечающий 502 для всех методов.
func MountUnreachable(app *fiber.App, prefix, reason string) {
	app.All(prefix, func(c *fiber.Ctx) error { return httperr.BadGateway(c, reason) })
	app.All(prefix+"/*", func(c *fiber.Ctx) error { return httperr.BadGateway(c, reason) })
}
