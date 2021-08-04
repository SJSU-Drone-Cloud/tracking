package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Tracking struct {
	ID           primitive.ObjectID  `json:"_id" bson:"_id"`
	DroneID      string              `json:"droneID" bson:"droneID"`
	TimeLocation []TimeLocationStamp `json:"timeLocation" bson:"timeLocation"`
	LastUpdated  time.Time           `json:"lastUpdated" bson:"lastUpdated"`
}

type TrackingDevice struct {
	DroneID string  `json:"droneID" bson:"droneID"`
	Lat     float64 `json:"lat,string" bson:"lat"`
	Lng     float64 `json:"lng,string" bson:"lng"`
}

type TimeLocationStamp struct {
	TimeStamp time.Time `json:"timestamp" bson:"timestamp"`
	Lat       float64   `json:"lat" bson:"lat"`
	Lng       float64   `json:"lng" bson:"lng"`
}
