package data

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

type Models struct {
	LogEntry LogEntry
}

type LogEntry struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string    `bson:"name,omitempty" json:"name,omitempty"`
	Data      string    `bson:"data,omitempty" json:"data,omitempty"`
	CreatedAt time.Time `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt time.Time `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}

func New(mongo *mongo.Client) Models {
	client = mongo
	return Models{
		LogEntry: LogEntry{},
	}
}

func (this *LogEntry) Insert(entry LogEntry) error {
	collection := client.Database("logs").Collection("logs")
	_, err := collection.InsertOne(context.TODO(), LogEntry{
		Name:      entry.Name,
		Data:      entry.Data,
		CreatedAt: entry.CreatedAt,
		UpdatedAt: entry.UpdatedAt,
	})
	if err != nil {
		log.Println("error inserting log", err)
		return err
	}
	return nil
}

func (this *LogEntry) All() (*[]LogEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("logs")
	opts := options.Find()
	opts.SetSort(bson.D{{"created_at", -1}})

	cursor, err := collection.Find(context.TODO(), bson.D{}, opts)
	if err != nil {
		log.Println("Error getting all logs", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []LogEntry
	for cursor.Next(ctx) {
		var item LogEntry

		err := cursor.Decode(&item)
		if err != nil {
			log.Println("Error decoding log", err)
			return nil, err
		}
		logs = append(logs, item)
	}
	return &logs, nil
}

func (this *LogEntry) GetOne(id string) (*LogEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("logs")
	docId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("Error in getting ID", err)
		return nil, err
	}

	var entry *LogEntry
	err = collection.FindOne(ctx, bson.M{"_id": docId}).Decode(entry)
	if err != nil {
		log.Println("Error in de referencing", err)
		return nil, err
	}
	return entry, nil
}

func (this *LogEntry) DropCollection(entry LogEntry) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("logs")
	if err := collection.Drop(ctx); err != nil {
		log.Println("Error dropping collection", err)
		return err
	}
	return nil
}

func (this *LogEntry) Update() (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("logs")
	docID, err := primitive.ObjectIDFromHex(this.ID)
	if err != nil {
		log.Println("Error in getting ID", err)
		return nil, err
	}
	res, err := collection.UpdateOne(ctx, bson.M{"_id": docID},
		bson.D{{
			"$set", bson.D{
				{"name", this.Name},
				{"data", this.Data},
				{"update_at", this.UpdatedAt},
			},
		},
		})
	if err != nil {
		log.Println("Error in updating", err)
		return nil, err
	}
	return res, nil
}
