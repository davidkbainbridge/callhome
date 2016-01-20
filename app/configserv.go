package main

import (
	"configserver"
	"log"
)

func main() {
	server := configserver.Server{
		ListenIP:               "0.0.0.0",
		ListenPort:             4321,
		ListenPath:             "callhome",
		ConfigurationDirectory: "/tmp",
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("failed to start configuration server: %s", err.Error())
	}
}
