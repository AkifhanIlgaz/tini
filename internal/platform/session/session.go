// Package session wires Fiber's session middleware to Redis-backed storage
// and defines the shape of what this app keeps in a session.
package session

import (
	"github.com/AkifhanIlgaz/tini/internal/config"
	"github.com/gofiber/fiber/v3"
	fsession "github.com/gofiber/fiber/v3/middleware/session"
	"github.com/gofiber/storage/redis/v3"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const userKey = "user"

// Role identifies what a logged-in user is allowed to do. It lives here
// rather than in the user feature package so that both session and
// internal/shared/middleware (which authorize requests by role) can depend
// on it without importing a feature — user.Role is a type alias to this
// (see internal/features/user/user_domain.go) so existing call sites
// (user.RoleBoss, etc.) are unaffected.
type Role string

const (
	RoleAdmin Role = "admin"
	RoleBoss  Role = "boss"
)

// IsValid reports whether r is one of the known Role values.
func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleBoss:
		return true
	default:
		return false
	}
}

// User is the snapshot stored in the session on login: enough to render
// nav/UI and authorize requests without a Mongo round-trip on every request.
// ID/VenueID mirror the Mongo _id/venue_id of the persisted user record.
type User struct {
	ID      bson.ObjectID
	VenueID bson.ObjectID
	Email   string
	Name    string
	Role    Role
}

func (u User) HasRole(role Role) bool {
	return u.Role == role
}

// NewStore builds the session store backed by Redis.
func NewStore(cfg config.Config) *fsession.Store {
	storage := redis.New(redis.Config{
		URL: cfg.Redis.URL,
	})

	store := fsession.NewStore(fsession.Config{
		Storage:           storage,
		IdleTimeout:       cfg.Session.Idle,
		CookieSecure:      cfg.IsProduction(),
		CookieHTTPOnly:    true,
		CookieSameSite:    "Lax",
		CookieSessionOnly: false,
	})

	store.RegisterType(User{})

	return store
}

// New returns the Fiber middleware that loads/saves the session on every request.
func New(store *fsession.Store) fiber.Handler {
	return fsession.New(fsession.Config{Store: store})
}

// Get returns the logged-in user snapshot for the current request, if any.
func GetCurrentUser(c fiber.Ctx) (User, bool) {
	sess := fsession.FromContext(c)
	if sess == nil {
		return User{}, false
	}

	u, ok := sess.Get(userKey).(User)

	return u, ok
}

// Login stores the user snapshot in the session, regenerating the session ID
// first to prevent session fixation.
func Login(c fiber.Ctx, u User) error {
	sess := fsession.FromContext(c)
	if sess == nil {
		return fsession.ErrEmptySessionID
	}

	if err := sess.Regenerate(); err != nil {
		return err
	}

	sess.Set(userKey, u)

	return nil
}

// Logout destroys the current session.
func Logout(c fiber.Ctx) error {
	sess := fsession.FromContext(c)
	if sess == nil {
		return nil
	}

	return sess.Destroy()
}
