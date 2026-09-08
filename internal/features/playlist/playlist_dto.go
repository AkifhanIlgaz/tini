package playlist

import (
	"errors"

	"github.com/AkifhanIlgaz/tini/internal/platform/youtube"
	"github.com/AkifhanIlgaz/tini/internal/shared/htmx"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// ListItemsRequest is PlaylistService.ListItems's input — VenueID comes from
// the session, never from a submitted form. Page/PageSize are 1-indexed and
// assumed already normalized (see PlaylistHandler.List's parsePage/
// parsePageSize).
type ListItemsRequest struct {
	VenueID  bson.ObjectID `form:"-"`
	Page     int           `form:"-"`
	PageSize int           `form:"-"`
}

func (r ListItemsRequest) Validate() error {
	if r.VenueID.IsZero() {
		return errors.New("playlist: venue id is required")
	}

	return nil
}

// Skip is this page's Mongo offset (options.Find().SetSkip's input).
func (r ListItemsRequest) Skip() int64 {
	return int64((r.Page - 1) * r.PageSize)
}

// AddLinkRequest is PlaylistHandler.Add's input — a single YouTube link submitted
// through the top-of-page form.
type AddLinkRequest struct {
	VenueID bson.ObjectID `form:"-"`
	AddedBy bson.ObjectID `form:"-"`
	URL     string        `form:"url"`
	// AsPlaylist is the "Playlist olarak ekle" checkbox — only consulted
	// when URL points at both a video and a playlist (ör.
	// "?v=...&list=..."); otherwise URL alone decides. Left unchecked, an
	// ambiguous link adds the single video.
	AsPlaylist bool   `form:"asPlaylist"`
	VideoID    string `form:"-"`
	PlaylistID string `form:"-"`
}

func (r *AddLinkRequest) Validate() error {
	if r.VenueID.IsZero() {
		return errors.New("playlist: venue id is required")
	}

	if r.URL == "" {
		return htmx.FieldErrors{"url": ErrURLRequired}
	}

	parsed, err := youtube.ParseURL(r.URL)
	if err != nil {
		return htmx.FieldErrors{"url": err}
	}

	switch {
	case parsed.VideoID != "" && parsed.PlaylistID != "" && r.AsPlaylist:
		r.PlaylistID = parsed.PlaylistID
	case parsed.VideoID != "":
		r.VideoID = parsed.VideoID
	default:
		r.PlaylistID = parsed.PlaylistID
	}

	return nil
}

// IsPlaylist reports whether Validate resolved this request to a playlist
// import (PlaylistService.ImportPlaylist) rather than a single-video add
// (PlaylistService.AddItem).
func (r AddLinkRequest) IsPlaylist() bool {
	return r.PlaylistID != ""
}

// NextTrackRequest is PlaylistService.NextTrack's input — advances playback
// to the item after CurrentYoutubeID in playlist order (created_at
// ascending), wrapping to the first item past the end. An empty
// CurrentYoutubeID (first load, or an item since deleted) starts from the
// beginning.
type NextTrackRequest struct {
	VenueID          bson.ObjectID `form:"-"`
	CurrentYoutubeID string        `form:"currentYoutubeId"`
}

func (r NextTrackRequest) Validate() error {
	if r.VenueID.IsZero() {
		return errors.New("playlist: venue id is required")
	}

	return nil
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
