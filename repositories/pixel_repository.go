package repositories

import (
	"context"

	"your_project/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PixelRepository interface {
	GetAllPixels(ctx context.Context) ([]models.Pixel, error)
	UpsertPixel(ctx context.Context, pixel models.Pixel) error
}

type pixelRepository struct {
	collection *mongo.Collection
}

func NewPixelRepository(db *mongo.Database) PixelRepository {
	return &pixelRepository{
		collection: db.Collection("pixels"),
	}
}

func (pr *pixelRepository) GetAllPixels(ctx context.Context) ([]models.Pixel, error) {
	cursor, err := pr.collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var pixels []models.Pixel
	if err := cursor.All(ctx, &pixels); err != nil {
		return nil, err
	}
	return pixels, nil
}

func (pr *pixelRepository) UpsertPixel(ctx context.Context, pixel models.Pixel) error {
	filter := bson.M{"x": pixel.X, "y": pixel.Y}
	update := bson.M{"$set": pixel}
	options := options.Update().SetUpsert(true)
	_, err := pr.collection.UpdateOne(ctx, filter, update, options)
	return err
}
