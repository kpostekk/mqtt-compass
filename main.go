package main

import (
	"log"
	compass "mqui/compass"
)

func main() {
	log.Default().Println("Starting Compass...")
	// compass.StartHeadless()

	compass.StartUi()
}
