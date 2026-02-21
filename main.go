package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dylansbyrd/gator/internal/config"
)

type state struct {
	cfg *config.Config
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Unable to read config due to error: %v", err)
	}
	fmt.Printf("Read config: %+v\n", cfg)

	s := state{&cfg}
	cmds := commands{make(map[string]cmdHandlerFunc)}
	err = cmds.register("login", handlerLogin)
	if err != nil {
		log.Fatalf("Failed registering command %s: %v", "login", err)
	}

	programArgs := os.Args
	if len(programArgs) < 2 {
		log.Fatalf("Not enough arguments. Please provide a command name to execute.")
	}

	cmdName := programArgs[1]
	cmdArgs := programArgs[2:]
	err = cmds.run(&s, command{cmdName, cmdArgs})
	if err != nil {
		log.Fatalf("Error running command %s: %v", cmdName, err)
	}
}
