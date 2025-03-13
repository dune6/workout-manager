package trainings

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

const (
	PullUp = "Pull up"
	PushUp = "Push up"
)

type Exercise struct {
	Type            string        `json:"type" bson:"type"`
	Count           int           `json:"count" bson:"count"`
	Weight          float32       `json:"weight" bson:"weight"`
	DurationWorkout time.Duration `json:"duration_workout,omitempty" bson:"duration_workout,omitempty"`
	DurationRest    time.Duration `json:"duration_rest,omitempty" bson:"duration_rest,omitempty"`
}

type Training struct {
	ID               primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Username         string             `json:"username" bson:"username"`
	Date             primitive.DateTime `json:"date" bson:"date"`
	Tonnage          float32            `json:"tonnage" bson:"tonnage"`
	Number           int                `json:"number" bson:"number"`
	TotalWorkoutTime time.Duration      `json:"total_workout_time,omitempty" bson:"total_workout_time,omitempty"`
	TotalRestTime    time.Duration      `json:"total_rest_time,omitempty" bson:"total_rest_time,omitempty"`
	Exercises        []Exercise         `json:"exercises" bson:"exercises"`
	Feedback         string             `json:"feedback,omitempty" bson:"feedback,omitempty"`
	Like             bool               `json:"like,omitempty" bson:"like,omitempty"`
}
