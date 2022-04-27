package main

import (
	"mize.app/db"
)

func StartServices() {
	// connect to the databases
	db.ConnectToDb()
}

func CleanUp() {
	// clean up resources
	db.CleanUp()
}
