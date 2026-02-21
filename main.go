package main

import (
	"log"

	"github.com/dylansbyrd/gator/internal/config"
)

func main() {
	conf, err := config.Read()
	if err != nil {
		log.Fatalf("Unable to read config due to error: %v", err)
	}
	conf.Print()

	log.Println("Setting user...")
	conf.SetUser("dylansbyrd")
	conf.Print()
}
