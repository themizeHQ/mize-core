package mongo

import (
	"context"
	"fmt"
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
	ChannelMember   *mongo.Collection

	// notifications
	Alert        *mongo.Collection
	Notification *mongo.Collection

	// messages
	Message            *mongo.Collection
	Conversation       *mongo.Collection
	ConversationMember *mongo.Collection

	// media
	Upload *mongo.Collection

	// team
	Team         *mongo.Collection
	TeamMember   *mongo.Collection
	TeamActivity *mongo.Collection

	// schedule
	Schedule *mongo.Collection

	// call log
	CallLog *mongo.Collection
)

func ConnectMongo() context.CancelFunc {
	uri := os.Getenv("DB_URL")

	if uri == "" {
		fmt.Println("set mongo url")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))

	if err != nil {
		if os.Getenv("GIN_MODE") == "debug" {
			fmt.Println(err)
		} else {
			fmt.Println("something went wrong")
		}
		return cancel
	}

	db := client.Database(os.Getenv("DB_NAME"))
	setUpIndexes(ctx, db)

	return cancel
}

func setUpIndexes(ctx context.Context, db *mongo.Database) {
	UserModel = db.Collection("Users")
	UserModel.Indexes().CreateMany(ctx, []mongo.IndexModel{{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}, {
		Keys:    bson.D{{Key: "phone", Value: 1}},
		Options: options.Index().SetUnique(true).SetSparse(true),
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
	WorkspaceModel.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "createdBy", Value: 1}},
		Options: options.Index(),
	})

	WorkspaceInvite = db.Collection("WorkspaceInvites")
	WorkspaceInvite.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "workspaceId", Value: 1}},
		Options: options.Index(),
	})

	Channel = db.Collection("Channels")
	Channel.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "workspaceId", Value: 1}},
		Options: options.Index(),
	})

	WorkspaceMember = db.Collection("WorkspaceMembers")
	WorkspaceMember.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "userName", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "userId", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "admin", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "workspaceId", Value: 1}},
		},
	})

	ChannelMember = db.Collection("ChannelMembers")
	ChannelMember.Indexes().CreateMany(ctx, []mongo.IndexModel{
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
			Keys: bson.D{{Key: "workspaceId", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "pinned", Value: 1}},
		},
	})

	Notification = db.Collection("Notifications")
	Notification.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "userId", Value: 1}},
		},
	})

	Alert = db.Collection("Alerts")
	Alert.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "workspaceId", Value: 1}},
		},
	})

	Message = db.Collection("Messages")
	Message.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "to", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "from", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "userName", Value: 1}},
		},
	})

	Upload = db.Collection("Uploads")
	Upload.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "uploadBy", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "format", Value: 1}},
		},
	})

	Conversation = db.Collection("Conversation")
	Conversation.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "participants", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "format", Value: 1}},
		},
	})

	ConversationMember = db.Collection("ConversationMembers")
	ConversationMember.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "workspaceConv", Value: 1}},
		},
	})

	Team = db.Collection("Teams")
	Team.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "workspaceId", Value: 1}},
		},
	})
	TeamMember = db.Collection("TeamMembers")
	TeamMember.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "workspaceId", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "teamId", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "userId", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "workspaceMemberId", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "userName", Value: 1}},
		},
	})

	TeamActivity = db.Collection("TeamActivity")
	TeamActivity.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "workspaceId", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "teamId", Value: 1}},
		},
	})

	Schedule = db.Collection("Schedules")
	Schedule.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "workspaceId", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "createdBy", Value: 1}},
		},
	})

	CallLog = db.Collection("CallLogs")
	CallLog.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "userId", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "userId", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "userId", Value: 1}},
		},
	})
}
