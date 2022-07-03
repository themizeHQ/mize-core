package user

import (
	workspace "mize.app/app/workspace/models"
	db "mize.app/db/mongo"
	repository "mize.app/repository/database/mongo"
)

var WorkspaceRepository repository.MongoRepository[workspace.Workspace]
var WorkspaceInviteRepository repository.MongoRepository[workspace.WorkspaceInvite]
var ChannelRepository repository.MongoRepository[workspace.Channel]
var WorkspaceMember repository.MongoRepository[workspace.WorkspaceMember]

func GetWorkspaceRepo() repository.MongoRepository[workspace.Workspace] {
	WorkspaceRepository = repository.MongoRepository[workspace.Workspace]{Model: db.WorkspaceModel}
	return WorkspaceRepository
}

func GetWorkspaceInviteRepo() repository.MongoRepository[workspace.WorkspaceInvite] {
	WorkspaceInviteRepository = repository.MongoRepository[workspace.WorkspaceInvite]{Model: db.WorkspaceInvite}
	return WorkspaceInviteRepository
}

func GetChannelRepo() repository.MongoRepository[workspace.Channel] {
	ChannelRepository = repository.MongoRepository[workspace.Channel]{Model: db.Channel}
	return ChannelRepository
}

func GetWorkspaceMember() repository.MongoRepository[workspace.WorkspaceMember] {
	WorkspaceMember = repository.MongoRepository[workspace.WorkspaceMember]{Model: db.WorkspaceMember}
	return WorkspaceMember
}
