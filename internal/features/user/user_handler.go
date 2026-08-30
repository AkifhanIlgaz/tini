package user

import (
	"errors"
	"fmt"

	"github.com/AkifhanIlgaz/tini/internal/features/user/views"
	"github.com/AkifhanIlgaz/tini/internal/platform/csrf"
	"github.com/AkifhanIlgaz/tini/internal/platform/session"
	"github.com/AkifhanIlgaz/tini/internal/shared/htmx"
	"github.com/AkifhanIlgaz/tini/internal/shared/middleware"
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	guard := middleware.AuthenticatedLayout(RoleBoss)

	app.Get("/users", guard, h.Users)
	app.Post("/users", guard, h.InviteAdmin)
	app.Post("/users/:id/delete", guard, h.DeleteAdmin)
}

func (h *Handler) Users(c fiber.Ctx) error {
	u, _ := session.GetCurrentUser(c)

	users, err := h.service.ListUsers(c.Context(), ListUsersRequest{VenueID: u.VenueID})
	if err != nil {
		return fmt.Errorf("user: users page: %w", err)
	}

	return render(c, views.Users(u, c.Path(), csrf.Token(c), toUserRows(users)))
}

// InviteAdmin pre-provisions an admin account by email — a bad email or
// one already in use fails softly with a toast, not a 500, since both are
// expected user-input mistakes rather than server errors.
func (h *Handler) InviteAdmin(c fiber.Ctx) error {
	u, _ := session.GetCurrentUser(c)

	var req InviteAdminRequest
	if err := c.Bind().Body(&req); err != nil {
		return fmt.Errorf("user: invite admin: bind: %w", err)
	}
	req.VenueID = u.VenueID

	if err := req.Validate(); err != nil {
		return htmx.Toast(c, htmx.ToastOptions{
			Title:   "Geçersiz e-posta",
			Variant: htmx.ToastDanger,
		})
	}

	if _, err := h.service.InviteAdmin(c.Context(), req); err != nil {
		if errors.Is(err, ErrUserAlreadyExists) {
			return htmx.Toast(c, htmx.ToastOptions{
				Title:   "Bu e-posta zaten kayıtlı",
				Variant: htmx.ToastDanger,
			})
		}

		return fmt.Errorf("user: invite admin: %w", err)
	}

	return htmx.Redirect(c, "/users")
}

// DeleteAdmin removes an admin — the boss's own row never carries a delete
// action (see views.UserRow.IsAdmin), and Repository.DeleteAdmin's filter
// only ever matches an admin belonging to the caller's own venue anyway.
func (h *Handler) DeleteAdmin(c fiber.Ctx) error {
	u, _ := session.GetCurrentUser(c)

	adminID, err := bson.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return fmt.Errorf("user: delete admin: parse id: %w", err)
	}

	req := RemoveAdminRequest{VenueID: u.VenueID, AdminID: adminID}
	if err := req.Validate(); err != nil {
		return fmt.Errorf("user: delete admin: %w", err)
	}

	if err := h.service.RemoveAdmin(c.Context(), req); err != nil {
		return fmt.Errorf("user: delete admin: %w", err)
	}

	return htmx.Redirect(c, "/users")
}

func toUserRows(users []User) []views.UserRow {
	rows := make([]views.UserRow, 0, len(users))
	for _, u := range users {
		rows = append(rows, views.UserRow{
			ID:        u.ID.Hex(),
			Name:      u.Name,
			Email:     u.Email,
			AvatarURL: u.AvatarURL,
			RoleLabel: roleLabel(u.Role),
			IsAdmin:   u.Role == RoleAdmin,
		})
	}

	return rows
}

func roleLabel(r Role) string {
	switch r {
	case RoleBoss:
		return "Patron"
	case RoleAdmin:
		return "Yönetici"
	default:
		return string(r)
	}
}

func render(c fiber.Ctx, component templ.Component) error {
	c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	return component.Render(c.Context(), c.Response().BodyWriter())
}
