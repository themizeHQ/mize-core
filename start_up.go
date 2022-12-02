package main

import (
	"mize.app/db"
	"mize.app/emitter"
	errormonitoring "mize.app/error_monitoring"
	"mize.app/logger"
	"mize.app/realtime"
	redis "mize.app/repository/database/redis"
)

func StartServices() {
	// initialise error monitoring
	errormonitoring.Initialise()
	// set up logger
	logger.InitializeLogger()
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
