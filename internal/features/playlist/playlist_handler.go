// Package playlist is the playlist feature: an add-link form plus a
// paginated table over an in-memory mock list — no real backend
// (repository/service) yet.
package playlist

import (
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/AkifhanIlgaz/tini/internal/features/playlist/views"
	"github.com/AkifhanIlgaz/tini/internal/platform/csrf"
	"github.com/AkifhanIlgaz/tini/internal/platform/session"
	"github.com/AkifhanIlgaz/tini/internal/shared/htmx"
	"github.com/AkifhanIlgaz/tini/internal/shared/middleware"
	"github.com/gofiber/fiber/v3"
)

const defaultPageSize = 10

var pageSizeOptions = []int{10, 20, 50}

type PlaylistHandler struct{}

func NewHandler() *PlaylistHandler {
	return &PlaylistHandler{}
}

func (h *PlaylistHandler) RegisterRoutes(app *fiber.App) {
	guard := middleware.AuthenticatedLayout()

	app.Get("/playlist", guard, h.List)
	app.Post("/playlist", guard, h.Add)
	app.Post("/playlist/:id/delete", guard, h.Delete)
}

// List paginates mockItems in memory — playlist.PlaylistRepository doesn't exist
// yet, so there's nothing real to query until then.
func (h *PlaylistHandler) List(c fiber.Ctx) error {
	u, _ := session.GetCurrentUser(c)

	items := mockItems()
	pageSize := parsePageSize(c.Query("pageSize"))

	totalItems := len(items)
	totalPages := max(1, (totalItems+pageSize-1)/pageSize)
	page := min(parsePage(c.Query("page")), totalPages)

	start := (page - 1) * pageSize
	end := min(start+pageSize, totalItems)

	pageInfo := views.PageInfo{
		Page:            page,
		PageSize:        pageSize,
		TotalPages:      totalPages,
		TotalItems:      totalItems,
		PageSizeOptions: pageSizeOptions,
	}

	return htmx.Render(c, views.Playlist(u, c.Path(), csrf.Token(c), toRows(items[start:end]), pageInfo))
}

// Add only demonstrates the field-error/redirect round-trip for now —
// nothing is actually persisted until playlist.PlaylistRepository exists.
func (h *PlaylistHandler) Add(c fiber.Ctx) error {
	var req AddLinkRequest
	if err := c.Bind().Body(&req); err != nil {
		return fmt.Errorf("playlist: add: bind: %w", err)
	}

	if err := req.Validate(); err != nil {
		var fieldErrs htmx.FieldErrors
		if errors.As(err, &fieldErrs) {
			return htmx.Render(c, views.AddLinkForm(req.URL, fieldErrs))
		}

		return fmt.Errorf("playlist: add: validate: %w", err)
	}

	return htmx.Redirect(c, "/playlist")
}

// Delete only demonstrates the redirect round-trip for now — nothing is
// actually removed until playlist.PlaylistRepository exists.
func (h *PlaylistHandler) Delete(c fiber.Ctx) error {
	if c.Params("id") == "" {
		return errors.New("playlist: delete: missing id")
	}

	return htmx.Redirect(c, "/playlist")
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

// mockItem stands in for a playlist item until playlist.PlaylistRepository exists.
type mockItem struct {
	id    string
	title string
	url   string
}

func mockItems() []mockItem {
	items := make([]mockItem, 0, 23)
	for i := 1; i <= 23; i++ {
		items = append(items, mockItem{
			id:    strconv.Itoa(i),
			title: fmt.Sprintf("Şarkı %d", i),
			url:   fmt.Sprintf("https://youtu.be/mock%04d", i),
		})
	}

	return items
}

func toRows(items []mockItem) []views.PlaylistRow {
	rows := make([]views.PlaylistRow, 0, len(items))
	for _, it := range items {
		rows = append(rows, views.PlaylistRow{ID: it.id, Title: it.title, URL: it.url})
	}

	return rows
}
