package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client

func InitMongoDB() {
	connectionString := "mongodb://user:pass@localhost:27021/firecrackerdb?authSource=admin&authMechanism=SCRAM-SHA-256"

	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for i := 0; i < 5; i++ {
		mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
		if err != nil {
			log.Println("Failed to create client:", err)
			time.Sleep(2 * time.Second)
			continue
		}
		err = mongoClient.Ping(ctx, nil)
		if err != nil {
			log.Println("Failed to ping:", err)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}

	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}

	log.Println("Connected to MongoDB!")
}

func SaveBridgeDetails(bridgeName string, tapName string, userID string, ipAddress string) error {
	if mongoClient == nil {
		log.Fatal("MongoDB client not initialized.")
		return nil
	}

	collection := mongoClient.Database("firecrackerdb").Collection("bridge-info")
	document := bson.D{
		{Key: "userID", Value: userID},
		{Key: "bridgeName", Value: bridgeName},
		{Key: "tapName", Value: tapName},
		{Key: "ipAddress", Value: ipAddress},
		{Key: "created_at", Value: time.Now()},
	}

	_, err := collection.InsertOne(context.Background(), document)
	if err != nil {
		log.Println("Error saving network details to MongoDB:", err)
		return err
	}

	log.Println("Network details saved to MongoDB.")

	return nil
}

func GetBridgeDetails() ([]bson.M, error) {
	if mongoClient == nil {
		log.Fatal("MongoDB client not initialized.")
		return nil, nil
	}

	collection := mongoClient.Database("firecrackerdb").Collection("bridge-info")

	cursor, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		log.Println("Error fetching bridge details from MongoDB:", err)
		return nil, err
	}

	var results []bson.M
	if err = cursor.All(context.Background(), &results); err != nil {
		log.Println("Error fetching bridge details from MongoDB:", err)
		return nil, err
	}

	return results, nil
}
