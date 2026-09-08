package playlist

import (
	"context"
	"fmt"

	"github.com/AkifhanIlgaz/tini/internal/features/user"
	"github.com/AkifhanIlgaz/tini/internal/platform/youtube"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type PlaylistService struct {
	repo     PlaylistRepository
	userRepo user.UserRepository
	youtube  *youtube.Client
}

func NewService(repo PlaylistRepository, userRepo user.UserRepository, youtubeClient *youtube.Client) *PlaylistService {
	return &PlaylistService{repo: repo, userRepo: userRepo, youtube: youtubeClient}
}

// PaginatedItems is PlaylistService.ListItems's result — req.Page/PageSize
// echoed back alongside Total/TotalPages so the handler doesn't have to
// recompute them. AddedByNames resolves each item's AddedBy to a display
// name — an id missing from the map means the user no longer exists (see
// user.UserRepository.FindByIDs).
type PaginatedItems struct {
	Items        []PlaylistItem
	AddedByNames map[bson.ObjectID]string
	Page         int
	PageSize     int
	Total        int64
	TotalPages   int
}

// ListItems returns req.VenueID's playlist items, req.Page/PageSize deep.
// req is assumed already validated.
func (s *PlaylistService) ListItems(ctx context.Context, req ListItemsRequest) (PaginatedItems, error) {
	total, err := s.repo.CountByVenueID(ctx, req.VenueID)
	if err != nil {
		return PaginatedItems{}, fmt.Errorf("playlist: list items: %w", err)
	}

	items, err := s.repo.FindByVenueID(ctx, req.VenueID, req.Skip(), int64(req.PageSize))
	if err != nil {
		return PaginatedItems{}, fmt.Errorf("playlist: list items: %w", err)
	}

	addedByNames, err := s.addedByNames(ctx, items)
	if err != nil {
		return PaginatedItems{}, fmt.Errorf("playlist: list items: %w", err)
	}

	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize != 0 {
		totalPages++
	}

	return PaginatedItems{
		Items:        items,
		AddedByNames: addedByNames,
		Page:         req.Page,
		PageSize:     req.PageSize,
		Total:        total,
		TotalPages:   totalPages,
	}, nil
}

// addedByNames batch-resolves each of items' distinct AddedBy into a display
// name (see user.UserRepository.FindByIDs) instead of one lookup per row.
func (s *PlaylistService) addedByNames(ctx context.Context, items []PlaylistItem) (map[bson.ObjectID]string, error) {
	seen := make(map[bson.ObjectID]struct{}, len(items))
	ids := make([]bson.ObjectID, 0, len(items))
	for _, item := range items {
		if _, ok := seen[item.AddedBy]; ok {
			continue
		}
		seen[item.AddedBy] = struct{}{}
		ids = append(ids, item.AddedBy)
	}

	if len(ids) == 0 {
		return map[bson.ObjectID]string{}, nil
	}

	users, err := s.userRepo.FindByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("resolve added by: %w", err)
	}

	names := make(map[bson.ObjectID]string, len(users))
	for _, u := range users {
		names[u.ID] = u.Name
	}

	return names, nil
}

// AddItem fetches req.VideoID's title/channel from YouTube and saves it as a
// new playlist item on req.VenueID. req is assumed already validated and to
// carry a VideoID (single-video add, as opposed to ImportPlaylist).
func (s *PlaylistService) AddItem(ctx context.Context, req AddLinkRequest) (PlaylistItem, error) {
	info, err := s.youtube.ExtractTrackInfo(req.VideoID)
	if err != nil {
		return PlaylistItem{}, fmt.Errorf("playlist: add item: %w", err)
	}

	item, err := s.repo.Insert(ctx, newPlaylistItemFromTrackInfo(req.VenueID, req.AddedBy, *info))
	if err != nil {
		return PlaylistItem{}, fmt.Errorf("playlist: add item: %w", err)
	}

	return item, nil
}

// ImportPlaylist fetches req.PlaylistID's tracks from YouTube and bulk-inserts
// them into req.VenueID's playlist, silently skipping tracks already present
// (see PlaylistRepository.InsertMany). It returns how many tracks were
// actually inserted. req is assumed already validated and to carry a
// PlaylistID (as opposed to AddItem's single-video add).
func (s *PlaylistService) ImportPlaylist(ctx context.Context, req AddLinkRequest) (int, error) {
	tracks, err := s.youtube.FetchPlaylistItems(req.PlaylistID)
	if err != nil {
		return 0, fmt.Errorf("playlist: import playlist: %w", err)
	}

	items := make([]PlaylistItem, len(tracks))
	for i, track := range tracks {
		items[i] = newPlaylistItemFromTrackInfo(req.VenueID, req.AddedBy, track)
	}

	inserted, err := s.repo.InsertMany(ctx, items)
	if err != nil {
		return 0, fmt.Errorf("playlist: import playlist: %w", err)
	}

	return inserted, nil
}

// DeleteItem removes req.ID from req.VenueID's playlist. req is assumed
// already validated.
func (s *PlaylistService) DeleteItem(ctx context.Context, req DeleteItemRequest) error {
	if err := s.repo.Delete(ctx, req.VenueID, req.ID); err != nil {
		return fmt.Errorf("playlist: delete item: %w", err)
	}

	return nil
}

// NextTrack returns req.VenueID's playlist item to play after
// req.CurrentYoutubeID (playlist order, created_at ascending), wrapping to
// the first item past the end — the temporary stand-in for a real queue's
// Next (see PlaylistRepository.NextAfter). req is assumed already validated.
func (s *PlaylistService) NextTrack(ctx context.Context, req NextTrackRequest) (PlaylistItem, error) {
	item, err := s.repo.NextAfter(ctx, req.VenueID, req.CurrentYoutubeID)
	if err != nil {
		return PlaylistItem{}, fmt.Errorf("playlist: next track: %w", err)
	}

	return item, nil
}

// newPlaylistItemFromTrackInfo maps a YouTube track lookup onto a
// PlaylistItem for venueID, added by addedBy.
func newPlaylistItemFromTrackInfo(venueID, addedBy bson.ObjectID, info youtube.TrackInfo) PlaylistItem {
	return PlaylistItem{
		VenueID:   venueID,
		YoutubeID: info.ID,
		Title:     info.Title,
		Channel:   info.Channel,
		Thumbnail: info.Thumbnail,
		AddedBy:   addedBy,
	}
}
