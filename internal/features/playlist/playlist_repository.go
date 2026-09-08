package playlist

import (
	"context"
	"errors"
	"fmt"
	"time"

	db "github.com/AkifhanIlgaz/tini/internal/platform/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type PlaylistRepository interface {
	FindByVenueID(ctx context.Context, venueID bson.ObjectID, skip, limit int64) ([]PlaylistItem, error)
	CountByVenueID(ctx context.Context, venueID bson.ObjectID) (int64, error)
	Insert(ctx context.Context, item PlaylistItem) (PlaylistItem, error)
	InsertMany(ctx context.Context, items []PlaylistItem) (int, error)
	Delete(ctx context.Context, venueID, id bson.ObjectID) error
	NextAfter(ctx context.Context, venueID bson.ObjectID, currentYoutubeID string) (PlaylistItem, error)
}

type playlistMongoRepository struct {
	collection *mongo.Collection
}

func NewRepository(client *db.Client) (PlaylistRepository, error) {
	collection := client.Database().Collection(PlaylistItemsCollectionName)

	_, err := collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{
			{Key: "youtube_id", Value: 1},
			{Key: "venue_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return &playlistMongoRepository{}, fmt.Errorf("playlist: create indexes: %w", err)
	}

	return &playlistMongoRepository{
		collection: collection,
	}, nil
}

func (r *playlistMongoRepository) FindByVenueID(ctx context.Context, venueID bson.ObjectID, skip, limit int64) ([]PlaylistItem, error) {
	filter := bson.M{
		"venue_id": venueID,
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetSkip(skip).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("playlist: find by venue id: %w", err)
	}

	var items []PlaylistItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, fmt.Errorf("playlist: find by venue id: %w", err)
	}

	return items, nil
}

func (r *playlistMongoRepository) CountByVenueID(ctx context.Context, venueID bson.ObjectID) (int64, error) {
	filter := bson.M{
		"venue_id": venueID,
	}

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("playlist: count by venue id: %w", err)
	}

	return count, nil
}

// Insert saves item as a new playlist entry — item.YoutubeID must be
// unique within item.VenueID (see the unique index NewRepository sets up),
// so a re-added link surfaces as ErrItemAlreadyExists rather than a raw
// duplicate-key error.
func (r *playlistMongoRepository) Insert(ctx context.Context, item PlaylistItem) (PlaylistItem, error) {
	item.CreatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, item)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return PlaylistItem{}, ErrItemAlreadyExists
		}

		return PlaylistItem{}, fmt.Errorf("playlist: insert: %w", err)
	}

	item.ID = result.InsertedID.(bson.ObjectID)

	return item, nil
}

// InsertMany bulk-inserts items (a YouTube playlist import) unordered, so
// one item colliding with the unique (youtube_id, venue_id) index doesn't
// block the rest — it returns how many were actually inserted, silently
// skipping duplicates.
func (r *playlistMongoRepository) InsertMany(ctx context.Context, items []PlaylistItem) (int, error) {
	now := time.Now()

	docs := make([]any, len(items))
	for i := range items {
		items[i].CreatedAt = now
		docs[i] = items[i]
	}

	result, err := r.collection.InsertMany(ctx, docs, options.InsertMany().SetOrdered(false))
	if err != nil {
		var bulkErr mongo.BulkWriteException
		if !errors.As(err, &bulkErr) {
			return 0, fmt.Errorf("playlist: insert many: %w", err)
		}
	}

	return len(result.InsertedIDs), nil
}

func (r *playlistMongoRepository) Delete(ctx context.Context, venueID, id bson.ObjectID) error {
	filter := bson.M{
		"_id":      id,
		"venue_id": venueID,
	}

	err := r.collection.FindOneAndDelete(ctx, filter).Err()
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrItemNotFound
		}

		return fmt.Errorf("playlist: delete: %w", err)
	}

	return nil
}

// NextAfter returns venueID's playlist item to play after currentYoutubeID,
// in playlist order (created_at ascending) — the temporary stand-in for a
// real queue's Next until that feature exists. An empty currentYoutubeID, or
// one no longer in the playlist (deleted since), starts from the beginning;
// running past the last item wraps back to the first. ErrItemNotFound is
// returned only when the playlist has no items at all.
func (r *playlistMongoRepository) NextAfter(ctx context.Context, venueID bson.ObjectID, currentYoutubeID string) (PlaylistItem, error) {
	sortByCreatedAtAsc := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: 1}})

	if currentYoutubeID != "" {
		var current PlaylistItem
		err := r.collection.FindOne(ctx, bson.M{"venue_id": venueID, "youtube_id": currentYoutubeID}).Decode(&current)
		if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
			return PlaylistItem{}, fmt.Errorf("playlist: next after: find current: %w", err)
		}

		if err == nil {
			var next PlaylistItem
			err := r.collection.FindOne(ctx, bson.M{
				"venue_id":   venueID,
				"created_at": bson.M{"$gt": current.CreatedAt},
			}, sortByCreatedAtAsc).Decode(&next)
			if err == nil {
				return next, nil
			}
			if !errors.Is(err, mongo.ErrNoDocuments) {
				return PlaylistItem{}, fmt.Errorf("playlist: next after: find next: %w", err)
			}
			// Fall through to wrap back to the first item below.
		}
	}

	var first PlaylistItem
	err := r.collection.FindOne(ctx, bson.M{"venue_id": venueID}, sortByCreatedAtAsc).Decode(&first)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return PlaylistItem{}, ErrItemNotFound
		}

		return PlaylistItem{}, fmt.Errorf("playlist: next after: find first: %w", err)
	}

	return first, nil
}
