package main

import (
	"mize.app/db"
	"mize.app/realtime"
	redis "mize.app/repository/database/redis"
)

func StartServices() {
	// connect to the databases
	db.ConnectToDb()
	// set up redis
	redis.SetUpRedisRepo()
	// initialise centrifugo
	realtime.InitialiseCentrifugoController()
}

func CleanUp() {
	// clean up resources
	db.CleanUp()
}
