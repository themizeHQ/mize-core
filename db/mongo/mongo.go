package db

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	// app
	AppModel *mongo.Collection

	// user
	UserModel *mongo.Collection

	// workspace
	WorkspaceModel  *mongo.Collection
	WorkspaceInvite *mongo.Collection
	Channel         *mongo.Collection
	WorkspaceMember *mongo.Collection
)

func ConnectMongo() context.CancelFunc {
	uri := os.Getenv("DB_URL")

	if uri == "" {
		log.Fatalln("Please set a valid MongoDB URL")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))

	if err != nil {
		panic(err)
	}

	db := client.Database("mize-core-prod")
	setUpIndexes(ctx, db)

	return cancel
}

func setUpIndexes(ctx context.Context, db *mongo.Database) {
	UserModel = db.Collection("Users")
	UserModel.Indexes().CreateMany(ctx, []mongo.IndexModel{{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}, {
		Keys:    bson.D{{Key: "userName", Value: 1}},
		Options: options.Index().SetUnique(true),
	}})

	AppModel = db.Collection("Apps")
	AppModel.Indexes().CreateMany(ctx, []mongo.IndexModel{{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	}})

	WorkspaceModel = db.Collection("Workspaces")

	WorkspaceInvite = db.Collection("WorkspaceInvites")

	Channel = db.Collection("Channels")

	WorkspaceMember = db.Collection("WorkspaceMember")
	WorkspaceMember.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "username", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "userId", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "admin", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "banned", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "workspaceId", Value: 1}},
		},
	})
}
