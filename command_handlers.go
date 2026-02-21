package main

import (
	"errors"
	"fmt"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("Not enough arguments to login command. Please provide a username.")
	}

	username := cmd.args[0]
	err := s.cfg.SetUser(username)
	if err != nil {
		return err
	}

	fmt.Printf("User %s has logged in.\n", username)
	return nil
}

