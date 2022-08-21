package main

import (
	"mize.app/db"
	"mize.app/emitter"
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
	// initialiae emitter listener
	emitter.EmitterListener()
}

func CleanUp() {
	// clean up resources
	db.CleanUp()
}
