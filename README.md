# Gator
Simple RSS feed aggregator CLI app written in Go. Part of Boot.Dev coursework.

## Requirements
Requires Go 1.25.5+
Requires Postgres version 15+
Built and tested for WSL

## Installation
To install, download the repository and run `go install` inside the directory.

## Configuration
Gator configuration is specified in a `.gatorconfig.json` file stored in the home directory. The following fields are
supported:
- `db_url` - the url for the Postgres database used to store information.
- (Optional) `current_user_name` - set by Gator when "logging in", this is the username for the current user.

## Usage
Gator is driven by CLI commands. Except for the `agg` command, all commands execute immediately to manage and browse RSS
feed entries. `agg` is used as a long-running process to fetch posts from RSS feeds at a regular cadence. The general idea
is there's a `gator agg` process running constantly to fetch post data while separate `gator` commands are fired to manage
the data.

### Commands
- `gator register <username>` - Creates a new user with `username`
- `gator login <username>` - Switches the current user to `username`
- `gator users` - Lists all registered users
- `gator addfeed <feed_name> <feed_url>` - Fetches the RSS feed located at `feed_url` and stores it as `feed_name`. The
  current user automatically follows the added feed
- `gator feeds` - List all added feeds
- `gator follow <feed_url>` - Causes the current user to follow the specified feed url if it is added
- `gator unfollow <feed_url>` - Causes the current user to stop following the specified feed url
- `gator following` - Lists all feeds followed by the current user
- `gator browse <post_limit=2>` - Lists up to `post_limit` RSS feed entries order by most recent
- `gator agg <time_between_requests` - Fetches all available posts from the oldest updated feed every `time_between_requests`.
  `time_between_requests` is specified as a duration string, e.g. `1s`, `1m`, `1h` etc. Please try not to spam requests
  and DOS anyone's servers.
- `gator reset` - Clears all information (i.e. users, feeds, and posts) from the database
