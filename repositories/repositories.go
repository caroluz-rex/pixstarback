// repositories/team_repository.go
package repositories

import (
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"your_project/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type TeamRepository interface {
	CreateTeam(ctx context.Context, team *models.Team) error
	GetTeams(ctx context.Context) ([]models.Team, error)
	GetTeamByID(ctx context.Context, id string) (*models.Team, error)
	AddMember(ctx context.Context, teamID string, member string) error
	RemoveMember(ctx context.Context, teamID string, member string) error
	GetTeamsByMember(ctx context.Context, member string) ([]models.Team, error)
}

type teamRepository struct {
	collection *mongo.Collection
}

func NewTeamRepository(db *mongo.Database) TeamRepository {
	return &teamRepository{
		collection: db.Collection("teams"),
	}
}

func (tr *teamRepository) CreateTeam(ctx context.Context, team *models.Team) error {
	result, err := tr.collection.InsertOne(ctx, team)
	if err != nil {
		return err
	}
	team.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (tr *teamRepository) GetTeams(ctx context.Context) ([]models.Team, error) {
	cursor, err := tr.collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var teams []models.Team
	if err := cursor.All(ctx, &teams); err != nil {
		return nil, err
	}
	return teams, nil
}

func (tr *teamRepository) GetTeamByID(ctx context.Context, id string) (*models.Team, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var team models.Team
	if err := tr.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&team); err != nil {
		return nil, err
	}
	return &team, nil
}

func (tr *teamRepository) AddMember(ctx context.Context, teamID string, member string) error {
	objID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": objID}
	update := bson.M{"$addToSet": bson.M{"members": member}}
	_, err = tr.collection.UpdateOne(ctx, filter, update)
	return err
}

func (tr *teamRepository) RemoveMember(ctx context.Context, teamID string, member string) error {
	objID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": objID}
	update := bson.M{"$pull": bson.M{"members": member}}
	_, err = tr.collection.UpdateOne(ctx, filter, update)
	return err
}

func (tr *teamRepository) GetTeamsByMember(ctx context.Context, member string) ([]models.Team, error) {
	filter := bson.M{"members": member}
	cursor, err := tr.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var teams []models.Team
	if err := cursor.All(ctx, &teams); err != nil {
		return nil, err
	}
	return teams, nil
}
