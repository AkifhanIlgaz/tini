package venue

import (
	"errors"
	"fmt"

	"github.com/AkifhanIlgaz/tini/internal/features/venue/views"
	"github.com/AkifhanIlgaz/tini/internal/platform/csrf"
	"github.com/AkifhanIlgaz/tini/internal/platform/session"
	"github.com/AkifhanIlgaz/tini/internal/shared/htmx"
	"github.com/AkifhanIlgaz/tini/internal/shared/middleware"
	"github.com/gofiber/fiber/v3"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	guard := middleware.AuthenticatedLayout()

	app.Get("/venue", guard, h.Info)
	app.Post("/venue/save", guard, h.Save)
}

func (h *Handler) Info(c fiber.Ctx) error {
	u, _ := session.GetCurrentUser(c)

	req := GetVenueRequest{
		VenueID: u.VenueID,
	}
	if err := req.Validate(); err != nil {
		return fmt.Errorf("venue: info: validate: %w", err)
	}

	v, err := h.service.GetVenue(c.Context(), req)
	if err != nil {
		return fmt.Errorf("venue: info: %w", err)
	}

	general, settings := toViewForms(v)

	return htmx.Render(c, views.Info(u, c.Path(), csrf.Token(c), general, settings, nil))
}

func (h *Handler) Save(c fiber.Ctx) error {
	u, _ := session.GetCurrentUser(c)

	req := UpdateVenueRequest{
		VenueID: u.VenueID,
	}

	if err := c.Bind().Body(&req); err != nil {
		return fmt.Errorf("venue: save: bind: %w", err)
	}

	// A FieldErrors failure is bad form input (blank name, a non-positive
	// setting) — re-render the card with each failing field marked invalid
	// instead of the generic toast path. VenueID always comes from the
	// session, so any other Validate() failure is unexpected.
	if err := req.Validate(); err != nil {
		var fieldErrs htmx.FieldErrors
		if errors.As(err, &fieldErrs) {
			general := views.GeneralForm{Name: req.Name}
			settings := views.SettingsForm{
				RoundIntervalMin:          req.RoundIntervalMin,
				CandidateCount:            req.CandidateCount,
				RecentlyPlayedCooldownMin: req.RecentlyPlayedCooldownMin,
				CandidateCooldownMin:      req.CandidateCooldownMin,
			}

			return htmx.Render(c, views.Info(u, c.Path(), csrf.Token(c), general, settings, fieldErrs))
		}

		return fmt.Errorf("venue: save: validate: %w", err)
	}

	v, err := h.service.UpdateVenue(c.Context(), req)
	if err != nil {
		return fmt.Errorf("venue: save: %w", err)
	}

	if err := htmx.Toast(c, htmx.ToastOptions{
		Title:       "Mekan bilgileri kaydedildi",
		Description: "Değişikliklerin kaydedildi.",
		Variant:     htmx.ToastSuccess,
	}); err != nil {
		return fmt.Errorf("venue: save: toast: %w", err)
	}

	general, settings := toViewForms(v)

	return htmx.Render(c, views.Info(u, c.Path(), csrf.Token(c), general, settings, nil))
}

// toViewForms maps a domain Venue into the views package's own form
// view-models (see views.GeneralForm/SettingsForm's doc comments for why
// they're kept separate from the domain type).
func toViewForms(v Venue) (views.GeneralForm, views.SettingsForm) {
	general := views.GeneralForm{
		Name: v.Name,
	}
	settings := views.SettingsForm{
		RoundIntervalMin:          v.Settings.RoundIntervalMin,
		CandidateCount:            v.Settings.CandidateCount,
		RecentlyPlayedCooldownMin: v.Settings.RecentlyPlayedCooldownMin,
		CandidateCooldownMin:      v.Settings.CandidateCooldownMin,
	}

	return general, settings
}
