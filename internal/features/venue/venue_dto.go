package venue

import (
	"errors"

	"github.com/AkifhanIlgaz/tini/internal/shared/htmx"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// GetVenueRequest is Service.GetVenue's input — VenueID comes from the
// session, never from a submitted form.
type GetVenueRequest struct {
	VenueID bson.ObjectID `form:"-"`
}

func (r GetVenueRequest) Validate() error {
	if r.VenueID.IsZero() {
		return errors.New("venue: venue id is required")
	}

	return nil
}

// UpdateVenueRequest is Service.UpdateVenue's input — the Mekan bilgileri
// page posts name and the round settings together from a single form (see
// views.settingsCard).
type UpdateVenueRequest struct {
	VenueID                   bson.ObjectID `form:"-"`
	Name                      string        `form:"name"`
	RoundIntervalMin          int           `form:"roundIntervalMin"`
	CandidateCount            int           `form:"candidateCount"`
	RecentlyPlayedCooldownMin int           `form:"recentlyPlayedCooldownMin"`
	CandidateCooldownMin      int           `form:"candidateCooldownMin"`
}

func (r UpdateVenueRequest) Validate() error {
	if r.VenueID.IsZero() {
		return errors.New("venue: venue id is required")
	}

	errs := htmx.FieldErrors{}

	if r.Name == "" {
		errs["name"] = ErrNameRequired
	}
	if r.RoundIntervalMin <= 0 {
		errs["roundIntervalMin"] = ErrRoundIntervalMinInvalid
	}
	if r.CandidateCount <= 0 {
		errs["candidateCount"] = ErrCandidateCountInvalid
	}
	if r.RecentlyPlayedCooldownMin <= 0 {
		errs["recentlyPlayedCooldownMin"] = ErrRecentlyPlayedCooldownMinInvalid
	}
	if r.CandidateCooldownMin <= 0 {
		errs["candidateCooldownMin"] = ErrCandidateCooldownMinInvalid
	}

	if len(errs) == 0 {
		return nil
	}

	return errs
}

// ToVenueSettings maps UpdateVenueRequest's settings fields into the domain
// VenueSettings the repository persists.
func (r UpdateVenueRequest) ToVenueSettings() VenueSettings {
	return VenueSettings{
		RoundIntervalMin:          r.RoundIntervalMin,
		CandidateCount:            r.CandidateCount,
		RecentlyPlayedCooldownMin: r.RecentlyPlayedCooldownMin,
		CandidateCooldownMin:      r.CandidateCooldownMin,
	}
}
