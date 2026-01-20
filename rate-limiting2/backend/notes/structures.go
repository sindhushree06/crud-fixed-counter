package backend

import (
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	Notes   *mongo.Collection
	DB      *mongo.Database
	limiter *RateLimiter
)

type RateLimiter struct {
	rdb        *redis.Client
	limit      int64
	windowTime time.Duration
}

type Visitor struct {
	tokens   int
	lastSeen time.Time
}

type Note struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id"`
	Content   string             `json:"content" bson:"content"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}
