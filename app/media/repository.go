package media

import (
	"sync"

	dbMongo "mize.app/db/mongo"
	mongoRepo "mize.app/repository/database/mongo"
)

var once = sync.Once{}

var UploadRepository mongoRepo.MongoRepository[Upload]

func GetUploadRepo() mongoRepo.MongoRepository[Upload] {
	once.Do(func() {
		UploadRepository = mongoRepo.MongoRepository[Upload]{Model: dbMongo.Upload}
	})
	return UploadRepository
}
