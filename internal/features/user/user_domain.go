// Package user is the user feature slice: the persisted user record and its
// Mongo-backed repository.
package user

import (
	"time"

	"github.com/AkifhanIlgaz/tini/internal/platform/session"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Role is an alias to session.Role — see that package for why the type
// lives there instead of here (breaking an import cycle between user and
// session/middleware, which both need to know about roles).
type Role = session.Role

const (
	RoleAdmin = session.RoleAdmin
	RoleBoss  = session.RoleBoss
)

// User is the persisted record for an account, keyed by Email (the stable
// identifier Google OAuth gives us).
type User struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	VenueID   bson.ObjectID `bson:"venue_id"`
	Email     string        `bson:"email"`
	Name      string        `bson:"name"`
	AvatarURL string        `bson:"avatar_url"`
	Role      Role          `bson:"role"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at"`
}
