package main

import (
	"context"
	"fmt"
	"log"
	"time"
	"errors"

	"github.com/dylansbyrd/gator/internal/database"
	"github.com/google/uuid"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Usage: %s <name>", cmd.Name)
	}

	username := cmd.Args[0]
	user, err := s.db.GetUser(context.Background(), username);
	if err != nil {
		return fmt.Errorf("couldn't find user: %w", err)
	}

	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("failed to set current user: %w", err)
	}

	fmt.Printf("User %s has logged in.\n", username)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Usage: %s <name>", cmd.Name)
	}

	username := cmd.Args[0]
	user, err := s.db.CreateUser(context.Background(), 
		database.CreateUserParams{
			ID: uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			Name: username,
		});

	if err != nil {
		return fmt.Errorf("failed to create user %s: %w", username, err)
	}

	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("failed to set current user: %w", err)
	}

	fmt.Printf("User %s created.\n", username)
	log.Printf("New user: %#v", user)
	return nil
}


func handlerReset(s* state, cmd command) error {
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to reset: %w", err)
	}

	fmt.Printf("Database reset.\n")
	s.cfg.SetUser("")
	return nil
}

func handlerUsers(s* state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to fetch users from database: %w", err)
	}

	for _, user := range users {
		if user.Name == s.cfg.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}
	return nil;
}

func handlerAgg(s* state, cmd command) error {
	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("failed to fetch feed: %w", err)
	}

	fmt.Printf("%#v\n", feed)

	return nil
}

func handlerAddFeed(s* state, cmd command) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("Usage: %s <name> <url>", cmd.Name)
	}

	currentUsername := s.cfg.CurrentUserName
	if currentUsername == "" {
		return errors.New("No current user. Please log in.")
	}

	ctx := context.Background()
	currentUser, err := s.db.GetUser(ctx, currentUsername)
	if err != nil {
		return fmt.Errorf("Failed getting current user: %w", err)
	}

	feed, err := s.db.CreateFeed(ctx,
		database.CreateFeedParams{
			ID: uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			Name: cmd.Args[0],
			Url: cmd.Args[1],
			UserID: currentUser.ID,
		},
	)
	if err != nil {
		return fmt.Errorf("Error creating feed: %w", err)
	}

	feedFollow, err := s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID: currentUser.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("Failed to follow feed: %w", err)
	}

	fmt.Printf("Feed %s created.\n", feed.Name)
	log.Printf("New feed: %#v\n", feed)
	fmt.Println("Feed followed successfully.")
	log.Printf("Feed follow: %#v", feedFollow)

	return nil
}

func handlerFeeds(s* state, cmd command) error {
	ctx := context.Background()
	feeds, err := s.db.GetFeeds(ctx)
	if err != nil {
		return fmt.Errorf("Error fetching feeds: %w", err)
	}

	for _, feed := range feeds {
		fmt.Printf("*%s*\n", feed.Name)
		fmt.Printf("%s\n", feed.Url)
		
		user, err := s.db.GetUserById(ctx, feed.UserID)
		if err != nil {
			// I don't love erroring out when we're mid print. May be worth storing the results and printing later?
			return fmt.Errorf("Failed to find user with id %v: %w", feed.UserID, err)
		}
		fmt.Printf("Created by %s\n", user.Name)
		fmt.Println()
	}

	return nil
}

func handlerFollow(s* state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Usage: %s <url>", cmd.Name)
	}

	currentUsername := s.cfg.CurrentUserName
	if currentUsername == "" {
		return errors.New("No current user. Please log in.")
	}

	ctx := context.Background()
	currentUser, err := s.db.GetUser(ctx, currentUsername)
	if err != nil {
		return fmt.Errorf("Failed getting current user: %w", err)
	}

	feedUrl := cmd.Args[0]
	feed, err := s.db.GetFeedByUrl(ctx, feedUrl)
	if err != nil {
		return fmt.Errorf("Failed to find feed %s: %w", feedUrl, err)
	}

	feedFollowRow, err := s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID: currentUser.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("Failed to follow feed: %w", err)
	}

	fmt.Printf("User %s successfully followed feed %s.\n", feedFollowRow.UserName, feedFollowRow.FeedName)
	return nil
}

func handlerFollowing(s* state, cmd command) error {
	currentUsername := s.cfg.CurrentUserName
	if currentUsername == "" {
		return errors.New("No current user. Please log in.")
	}

	ctx := context.Background()
	currentUser, err := s.db.GetUser(ctx, currentUsername)
	if err != nil {
		return fmt.Errorf("Failed getting current user: %w", err)
	}

	followedFeeds, err := s.db.GetFeedFollowsForUser(ctx, currentUser.ID)
	if err != nil {
		return fmt.Errorf("Failed getting feeds for current user: %w", err)
	}

	if len(followedFeeds) == 0 {
		fmt.Printf("User %s is not currently following any feeds.\n", currentUsername)
		return nil
	}

	fmt.Printf("Feed follows for user %s:\n", currentUsername)
	for _, followedFeed := range followedFeeds {
		fmt.Printf("%s\n", followedFeed.FeedName)
	}
	return nil
}
