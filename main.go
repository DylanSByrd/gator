package main

import (
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

	s := state{&cfg}

	cmds := commands{make(map[string]cmdHandlerFunc)}
	cmds.register("login", handlerLogin)

	programArgs := os.Args
	if len(programArgs) < 2 {
		log.Fatalf("Usage: cli <command> [args...]")
	}

	cmdName := programArgs[1]
	cmdArgs := programArgs[2:]
	err = cmds.run(&s, command{cmdName, cmdArgs})
	if err != nil {
		log.Fatalf("Error running command %s: %v", cmdName, err)
	}
}
