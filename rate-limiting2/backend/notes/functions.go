package backend

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Init() (err error) {
	Database, err := ConnectMongo()
	if err != nil {
		err = ErrAccessDenied
		log.Error(err)
		return
	}
	rdb := NewRedisClient()
	limiter = NewRateLimiter(rdb, 5, time.Minute)
	Notes = Database.Collection("notes")
	return
}

func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
}

func NewRateLimiter(rdb *redis.Client, limit int64, window time.Duration) (r1 *RateLimiter) {
	return &RateLimiter{
		rdb:        rdb,
		limit:      limit,
		windowTime: window,
	}
}

func (r1 *RateLimiter) Allow(ip string) bool {
	if ip == "" {
		return false
	}
	key := fmt.Sprintf("rate_limit:ip:%s", ip)
	pipe := r1.rdb.TxPipeline()
	count := pipe.Incr(context.TODO(), key)
	//set expiry time only when the key is created newly
	pipe.ExpireNX(context.TODO(), key, r1.windowTime)
	//created pipeline so that all increment, expiry will be executed together through exec
	_, err := pipe.Exec(context.TODO())
	if err != nil {
		return false
	}
	if count.Val() > r1.limit {
		return false
	}
	return true
}

func checkRateLimit(c *fiber.Ctx) bool {
	ip := c.IP()
	if !limiter.Allow(ip) {
		c.Status(429).JSON(fiber.Map{
			"error": "too many requests",
		})
		return false
	}
	return true
}

func CreateNotes(c *fiber.Ctx) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if !checkRateLimit(c) {
		return nil
	}
	notesData := Note{}
	err = c.BodyParser(&notesData)
	if err != nil {
		err = fmt.Errorf("error parsing the content", err)
		log.Error(err)
		return err
	}
	notesData.ID = primitive.NewObjectID()
	notesData.CreatedAt = time.Now()
	content := notesData.Content
	if content == "" {
		err = fmt.Errorf("content is empty", err)
		log.Error(err)
		return
	}
	notesInserted, err := Notes.InsertOne(ctx, notesData)
	if err != nil {
		err = fmt.Errorf("error parsing the content", err)
		log.Error(err)
		return
	}
	return c.Status(fiber.StatusCreated).JSON(notesInserted)
}

func UpdateNotes(c *fiber.Ctx) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if !checkRateLimit(c) {
		return nil
	}
	notesData := Note{}
	err = c.BodyParser(&notesData)
	if err != nil {
		err = fmt.Errorf("error parsing the content", err)
		log.Error(err)
		return err
	}
	content := notesData.Content
	filter := bson.D{
		{
			Key:   "_id",
			Value: notesData.ID,
		},
	}
	set := bson.D{
		{
			Key:   "content",
			Value: content,
		},
	}
	update := bson.D{
		{
			Key:   "$set",
			Value: set,
		},
	}
	_, err = Notes.UpdateOne(ctx, filter, update)
	if err != nil {
		err = fmt.Errorf("error updating the notes", err)
		log.Error(err)
		return
	}
	return
}

func DeleteNotes(c *fiber.Ctx) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if !checkRateLimit(c) {
		return nil
	}
	var body map[string]string
	if err := c.BodyParser(&body); err != nil {
		log.Error(err)
		return fmt.Errorf("error fetching the id")
	}
	ID := body["_id"]
	id, err := primitive.ObjectIDFromHex(ID)
	filter := bson.D{
		{
			Key:   "_id",
			Value: id,
		},
	}
	_, err = Notes.DeleteOne(ctx, filter)
	if err != nil {
		err = fmt.Errorf("error deleting the content", err)
		log.Error(err)
		return
	}
	return
}

func ConnectMongo() (DB *mongo.Database, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(
		"mongodb://localhost:27017",
	))
	if err != nil {
		log.Fatal(err)
	}
	DB = client.Database("notes")
	fmt.Println("✅ MongoDB connected")
	return
}

func GetNotes(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := bson.D{}
	cursor, err := Notes.Find(ctx, filter)
	if err != nil {
		log.Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to fetch notes",
		})
	}
	defer cursor.Close(ctx)
	var notes []Note
	if err := cursor.All(ctx, &notes); err != nil {
		log.Error(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to decode notes",
		})
	}
	return c.Status(fiber.StatusOK).JSON(notes)
}
