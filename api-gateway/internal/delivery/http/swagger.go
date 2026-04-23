package httpdelivery

import (
	"github.com/gofiber/fiber/v2"
)

// SwaggerAggregator — простейший Swagger UI на CDN с переключателем
// между тремя backend-сервисами. Реальные spec-файлы каждый сервис
// отдаёт сам у себя на /swagger/doc.json (см. swag init).
//
// Открывается на http://localhost:8080/swagger/.
const swaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>student-pm — API docs</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
  <style>
    body { margin: 0; font-family: -apple-system, BlinkMacSystemFont, sans-serif; }
    .topbar { background:#1b1b1b; color:#fff; padding:12px 20px; display:flex; align-items:center; gap:16px; }
    .topbar h1 { font-size:18px; margin:0; font-weight:600; }
    .topbar .links a { color:#9bd; margin-right:14px; text-decoration:none; font-size:14px; }
    .topbar .links a.active { color:#fff; font-weight:600; }
  </style>
</head>
<body>
  <div class="topbar">
    <h1>student-pm — API</h1>
    <div class="links" id="links"></div>
  </div>
  <div id="swagger-ui"></div>

  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-standalone-preset.js"></script>
  <script>
    const SERVICES = [
      { name: "auth-service",     url: "http://localhost:8081/swagger/doc.json" },
      { name: "groups-service",   url: "http://localhost:8082/swagger/doc.json" },
      { name: "projects-service", url: "http://localhost:8083/swagger/doc.json" },
    ];

    function render(target) {
      window.ui = SwaggerUIBundle({
        url: target.url,
        dom_id: "#swagger-ui",
        deepLinking: true,
        presets: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset],
        layout: "StandaloneLayout",
      });
      document.querySelectorAll("#links a").forEach(a => {
        a.classList.toggle("active", a.dataset.url === target.url);
      });
    }

    const wrap = document.getElementById("links");
    SERVICES.forEach((s, i) => {
      const a = document.createElement("a");
      a.textContent = s.name;
      a.href = "#" + s.name;
      a.dataset.url = s.url;
      a.onclick = e => { e.preventDefault(); render(s); };
      wrap.appendChild(a);
      if (i === 0) render(s);
    });
  </script>
</body>
</html>`

// SwaggerHandler — отдаёт HTML; если путь — /swagger/doc.json, проксирует
// в auth-service для удобства (gateway сам не выпускает spec).
func SwaggerHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// /swagger/* → /swagger/index.html, doc.json и т.п.
		// Чтобы не тащить весь UI, отдаём один HTML на любой суб-путь
		// и пусть JS внутри сам ходит на сервисы.
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(swaggerHTML)
	}
}
