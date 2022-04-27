package main

import (
	"mize.app/db"
	redis "mize.app/repository/database/redis"

)

func StartServices() {
	// connect to the databases
	db.ConnectToDb()
	// set up redis
	redis.SetUpRedisRepo()
}

func CleanUp() {
	// clean up resources
	db.CleanUp()
}
