package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// MongoDB holds the client and database references for MongoDB operations.
type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// ConnectMongoDB initializes a connection to MongoDB with connection timeouts.
func ConnectMongoDB(ctx context.Context, uri, dbName string) (*MongoDB, error) {
	clientOptions := options.Client().
		ApplyURI(uri).
		SetServerSelectionTimeout(3 * time.Second).
		SetConnectTimeout(3 * time.Second)

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create MongoDB client: %w", err)
	}

	// Verify connectivity with a short timeout
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx, readpref.Primary()); err != nil {
		log.Printf("[WARN] MongoDB ping failed at startup: %v", err)
		// Return the initialized client so health checks can report status and attempt reconnects
		return &MongoDB{
			Client:   client,
			Database: client.Database(dbName),
		}, err
	}

	log.Println("[INFO] Successfully connected to MongoDB")
	return &MongoDB{
		Client:   client,
		Database: client.Database(dbName),
	}, nil
}

// Ping checks if the MongoDB server is reachable.
func (m *MongoDB) Ping(ctx context.Context) error {
	if m == nil || m.Client == nil {
		return fmt.Errorf("mongo client is not initialized")
	}
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return m.Client.Ping(pingCtx, readpref.Primary())
}

// Disconnect gracefully shuts down the MongoDB client connection.
func (m *MongoDB) Disconnect(ctx context.Context) error {
	if m == nil || m.Client == nil {
		return nil
	}
	log.Println("[INFO] Disconnecting from MongoDB...")
	return m.Client.Disconnect(ctx)
}
