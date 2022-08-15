package repository

import (
	"mize.app/app/workspace/models"
	db "mize.app/db/mongo"
	repository "mize.app/repository/database/mongo"
)

var WorkspaceRepository repository.MongoRepository[models.Workspace]
var WorkspaceInviteRepository repository.MongoRepository[models.WorkspaceInvite]
var ChannelRepository repository.MongoRepository[models.Channel]
var ChannelMemberRepository repository.MongoRepository[models.ChannelMember]
var WorkspaceMember repository.MongoRepository[models.WorkspaceMember]

func GetWorkspaceRepo() repository.MongoRepository[models.Workspace] {
	WorkspaceRepository = repository.MongoRepository[models.Workspace]{Model: db.WorkspaceModel}
	return WorkspaceRepository
}

func GetWorkspaceInviteRepo() repository.MongoRepository[models.WorkspaceInvite] {
	WorkspaceInviteRepository = repository.MongoRepository[models.WorkspaceInvite]{Model: db.WorkspaceInvite}
	return WorkspaceInviteRepository
}

func GetChannelRepo() repository.MongoRepository[models.Channel] {
	ChannelRepository = repository.MongoRepository[models.Channel]{Model: db.Channel}
	return ChannelRepository
}

func GetWorkspaceMember() repository.MongoRepository[models.WorkspaceMember] {
	WorkspaceMember = repository.MongoRepository[models.WorkspaceMember]{Model: db.WorkspaceMember}
	return WorkspaceMember
}

func GetChannelMemberRepo() repository.MongoRepository[models.ChannelMember] {
	ChannelMemberRepository = repository.MongoRepository[models.ChannelMember]{Model: db.ChannelMember}
	return ChannelMemberRepository
}
