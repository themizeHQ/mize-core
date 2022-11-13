package repository

import (
	"sync"

	"mize.app/app/teams/models"
	dbMongo "mize.app/db/mongo"
	mongoRepo "mize.app/repository/database/mongo"
)

var once = sync.Once{}

var TeamRepository mongoRepo.MongoRepository[models.Team]

var TeamMembersRepository mongoRepo.MongoRepository[models.TeamMembers]

func GetTeamRepo() mongoRepo.MongoRepository[models.Team] {
	once.Do(func() {
		TeamRepository = mongoRepo.MongoRepository[models.Team]{Model: dbMongo.Team}
	})
	return TeamRepository
}

func GetTeamMemberRepo() mongoRepo.MongoRepository[models.TeamMembers] {
	TeamMembersRepository = mongoRepo.MongoRepository[models.TeamMembers]{Model: dbMongo.TeamMember}
	return TeamMembersRepository
}
