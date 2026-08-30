// Package venue is the venue feature: for now just a placeholder "Mekan
// bilgileri" page — see venue_domain.go for why the repository/service
// aren't built out yet.
package venue

import (
	"github.com/AkifhanIlgaz/tini/internal/features/venue/views"
	"github.com/AkifhanIlgaz/tini/internal/platform/csrf"
	"github.com/AkifhanIlgaz/tini/internal/platform/session"
	"github.com/AkifhanIlgaz/tini/internal/shared/middleware"
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Get("/venue", middleware.AuthenticatedLayout(), h.Info)
}

func (h *Handler) Info(c fiber.Ctx) error {
	u, _ := session.GetCurrentUser(c)

	return render(c, views.Info(u, c.Path(), csrf.Token(c)))
}

func render(c fiber.Ctx, component templ.Component) error {
	c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	return component.Render(c.Context(), c.Response().BodyWriter())
}
