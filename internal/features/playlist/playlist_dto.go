package playlist

import (
	"github.com/AkifhanIlgaz/tini/internal/platform/youtube"
	"github.com/AkifhanIlgaz/tini/internal/shared/htmx"
)

// AddLinkRequest is Handler.Add's input — a single YouTube link submitted
// through the top-of-page form.
type AddLinkRequest struct {
	URL        string `form:"url"`
	VideoID    string `form:"-"`
	PlaylistID string `form:"-"`
}

func (r *AddLinkRequest) Validate() error {
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
