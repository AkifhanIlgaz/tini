package playlist

import (
	"errors"

	"github.com/AkifhanIlgaz/tini/internal/platform/youtube"
	"github.com/AkifhanIlgaz/tini/internal/shared/htmx"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ListItemsRequest is PlaylistService.ListItems's input — VenueID comes from the
// session, never from a submitted form.
type ListItemsRequest struct {
	VenueID bson.ObjectID `form:"-"`
}

func (r ListItemsRequest) Validate() error {
	if r.VenueID.IsZero() {
		return errors.New("playlist: venue id is required")
	}

	return nil
}

// AddLinkRequest is PlaylistHandler.Add's input — a single YouTube link submitted
// through the top-of-page form.
type AddLinkRequest struct {
	VenueID    bson.ObjectID `form:"-"`
	URL        string        `form:"url"`
	VideoID    string        `form:"-"`
	PlaylistID string        `form:"-"`
}

func (r *AddLinkRequest) Validate() error {
	if r.VenueID.IsZero() {
		return errors.New("playlist: venue id is required")
	}

	errs := htmx.FieldErrors{}

	if r.URL == "" {
		errs["url"] = ErrURLRequired
	}

	parsed, err := youtube.ParseURL(r.URL)
	if err != nil {
		errs["url"] = err
	}

	if len(errs) == 0 {
		r.VideoID = parsed.VideoID
		r.PlaylistID = parsed.PlaylistID

		return nil
	}

	return errs
}

// DeleteItemRequest is PlaylistHandler.Delete's input — ID comes from the route
// param, VenueID from the session.
type DeleteItemRequest struct {
	VenueID bson.ObjectID `form:"-"`
	ID      bson.ObjectID `form:"-"`
}

func (r DeleteItemRequest) Validate() error {
	if r.VenueID.IsZero() {
		return errors.New("playlist: venue id is required")
	}
	if r.ID.IsZero() {
		return errors.New("playlist: item id is required")
	}

	return nil
}
