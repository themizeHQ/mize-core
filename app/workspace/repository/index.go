package user

import (
	"sync"

	"mize.app/app/workspace/models"
	"mize.app/db/mongo"
	"mize.app/repository/database/mongo"
)

var once = sync.Once{}

var WorkspaceRepository repository.MongoRepository[workspace.Workspace]
var WorkspaceInviteRepository repository.MongoRepository[workspace.WorkspaceInvite]
var ChannelRepository = repository.MongoRepository[workspace.Channel]{Model: db.Channel}

func GetWorkspaceRepo() repository.MongoRepository[workspace.Workspace] {
	once.Do(func() {
		WorkspaceRepository = repository.MongoRepository[workspace.Workspace]{Model: db.WorkspaceModel}
	})
	return WorkspaceRepository
}

func GetWorkspaceInviteRepo() repository.MongoRepository[workspace.WorkspaceInvite] {
	WorkspaceInviteRepository = repository.MongoRepository[workspace.WorkspaceInvite]{Model: db.WorkspaceInvite}
	return WorkspaceInviteRepository
}

func GetChannelRepo() repository.MongoRepository[workspace.Channel] {
	once.Do(func() {
		ChannelRepository = repository.MongoRepository[workspace.Channel]{Model: db.Channel}
	})
	return ChannelRepository
}
