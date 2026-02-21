package main

import (
	"fmt"
)

type command struct {
	Name string
	Args []string
}

type cmdHandlerFunc func(*state, command) error

type commands struct {
	handlers map[string]cmdHandlerFunc
}

func (cmds *commands) register(name string, handler cmdHandlerFunc) {
	cmds.handlers[name] = handler
}

func (cmds *commands) run(s *state, cmd command) error {
	handler, exists := cmds.handlers[cmd.Name]
	if !exists {
		return fmt.Errorf("Not such command '%s'", cmd.Name)
	}

	return handler(s, cmd)
}

