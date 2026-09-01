package playlist

import (
	"context"
	"errors"
	"fmt"

	db "github.com/AkifhanIlgaz/tini/internal/platform/mongo"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type PlaylistRepository interface {
	FindByVenueID(ctx context.Context, venueID bson.ObjectID) ([]PlaylistItem, error)
	Insert(ctx context.Context, item PlaylistItem) (PlaylistItem, error)
	Delete(ctx context.Context, venueID, id bson.ObjectID) error
}

type playlistMongoRepository struct {
	collection *mongo.Collection
}

func NewRepository(client *db.Client) (PlaylistRepository, error) {
	collection := client.Database().Collection(PlaylistItemsCollectionName)

	return &playlistMongoRepository{
		collection: collection,
	}, nil
}

func (r *playlistMongoRepository) FindByVenueID(ctx context.Context, venueID bson.ObjectID) ([]PlaylistItem, error) {
	return nil, errors.New("playlist: find by venue id: not implemented")
}

func (r *playlistMongoRepository) Insert(ctx context.Context, item PlaylistItem) (PlaylistItem, error) {
	return PlaylistItem{}, errors.New("playlist: insert: not implemented")
}

func (r *playlistMongoRepository) Delete(ctx context.Context, venueID, id bson.ObjectID) error {
	return fmt.Errorf("playlist: delete: not implemented")
}
