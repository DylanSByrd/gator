package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/dylansbyrd/gator/internal/database"
	"github.com/google/uuid"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Usage: %v <name>", cmd.Name)
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
		return fmt.Errorf("Usage: %v <name>", cmd.Name)
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

func scrapeFeeds(s* state) error {
	feedDAO, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("Failed to find next feed to fetch: %w", err)
	}

	// Mark the feed as fetched even if it fails so we don't get stuck trying to fetch the same feed forever
	// for all we know the feed at the url no longer exists
	_, err = s.db.MarkFeedFetched(context.Background(), feedDAO.ID)
	if err != nil {
		return fmt.Errorf("Failed to mark feed %s as fetched: %w", feedDAO.Name, err)
	}

	feed, err := fetchFeed(context.Background(), feedDAO.Url)
	if err != nil {
		return fmt.Errorf("Failed to fetch %s: %w", feedDAO.Name, err)
	}

	for _, post := range feed.Channel.Items {
		timestamp := sql.NullTime{}
		if parsedTime, err := time.Parse(time.RFC1123Z, post.PubDate); err == nil {
			timestamp = sql.NullTime{
				Time: parsedTime,
				Valid: true,
			}
		}

		_, err := s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID: uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			Title: post.Title,
			Url: post.Link,
			Description: sql.NullString{
				String: post.Description,
				Valid: true,
			},
			PublishedAt: timestamp,
			FeedID: feedDAO.ID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				continue
			} else {
				log.Printf("Failed to create post: %v", err)
			}
		}
	}

	fmt.Printf("Feed %s collected, %v posts found", feedDAO.Name, len(feed.Channel.Items))
	return nil
}

func handlerAgg(s* state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Usage: %v <time_between_requests>", cmd.Name)
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("Failed to start aggregating due to parse error: %w", err)
	}

	fmt.Printf("Collecting feeds every %s\n", timeBetweenRequests.String()) 
	ticker := time.NewTicker(timeBetweenRequests)
	for ;; <-ticker.C {
		scrapeFeeds(s)
	}
}

func handlerUserAddFeed(s* state, cmd command, user database.User) error {
	if len(cmd.Args) != 2 {
		return fmt.Errorf("Usage: %v <name> <url>", cmd.Name)
	}

	feed, err := s.db.CreateFeed(context.Background(),
		database.CreateFeedParams{
			ID: uuid.New(),
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
			Name: cmd.Args[0],
			Url: cmd.Args[1],
			UserID: user.ID,
		},
	)
	if err != nil {
		return fmt.Errorf("Error creating feed: %w", err)
	}

	feedFollow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID: user.ID,
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
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("Error fetching feeds: %w", err)
	}

	for _, feed := range feeds {
		fmt.Printf("*%s*\n", feed.Name)
		fmt.Printf("%s\n", feed.Url)
		
		user, err := s.db.GetUserById(context.Background(), feed.UserID)
		if err != nil {
			// I don't love erroring out when we're mid print. May be worth storing the results and printing later?
			return fmt.Errorf("Failed to find user with id %v: %w", feed.UserID, err)
		}
		fmt.Printf("Created by %s\n", user.Name)
		fmt.Printf("Last fetched at %v\n", feed.LastFetchedAt.Time)
		fmt.Println()
	}

	return nil
}

func handlerUserFollow(s* state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Usage: %v <url>", cmd.Name)
	}

	feedUrl := cmd.Args[0]
	feed, err := s.db.GetFeedByUrl(context.Background(), feedUrl)
	if err != nil {
		return fmt.Errorf("Failed to find feed %s: %w", feedUrl, err)
	}

	feedFollowRow, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("Failed to follow feed: %w", err)
	}

	fmt.Printf("User %s successfully followed feed %s.\n", feedFollowRow.UserName, feedFollowRow.FeedName)
	return nil
}

func handlerUserFollowing(s* state, cmd command, user database.User) error {
	followedFeeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("Failed getting feeds for current user: %w", err)
	}

	if len(followedFeeds) == 0 {
		fmt.Printf("User %s is not currently following any feeds.\n", user.Name)
		return nil
	}

	fmt.Printf("Feed follows for user %s:\n", user.Name)
	for _, followedFeed := range followedFeeds {
		fmt.Printf("%s\n", followedFeed.FeedName)
	}
	return nil
}

func handlerUserUnfollow(s* state, cmd command, user database.User) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("Usage: %v <url>", cmd.Name)
	}

	feedUrl := cmd.Args[0]
	feed, err := s.db.GetFeedByUrl(context.Background(), feedUrl)
	if err != nil {
		return fmt.Errorf("Failed to find feed %s: %w", feedUrl, err)
	}

	err = s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("Failed to unfollow: %w", err)
	}

	fmt.Printf("%s unfollowed successfully!\n", feed.Name)
	return nil
}

func handlerUserBrowse(s* state, cmd command, user database.User) error {
	postCount := 2
	if len(cmd.Args) > 0 {
		if parsed, err := strconv.Atoi(cmd.Args[0]); err == nil {
			postCount = parsed
		} else {
			return fmt.Errorf("Usage: %s <optional: number_of_posts>", cmd.Name)
		}
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit: int32(postCount),
	})
	if err != nil {
		return fmt.Errorf("Failed to get posts from database: %w", err)
	}

	if len(posts) == 0 {
		fmt.Printf("No posts found for user %s\n", user.Name)
		return nil
	}

	fmt.Printf("Found %d post(s) for user %s:\n", len(posts), user.Name)
	for _, post := range posts {
		if post.PublishedAt.Valid {
			fmt.Printf("%s ", post.PublishedAt.Time.Format("Mon Jan 2"))
		}
		fmt.Printf("from %s:\n", post.FeedName)
		fmt.Printf("--- %s ---\n", post.Title)

		if post.Description.Valid {
			fmt.Printf("    %s\n", post.Description.String)
		} else {
			fmt.Printf("    (no description)\n")
		}

		fmt.Printf("Link: %s\n", post.Url)
		fmt.Println("========================================")
	}
	return nil
}

