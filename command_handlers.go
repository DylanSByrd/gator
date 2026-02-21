package main

import (
	"fmt"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("Usage: %s <name>", cmd.Name)
	}

	username := cmd.Args[0]
	err := s.cfg.SetUser(username)
	if err != nil {
		return fmt.Errorf("failed to set current user: %w", err)
	}

	fmt.Printf("User %s has logged in.\n", username)
	return nil
}

