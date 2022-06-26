package user

import (
	"mize.app/app/workspace/models"
	"mize.app/db/mongo"
	"mize.app/repository/database/mongo"
)

var WorkspaceRepository = repository.MongoRepository[workspace.Workspace]{Model: &db.WorkspaceModel}
var WorkspaceInviteRepository = repository.MongoRepository[workspace.WorkspaceInvite]{Model: &db.WorkspaceInvite}
var ChannelRepository = repository.MongoRepository[workspace.Channel]{Model: &db.Channel}
