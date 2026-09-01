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

type Repository interface {
	FindByID(ctx context.Context, id bson.ObjectID) (Venue, error)
	Update(ctx context.Context, id bson.ObjectID, name string, settings VenueSettings) (Venue, error)
}

type mongoRepository struct {
	collection *mongo.Collection
}

func NewRepository(client *db.Client) (Repository, error) {
	collection := client.Database().Collection(CollectionName)

	return &mongoRepository{
		collection: collection,
	}, nil
}

func (r *mongoRepository) FindByID(ctx context.Context, id bson.ObjectID) (Venue, error) {
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

func (r *mongoRepository) Update(ctx context.Context, id bson.ObjectID, name string, settings VenueSettings) (Venue, error) {
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
