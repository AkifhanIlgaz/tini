// Package playlist is the playlist feature: an add-link form (single video
// or playlist import) plus a paginated table over a venue's playlist.
package playlist

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/AkifhanIlgaz/tini/internal/features/playlist/views"
	"github.com/AkifhanIlgaz/tini/internal/platform/csrf"
	"github.com/AkifhanIlgaz/tini/internal/platform/session"
	"github.com/AkifhanIlgaz/tini/internal/platform/youtube"
	"github.com/AkifhanIlgaz/tini/internal/shared/htmx"
	"github.com/AkifhanIlgaz/tini/internal/shared/middleware"
	"github.com/gofiber/fiber/v3"
	"go.mongodb.org/mongo-driver/v2/bson"
)

const defaultPageSize = 10

var pageSizeOptions = []int{10, 20, 50}

type PlaylistHandler struct {
	service *PlaylistService
}

func NewHandler(service *PlaylistService) *PlaylistHandler {
	return &PlaylistHandler{service: service}
}

func (h *PlaylistHandler) RegisterRoutes(app *fiber.App) {
	guard := middleware.AuthenticatedLayout()

	app.Get("/playlist", guard, h.List)
	app.Post("/playlist", guard, h.Add)
	app.Post("/playlist/now-playing/next", guard, h.NextTrack)
	app.Post("/playlist/:id/delete", guard, h.Delete)
}

func (h *PlaylistHandler) List(c fiber.Ctx) error {
	u, _ := session.GetCurrentUser(c)

	req := ListItemsRequest{
		VenueID:  u.VenueID,
		Page:     parsePage(c.Query("page")),
		PageSize: parsePageSize(c.Query("pageSize")),
	}
	if err := req.Validate(); err != nil {
		return fmt.Errorf("playlist: list: validate: %w", err)
	}

	result, err := h.service.ListItems(c.Context(), req)
	if err != nil {
		return fmt.Errorf("playlist: list: %w", err)
	}

	// Page may be past the end (ör. the last item on a page just got
	// deleted) — clamp and refetch rather than rendering an empty page with
	// a pager that disagrees with it.
	if result.Page > result.TotalPages && result.TotalPages > 0 {
		req.Page = result.TotalPages
		result, err = h.service.ListItems(c.Context(), req)
		if err != nil {
			return fmt.Errorf("playlist: list: %w", err)
		}
	}

	pageInfo := views.PageInfo{
		Page:            result.Page,
		PageSize:        result.PageSize,
		TotalPages:      max(1, result.TotalPages),
		TotalItems:      int(result.Total),
		PageSizeOptions: pageSizeOptions,
	}

	return htmx.Render(c, views.Playlist(u, c.Path(), csrf.Token(c), toRows(result.Items, result.AddedByNames), pageInfo))
}

func (h *PlaylistHandler) Add(c fiber.Ctx) error {
	u, _ := session.GetCurrentUser(c)

	req := AddLinkRequest{
		VenueID: u.VenueID,
		AddedBy: u.ID,
	}
	if err := c.Bind().Body(&req); err != nil {
		return fmt.Errorf("playlist: add: bind: %w", err)
	}
	req.VenueID = u.VenueID
	req.AddedBy = u.ID

	if err := req.Validate(); err != nil {
		var fieldErrs htmx.FieldErrors
		if errors.As(err, &fieldErrs) {
			return htmx.Render(c, views.AddLinkForm(req.URL, fieldErrs))
		}

		return fmt.Errorf("playlist: add: validate: %w", err)
	}

	if req.IsPlaylist() {
		return h.importPlaylist(c, req)
	}

	return h.addItem(c, req)
}

func (h *PlaylistHandler) addItem(c fiber.Ctx, req AddLinkRequest) error {
	if _, err := h.service.AddItem(c.Context(), req); err != nil {
		if errors.Is(err, ErrItemAlreadyExists) {
			return htmx.Render(c, views.AddLinkForm(req.URL, htmx.FieldErrors{"url": ErrURLAlreadyAdded}))
		}
		if isYoutubeUserError(err) {
			return htmx.Render(c, views.AddLinkForm(req.URL, htmx.FieldErrors{"url": err}))
		}

		return fmt.Errorf("playlist: add: %w", err)
	}

	return htmx.Redirect(c, "/playlist")
}

func (h *PlaylistHandler) importPlaylist(c fiber.Ctx, req AddLinkRequest) error {
	added, err := h.service.ImportPlaylist(c.Context(), req)
	if err != nil {
		if isYoutubeUserError(err) {
			return htmx.Render(c, views.AddLinkForm(req.URL, htmx.FieldErrors{"url": err}))
		}

		return fmt.Errorf("playlist: add: import: %w", err)
	}

	if err := htmx.Toast(c, htmx.ToastOptions{
		Title:       "Playlist içe aktarıldı",
		Description: fmt.Sprintf("%d şarkı eklendi.", added),
		Variant:     htmx.ToastSuccess,
	}); err != nil {
		return fmt.Errorf("playlist: add: toast: %w", err)
	}

	return htmx.Redirect(c, "/playlist")
}

// isYoutubeUserError reports whether err is one of youtube.Client's
// expected, user-facing failures (bad/missing playlist, no API key, ...) —
// these surface as the url field's error instead of the generic 500 path.
func isYoutubeUserError(err error) bool {
	return errors.Is(err, youtube.ErrInvalidURL) ||
		errors.Is(err, youtube.ErrRequestFailed) ||
		errors.Is(err, youtube.ErrAPIKeyMissing) ||
		errors.Is(err, youtube.ErrPlaylistNotFound) ||
		errors.Is(err, youtube.ErrPlaylistEmpty) ||
		errors.Is(err, youtube.ErrUnsupportedPlaylist)
}

func (h *PlaylistHandler) Delete(c fiber.Ctx) error {
	u, _ := session.GetCurrentUser(c)

	id, err := bson.ObjectIDFromHex(c.Params("id"))
	if err != nil {
		return fmt.Errorf("playlist: delete: parse id: %w", err)
	}

	req := DeleteItemRequest{
		VenueID: u.VenueID,
		ID:      id,
	}
	if err := req.Validate(); err != nil {
		return fmt.Errorf("playlist: delete: validate: %w", err)
	}

	if err := h.service.DeleteItem(c.Context(), req); err != nil {
		if errors.Is(err, ErrItemNotFound) {
			return htmx.Redirect(c, "/playlist")
		}

		return fmt.Errorf("playlist: delete: %w", err)
	}

	return htmx.Redirect(c, "/playlist")
}

// NextTrack drives the persistent player bar (internal/shared/layout.Dashboard):
// on the bar's initial load (no currentYoutubeId) and every time
// static/js/player.js advances past a finished/errored video, it POSTs here
// for the next item in playlist order (see PlaylistService.NextTrack) and
// swaps the result into #now-playing-info.
func (h *PlaylistHandler) NextTrack(c fiber.Ctx) error {
	u, _ := session.GetCurrentUser(c)

	var req NextTrackRequest
	if err := c.Bind().Body(&req); err != nil {
		return fmt.Errorf("playlist: next track: bind: %w", err)
	}
	req.VenueID = u.VenueID

	if err := req.Validate(); err != nil {
		return fmt.Errorf("playlist: next track: validate: %w", err)
	}

	item, err := h.service.NextTrack(c.Context(), req)
	if err != nil {
		if errors.Is(err, ErrItemNotFound) {
			return htmx.Render(c, views.NowPlaying(nil))
		}

		return fmt.Errorf("playlist: next track: %w", err)
	}

	return htmx.Render(c, views.NowPlaying(&views.NowPlayingTrack{
		YoutubeID: item.YoutubeID,
		Title:     item.Title,
		Channel:   item.Channel,
		Thumbnail: item.Thumbnail,
	}))
}

func parsePage(raw string) int {
	page, err := strconv.Atoi(raw)
	if err != nil || page < 1 {
		return 1
	}

	return page
}

func parsePageSize(raw string) int {
	pageSize, err := strconv.Atoi(raw)
	if err != nil || !slices.Contains(pageSizeOptions, pageSize) {
		return defaultPageSize
	}

	return pageSize
}

// turkishMonths is time.Month's Turkish name, 1-indexed like time.Month
// itself — Go's time package has no locale support, so formatCreatedAt
// builds the string by hand instead of a Format layout.
var turkishMonths = [...]string{
	"", "Ocak", "Şubat", "Mart", "Nisan", "Mayıs", "Haziran",
	"Temmuz", "Ağustos", "Eylül", "Ekim", "Kasım", "Aralık",
}

func formatCreatedAt(t time.Time) string {
	return fmt.Sprintf("%d %s %d", t.Day(), turkishMonths[t.Month()], t.Year())
}

func toRows(items []PlaylistItem, addedByNames map[bson.ObjectID]string) []views.PlaylistRow {
	rows := make([]views.PlaylistRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, views.PlaylistRow{
			ID:        item.ID.Hex(),
			Title:     item.Title,
			Channel:   item.Channel,
			Thumbnail: item.Thumbnail,
			URL:       "https://youtu.be/" + item.YoutubeID,
			AddedBy:   addedByNames[item.AddedBy],
			CreatedAt: formatCreatedAt(item.CreatedAt),
		})
	}

	return rows
}
