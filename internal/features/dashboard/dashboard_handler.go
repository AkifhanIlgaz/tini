// Package dashboard is the app's home feature: the Anasayfa overview page,
// plus a placeholder page for the sidebar links (Oylama) that don't have
// their own feature yet.
package dashboard

import (
	"github.com/AkifhanIlgaz/tini/internal/features/dashboard/views"
	"github.com/AkifhanIlgaz/tini/internal/platform/csrf"
	"github.com/AkifhanIlgaz/tini/internal/platform/session"
	"github.com/AkifhanIlgaz/tini/internal/shared/htmx"
	"github.com/AkifhanIlgaz/tini/internal/shared/layout"
	"github.com/AkifhanIlgaz/tini/internal/shared/middleware"
	"github.com/gofiber/fiber/v3"
)

type DashboardHandler struct{}

func NewHandler() *DashboardHandler {
	return &DashboardHandler{}
}

func (h *DashboardHandler) RegisterRoutes(app *fiber.App) {
	guard := middleware.AuthenticatedLayout()

	app.Get("/dashboard", guard, h.Dashboard)
	app.Get("/oylama", guard, h.page("Oylama"))
}

func (h *DashboardHandler) Dashboard(c fiber.Ctx) error {
	u, _ := session.GetCurrentUser(c)

	return htmx.Render(c, views.Dashboard(u, c.Path(), csrf.Token(c)))
}

// page returns a handler for one of the not-yet-built sidebar pages: title
// becomes both the <title> and the active breadcrumb.
func (h *DashboardHandler) page(title string) fiber.Handler {
	return func(c fiber.Ctx) error {
		u, _ := session.GetCurrentUser(c)

		trail := []layout.Crumb{{Label: "Anasayfa", Href: "/dashboard"}, {Label: title}}

		return htmx.Render(c, views.Placeholder(u, c.Path(), csrf.Token(c), title, trail))
	}
}
