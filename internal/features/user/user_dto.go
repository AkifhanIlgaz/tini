package user

import (
	"errors"
	"net/mail"

	"github.com/AkifhanIlgaz/tini/internal/shared/htmx"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// InviteAdminRequest is Service.InviteAdmin's input — a boss only supplies
// an email for the venue they own; the invited admin's name/avatar are
// unknown until their first Google login. VenueID never comes from the
// client.
type InviteAdminRequest struct {
	VenueID bson.ObjectID `form:"-"`
	Email   string        `form:"email"`
}

func (r InviteAdminRequest) Validate() error {
	errs := htmx.FieldErrors{}

	if _, err := mail.ParseAddress(r.Email); err != nil {
		errs["email"] = ErrEmailInvalid
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}

// ToUser maps an InviteAdminRequest into the domain User the repository
// persists.
func (r InviteAdminRequest) ToUser() User {
	return User{
		Email:   r.Email,
		Role:    RoleAdmin,
		VenueID: r.VenueID,
	}
}

// RemoveAdminRequest is Service.RemoveAdmin's input — both fields come from
// the session/URL, never from a submitted form.
type RemoveAdminRequest struct {
	VenueID bson.ObjectID `form:"-"`
	AdminID bson.ObjectID `form:"-"`
}

func (r RemoveAdminRequest) Validate() error {
	if r.VenueID.IsZero() {
		return errors.New("user: venue id is required")
	}
	if r.AdminID.IsZero() {
		return errors.New("user: admin id is required")
	}

	return nil
}

// ListUsersRequest is Service.ListUsers's input — VenueID comes from the
// session, never from a submitted form.
type ListUsersRequest struct {
	VenueID bson.ObjectID `form:"-"`
}

func (r ListUsersRequest) Validate() error {
	if r.VenueID.IsZero() {
		return errors.New("user: venue id is required")
	}

	return nil
}
