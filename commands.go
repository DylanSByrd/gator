package main

import (
	"fmt"
)

type command struct {
	name string
	args []string
}

type cmdHandlerFunc func(*state, command) error

type commands struct {
	handlers map[string]cmdHandlerFunc
}

func (cmds *commands) run(s *state, cmd command) error {
	handler, exists := cmds.handlers[cmd.name]
	if !exists {
		return fmt.Errorf("Not such command '%s'", cmd.name)
	}

	return handler(s, cmd)
}

func (cmds *commands) register(name string, handler cmdHandlerFunc) error {
	if exists := cmds.handlers[name]; exists != nil {
		return fmt.Errorf("Attempted to register duplicate command '%s'", name)
	}

	cmds.handlers[name] = handler
	return nil
}
