package main

import (
	"guoshi/app"
	"log"
)

func main() {
	if err := app.GetApp().Prepare(); err != nil {
		log.Fatal(err)
	}

	server := app.GetApp().Config.Server
	if err := Listen(server.Address, server.Port); err != nil {
		panic(err)
	}
}
