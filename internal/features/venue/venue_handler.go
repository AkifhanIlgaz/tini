// Package venue is the venue feature: for now just the "Mekan bilgileri"
// page's UI (QR kod + ayarlar formu) over mock data — see venue_domain.go
// for why the repository/service aren't built out yet.
package venue

import (
	"fmt"

	"github.com/AkifhanIlgaz/tini/internal/features/venue/views"
	"github.com/AkifhanIlgaz/tini/internal/platform/csrf"
	"github.com/AkifhanIlgaz/tini/internal/platform/session"
	"github.com/AkifhanIlgaz/tini/internal/shared/htmx"
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

	app.Get("/venue", guard, h.Info)
	app.Post("/venue/save", guard, h.Save)
}

// Info renders mock data — venue.Repository doesn't exist yet (see
// venue_domain.go), so there's nothing real to load until then.
func (h *Handler) Info(c fiber.Ctx) error {
	u, _ := session.GetCurrentUser(c)

	general := views.GeneralForm{
		Name: "Demo Mekan",
	}
	settings := views.SettingsForm{
		RoundIntervalMin:          5,
		CandidateCount:            4,
		RecentlyPlayedCooldownMin: 60,
		CandidateCooldownMin:      30,
	}

	return render(c, views.Info(u, c.Path(), csrf.Token(c), general, settings))
}

// Save only demonstrates the toast round-trip for now — nothing is
// actually persisted until venue.Repository exists.
func (h *Handler) Save(c fiber.Ctx) error {
	err := htmx.Toast(c, htmx.ToastOptions{
		Title:       "Mekan bilgileri kaydedildi",
		Description: "Değişikliklerin kaydedildi.",
		Variant:     htmx.ToastSuccess,
	})
	if err != nil {
		return fmt.Errorf("venue: save toast: %w", err)
	}

	return nil
}

func render(c fiber.Ctx, component templ.Component) error {
	c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	return component.Render(c.Context(), c.Response().BodyWriter())
}
