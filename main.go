package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/michaeldakin/gator/internal/config"
	"github.com/michaeldakin/gator/internal/database"

	_ "github.com/lib/pq"
)

var (
	logger *slog.Logger
)

type State struct {
	cfg *config.Config
	db  *database.Queries
}

func main() {
	slogOpts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}
	logger = slog.New(slog.NewJSONHandler(os.Stdout, slogOpts))
	slog.SetDefault(logger)

	// Config
	cfg, err := config.Read()
	if err != nil {
		logger.Error("unable to read config", "error", err)
		os.Exit(1)
	}

	// Postgres
	db, err := sql.Open("postgres", cfg.DbUrl)
	if err != nil {
		logger.Error("unable to connect to postgres", "error", err)
		os.Exit(1)
	}
	dbQueries := database.New(db)

	newState := State{
		cfg: &cfg,
		db:  dbQueries,
	}

	// Command Handlers
	cmds := Commands{
		handlers: make(map[string]func(*State, Command) error),
	}

	// Register commands
	err = InitCommands(cmds)
	if err != nil {
		logger.Error("InitCommands", "error", err)
		os.Exit(1)
	}

	// Handle args
	userArgs := os.Args
    if len(userArgs) <= 1 {
        fmt.Println("Usage: ./gator <command> <arg [optional]>")
        fmt.Println()
        fmt.Println("Valid commands:")
        fmt.Println("   register - register new user")
        fmt.Println("   login    - login with an existing user")
        fmt.Println("   users    - display list of registered users")
        fmt.Println()
        os.Exit(1)
    }

	userInputCmd := userArgs[1]
	var userInputParam string

	if userInputCmd != "users" {
		userInputParam = userArgs[2]
	}

	if userInputCmd == "" {
		logger.Debug("args error", "len", len(userArgs), "args", userArgs[1:])
		logger.Error("not enough args provided")
		os.Exit(1)
	}

	logger.Debug("args provided.",
		slog.Group("args",
			slog.String("cmd", userInputCmd),
			slog.String("arg", userInputParam),
		),
	)

	// Handle user input
	// arg1 - Command
	// arg2 - Params [optional]
	switch userInputCmd {
	case "login":
		loginCmd := Command{
			name: userInputCmd,
			args: []string{userInputParam},
		}

		err := cmds.run(&newState, loginCmd)
		if err != nil {
			logger.Error("cmds.run", "error", err)
			os.Exit(1)
		}

		// Read the config file and get the new user
		cfg, err := config.Read()
		if err != nil {
			logger.Error("unable to read config", "error", err)
			os.Exit(1)
		}
		logger.Info("updated logged in user", "CurrentUserName", cfg.CurrentUserName)
	case "register":
		registerCmd := Command{
			name: userInputCmd,
			args: []string{userInputParam},
		}

		err := cmds.run(&newState, registerCmd)
		if err != nil {
			logger.Error("cmds.run", "error", err)
			os.Exit(1)
		}
	case "users":
		listUsersCmd := Command{
			name: userInputCmd,
		}
		err := cmds.run(&newState, listUsersCmd)
		if err != nil {
			logger.Error("cmds.run", "error", err)
			os.Exit(1)
		}
	default:
		logger.Error("invalid command")
	}
}

type Command struct {
	name string
	args []string
}

type Commands struct {
	handlers map[string]func(*State, Command) error
}

func InitCommands(cmds Commands) error {
	for c, h := range cmds.handlers {
		err := cmds.register(c, h)
		if err != nil {
			return fmt.Errorf("cmds.register error: %v\n", err)
		}
	}

	// err = cmds.register("register", registerHandler)
	// if err != nil {
	// 	return fmt.Errorf("cmds.register error: %v\n", err)
	// }
	//
	// err = cmds.register("users", usersHandler)
	// if err != nil {
	// 	return fmt.Errorf("cmds.register error: %v\n", err)
	// }

	return nil
}

// This method registers a new handler function for a command name.
func (c *Commands) register(name string, f func(*State, Command) error) error {
	if _, ok := c.handlers[name]; ok {
		return errors.New("handler function already exists")
	}

	c.handlers[name] = f
	logger.Info("registered command", "cmd", c.handlers[name])
	return nil
}

// This method runs a given command with the provided state if it exists.
func (c *Commands) run(state *State, cmd Command) error {
	logger.Debug(
		"cmd info",
		slog.Group("args",
			"cmd.name", cmd.name,
			"cmd.args", cmd.args,
		),
	)

	switch cmd.name {
	case "login":
		err := loginHandler(state, cmd)
		if err != nil {
			return err
		}
	case "register":
		err := registerHandler(state, cmd)
		if err != nil {
			return err
		}
	case "users":
		err := usersHandler(state, cmd)
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid command")
	}

	return nil
}

// gator login <username>
func loginHandler(s *State, c Command) error {
	logger.Debug("cfg output.",
		slog.Group("args",
			"cmd.name", c.name,
			"cmd.args", c.args,
		),
	)

	ctx := context.Background()
	cmdParam := c.args[0]
	userExists, err := s.db.GetUserByName(ctx, cmdParam)
	if err != nil {
		return err
	}

	if userExists.Name == "" {
		return errors.New("no name was provided!")
	}

	err = s.cfg.SetUser(c.args[0])
	if err != nil {
		return fmt.Errorf("unable to SetUser: %w\n", err)
	}
	logger.Info("User has been updated.", "user", c.args[0])

	return nil
}

// gator register <username>
func registerHandler(s *State, c Command) error {
	logger.Debug("registerHandler output",
		slog.Group("args",
			"cmd.name", c.name,
			"cmd.args", c.args,
		),
	)
	/*
	 * 1. Confirm arg2 was passed <username>
	 * 2. Create a new user in the database
	 *   -> It should have access to the CreateUser query through the state -> db struct.
	 *   -> Pass context.Background() to the query to create an empty Context argument.
	 *   -> Use the uuid.New() function to generate a new UUID for the user.
	 *   -> created_at and updated_at should be the current time.
	 *   -> Use the provided name.
	 *   -> Exit with code 1 if a user with that name already exists.
	 * 3. Set the current user in the config to the given name.
	 * 4. Print a message that the user was created, and log the user's data to the console for your own debugging.
	 */
	ctx := context.Background()
	cmdParam := c.args[0]

	userParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmdParam,
	}

	// Check if user already exists, if so return
	_, err := s.db.GetUser(ctx, userParams.ID)
	if !errors.Is(err, sql.ErrNoRows) {
		return errors.New("user exists")
	}

	if err != nil {
		return err
	}

	newUser, err := s.db.CreateUser(ctx, userParams)
	if err != nil {
		return err
	}

	err = s.cfg.SetUser(newUser.Name)
	if err != nil {
		return err
	}

	logger.Debug("user created", "name", newUser.Name, "uuid", newUser.ID, "created_at", newUser.CreatedAt, "last_updated", newUser.UpdatedAt)

	return nil
}

// gator users
func usersHandler(s *State, c Command) error {
	ctx := context.Background()
	users, err := s.db.GetAllUsers(ctx)
	if err != nil {
		return err
	}

	for _, user := range users {
		fmt.Printf("%+v\n", user)
	}

	return nil
}
