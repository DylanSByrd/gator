package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/dylansbyrd/gator/internal/database"
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

type userCmdHandlerFunc func(s *state, cmd command, user database.User) error

func middlewareLoggedIn (userCmdHandler userCmdHandlerFunc) cmdHandlerFunc {
	return func(s* state, cmd command) error {
		currentUsername := s.cfg.CurrentUserName
		if currentUsername == "" {
			return errors.New("No current user. Please log in.")
		}

		currentUser, err := s.db.GetUser(context.Background(), currentUsername)
		if err != nil {
			return fmt.Errorf("Failed getting current user: %w", err)
		}

		return userCmdHandler(s, cmd, currentUser)
	}
}

