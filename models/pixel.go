package models

type Pixel struct {
	X     int    `json:"x" bson:"x"`
	Y     int    `json:"y" bson:"y"`
	Color string `json:"color" bson:"color"`
}
