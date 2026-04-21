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

	// Фиксируем Director: подменяем хост, оставляем путь и query как есть.
	origDirector := rp.Director
	rp.Director = func(req *http.Request) {
		origDirector(req)
		req.Host = u.Host
		// Никаких mutation тела/пути — gateway работает только префиксом.
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

// Handler — Fiber-handler, который проксирует запрос на upstream.
// Перед прокси добавляет служебные заголовки X-Request-ID, X-User-ID, X-User-Role
// (если они уже были установлены middleware'ами выше).
func (p *Proxy) Handler() fiber.Handler {
	stdHandler := adaptor.HTTPHandler(p.rp)
	return func(c *fiber.Ctx) error {
		// Прокидываем атрибуты, которые удобно иметь сервисам в логах.
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

// Mount монтирует группу path → proxy.
// path — префикс вида "/auth", "/groups", "/projects", "/tasks", "/users".
// Все методы и подпути идут в upstream без модификации.
func Mount(app *fiber.App, prefix string, p *Proxy) {
	app.All(prefix, p.Handler())
	app.All(prefix+"/*", p.Handler())
}

// MountUnreachable — заглушка на случай, если апстрим не настроен,
// чтобы клиент получал понятную 502 вместо 404.
func MountUnreachable(app *fiber.App, prefix, reason string) {
	app.All(prefix, func(c *fiber.Ctx) error { return httperr.BadGateway(c, reason) })
	app.All(prefix+"/*", func(c *fiber.Ctx) error { return httperr.BadGateway(c, reason) })
}
