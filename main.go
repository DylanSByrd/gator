package main

import (
	"log"
	"os"
	"database/sql"

	"github.com/dylansbyrd/gator/internal/config"
	"github.com/dylansbyrd/gator/internal/database"

	_ "github.com/lib/pq"
)

type state struct {
	cfg *config.Config
	db  *database.Queries
}

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Unable to read config due to error: %v", err)
	}

	db, err := sql.Open("postgres", cfg.DbUrl)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer db.Close()
	dbQueries := database.New(db)

	s := state{
		cfg: &cfg,
		db: dbQueries,
	}

	cmds := commands{make(map[string]cmdHandlerFunc)}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", handlerAddFeed)
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", handlerFollow)
	cmds.register("following", handlerFollowing)

	programArgs := os.Args
	if len(programArgs) < 2 {
		log.Fatalf("Usage: cli <command> [args...]")
	}

	cmdName := programArgs[1]
	cmdArgs := programArgs[2:]
	err = cmds.run(&s, command{cmdName, cmdArgs})
	if err != nil {
		log.Fatalf("%v", err)
	}
}
