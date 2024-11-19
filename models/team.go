// models/team.go
package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Team struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name    string             `bson:"name" json:"name"`
	Members []string           `bson:"members" json:"members"`
}
