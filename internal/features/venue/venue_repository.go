package venue

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

type VenueRepository interface {
	FindByID(ctx context.Context, id bson.ObjectID) (Venue, error)
	Update(ctx context.Context, id bson.ObjectID, name string, settings VenueSettings) (Venue, error)
}

type venueMongoRepository struct {
	collection *mongo.Collection
}

func NewRepository(client *db.Client) (VenueRepository, error) {
	collection := client.Database().Collection(VenueCollectionName)

	return &venueMongoRepository{
		collection: collection,
	}, nil
}

func (r *venueMongoRepository) FindByID(ctx context.Context, id bson.ObjectID) (Venue, error) {
	var v Venue

	filter := bson.M{
		"_id": id,
	}

	err := r.collection.FindOne(ctx, filter).Decode(&v)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Venue{}, ErrVenueNotFound
		}

		return Venue{}, fmt.Errorf("venue: find by id: %w", err)
	}

	return v, nil
}

func (r *venueMongoRepository) Update(ctx context.Context, id bson.ObjectID, name string, settings VenueSettings) (Venue, error) {
	var v Venue

	filter := bson.M{
		"_id": id,
	}

	update := bson.M{
		"$set": bson.M{
			"name":       name,
			"settings":   settings,
			"updated_at": time.Now(),
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&v)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Venue{}, ErrVenueNotFound
		}

		return Venue{}, fmt.Errorf("venue: update: %w", err)
	}

	return v, nil
}
