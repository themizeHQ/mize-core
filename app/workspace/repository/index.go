package user

import (
	"mize.app/app/workspace/models"
	"mize.app/db/mongo"
	"mize.app/repository/database/mongo"
)

var WorkspaceRepository = repository.MongoRepository[workspace.Workspace]{Model: &db.WorkspaceModel}
