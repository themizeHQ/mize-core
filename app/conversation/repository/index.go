package repository

import (
	"mize.app/app/conversation/models"
	dbMongo "mize.app/db/mongo"
	mongoRepo "mize.app/repository/database/mongo"
)

var MessageRepository mongoRepo.MongoRepository[models.Message]
var ConversationRepository mongoRepo.MongoRepository[models.Conversation]
var ReactionRepository mongoRepo.MongoRepository[models.Reaction]
var ConversationMemberRepository mongoRepo.MongoRepository[models.ConversationMember]

func GetMessageRepo() mongoRepo.MongoRepository[models.Message] {
	MessageRepository = mongoRepo.MongoRepository[models.Message]{Model: dbMongo.Message}
	return MessageRepository
}

func GetConversationRepo() mongoRepo.MongoRepository[models.Conversation] {
	ConversationRepository = mongoRepo.MongoRepository[models.Conversation]{Model: dbMongo.Conversation}
	return ConversationRepository
}

func GetConversationMemberRepo() mongoRepo.MongoRepository[models.ConversationMember] {
	ConversationMemberRepository = mongoRepo.MongoRepository[models.ConversationMember]{Model: dbMongo.ConversationMember}
	return ConversationMemberRepository
}

func GetReactionRepo() mongoRepo.MongoRepository[models.Reaction] {
	ReactionRepository = mongoRepo.MongoRepository[models.Reaction]{Model: dbMongo.Reaction}
	return ReactionRepository
}
