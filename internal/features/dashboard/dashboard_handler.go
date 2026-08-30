// Package dashboard is the app's home feature: the Anasayfa overview page,
// plus placeholder pages for the sidebar links (Playlist, Oylama) that
// don't have their own feature yet.
package dashboard

import (
	"github.com/AkifhanIlgaz/tini/internal/features/dashboard/views"
	"github.com/AkifhanIlgaz/tini/internal/platform/csrf"
	"github.com/AkifhanIlgaz/tini/internal/platform/session"
	"github.com/AkifhanIlgaz/tini/internal/shared/layout"
	"github.com/AkifhanIlgaz/tini/internal/shared/middleware"
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	guard := middleware.AuthenticatedLayout()

	app.Get("/dashboard", guard, h.Dashboard)
	app.Get("/playlist", guard, h.page("Playlist"))
	app.Get("/oylama", guard, h.page("Oylama"))
}

func (h *Handler) Dashboard(c fiber.Ctx) error {
	u, _ := session.GetCurrentUser(c)

	return render(c, views.Dashboard(u, c.Path(), csrf.Token(c)))
}

// page returns a handler for one of the not-yet-built sidebar pages: title
// becomes both the <title> and the active breadcrumb.
func (h *Handler) page(title string) fiber.Handler {
	return func(c fiber.Ctx) error {
		u, _ := session.GetCurrentUser(c)

		trail := []layout.Crumb{{Label: "Anasayfa", Href: "/dashboard"}, {Label: title}}

		return render(c, views.Placeholder(u, c.Path(), csrf.Token(c), title, trail))
	}
}

func render(c fiber.Ctx, component templ.Component) error {
	c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	return component.Render(c.Context(), c.Response().BodyWriter())
}
