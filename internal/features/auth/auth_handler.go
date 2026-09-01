package auth

import (
	"fmt"
	"net/http"

	"github.com/AkifhanIlgaz/tini/internal/config"
	"github.com/AkifhanIlgaz/tini/internal/features/auth/views"
	"github.com/AkifhanIlgaz/tini/internal/features/user"
	"github.com/AkifhanIlgaz/tini/internal/platform/csrf"
	"github.com/AkifhanIlgaz/tini/internal/platform/session"
	"github.com/AkifhanIlgaz/tini/internal/shared/htmx"
	"github.com/AkifhanIlgaz/tini/internal/shared/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

type AuthHandler struct {
	users user.UserRepository
}

func NewAuthHandler(users user.UserRepository) *AuthHandler {
	return &AuthHandler{users: users}
}

// RegisterProviders points gothic at its own short-lived cookie store for the
// OAuth handshake state — separate from the app's Redis-backed session store,
// since the handshake never outlives the redirect round-trip.
func (h *AuthHandler) RegisterProviders(cfg config.Config) {
	const handshakeMaxAge = 10 * 60 // the handshake never outlives the redirect round-trip

	store := sessions.NewCookieStore([]byte(cfg.Session.GothicSecret))
	store.MaxAge(handshakeMaxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = cfg.IsProduction()
	// gorilla/sessions defaults new cookie stores to SameSite=None, which
	// browsers silently drop unless Secure is also true — breaking the
	// handshake entirely over plain http in dev. Match the app's other
	// cookies (session, csrf) at Lax instead.
	store.Options.SameSite = http.SameSiteLaxMode

	gothic.Store = store

	// goth's Google provider defaults to the "email" scope alone when none
	// are given — without "profile" too, Google's userinfo response never
	// includes name/picture, leaving gothUser.Name/AvatarURL empty.
	goth.UseProviders(google.New(cfg.Google.ClientID, cfg.Google.ClientSecret, cfg.Google.CallbackURL, "email", "profile"))
}

func (h *AuthHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/login", middleware.UnauthenticatedLayout(), h.LoginPage)

	authRoute := app.Group("/auth")

	authRoute.Get("/:provider", h.Login)
	authRoute.Get("/:provider/callback", h.Callback)
	authRoute.Post("/logout", h.Logout)
}

func (h *AuthHandler) LoginPage(c fiber.Ctx) error {
	return htmx.Render(c, views.Login(csrf.Token(c)))
}

// Login and Callback take the provider from Fiber's :provider route param
// (c.Params) rather than r.PathValue: the request handed to the wrapped
// net/http handler is built by fasthttpadaptor.ConvertRequest, which never
// populates http.Request's ServeMux-only PathValue data. The provider name
// only reaches gothic through the query string it injects here.
func (h *AuthHandler) Login(c fiber.Ctx) error {
	provider := c.Params("provider")

	return adaptor.HTTPHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		q.Add("provider", provider)
		r.URL.RawQuery = q.Encode()

		gothic.BeginAuthHandler(w, r)
	})(c)
}

// Callback completes the OAuth handshake, then upserts the provider's user
// info into Mongo and logs the app in via its own Redis-backed session —
// entirely separate from gothic's short-lived handshake cookie, which
// gothic.CompleteUserAuth already clears on success.
func (h *AuthHandler) Callback(c fiber.Ctx) error {
	provider := c.Params("provider")

	var gothUser goth.User
	var authErr error

	err := adaptor.HTTPHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		q.Add("provider", provider)
		r.URL.RawQuery = q.Encode()

		gothUser, authErr = gothic.CompleteUserAuth(w, r)
	})(c)
	if err != nil {
		return err
	}
	if authErr != nil {
		return fmt.Errorf("auth: complete user auth: %w", authErr)
	}

	u, err := h.users.Upsert(c.Context(), gothUser.Email, gothUser.Name, gothUser.AvatarURL)
	if err != nil {
		return fmt.Errorf("auth: upsert user: %w", err)
	}

	sessUser := session.User{ID: u.ID, VenueID: u.VenueID, Email: u.Email, Name: u.Name, Role: u.Role}
	if err := session.Login(c, sessUser); err != nil {
		return fmt.Errorf("auth: login: %w", err)
	}

	// A CSRF token observed before authentication must not be reusable
	// after it — regenerate alongside the session.
	if err := csrf.Rotate(c); err != nil {
		return fmt.Errorf("auth: rotate csrf: %w", err)
	}

	return c.Redirect().To("/dashboard")
}

// Logout is a POST since it has a side effect (destroying the session), and
// so it's called via hx-post from the dashboard shell's "Çıkış yap" button.
func (h *AuthHandler) Logout(c fiber.Ctx) error {
	if err := session.Logout(c); err != nil {
		return fmt.Errorf("auth: logout: %w", err)
	}

	if err := csrf.Rotate(c); err != nil {
		return fmt.Errorf("auth: rotate csrf: %w", err)
	}

	return htmx.Redirect(c, "/login")
}
