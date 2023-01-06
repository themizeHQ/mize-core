package main

import (
	"mize.app/db"
	"mize.app/emitter"
	errormonitoring "mize.app/error_monitoring"
	eventsqueue "mize.app/events_queue"
	"mize.app/logger"
	"mize.app/realtime"
	redis "mize.app/repository/database/redis"
)

func StartServices() {
	// set up logger
	logger.InitializeLogger()
	// initialise error monitoring
	errormonitoring.Initialise()
	// connect to the databases
	db.ConnectToDb()
	// set up redis
	redis.SetUpRedisRepo()
	// initialise centrifugo
	realtime.InitialiseCentrifugoController()
	// initialiae emitter listener
	emitter.EmitterListener()

	eventsqueue.ConnectToEventStream()

	logger.Info("all services started")
}

func CleanUp() {
	// clean up resources
	db.CleanUp()

	eventsqueue.CloseProducerClient()

	logger.Info("all services and resources offline")
}
