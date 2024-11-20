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

type Commands struct {
	handlers map[string]func(*State, Command) error
}

type Command struct {
	name string
	args []string
}

func orError(s string, err error) {
	if err != nil {
		logger.Error(s, "error", err)
		os.Exit(1)
	}
}

func newLogger() {
}

func main() {
	// Logger
	slogOpts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}
	logger = slog.New(slog.NewJSONHandler(os.Stdout, slogOpts))
	slog.SetDefault(logger)

	// Args
	// minArgs := 2
	// slog.Debug("args", "min_args", minArgs, "args_provided", len(userArgs))
	// if len(userArgs) < minArgs {
	// 	slog.Error("not enough args", "len", len(userArgs), "min_args", minArgs)
	// 	os.Exit(1)
	// }

	// Config
	cfg, err := config.Read()
	if err != nil {
		logger.Error("unable to read config", "error", err)
		os.Exit(1)
	}

	// Postgres DB
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

	cmds := Commands{
		handlers: make(map[string]func(*State, Command) error),
	}

	userArgs := os.Args
	userInputCmd := userArgs[1]
    var userInputParam string
	if userInputCmd != "users" {
		userInputParam = userArgs[2]
	}

	if userInputCmd == "" { // || userInputParam == ""
		slog.Debug("args error", "len", len(userArgs), "args", userArgs[1:])
		logger.Error("not enough args provided")
		os.Exit(1)
	}

	slog.Debug("args provided.",
		slog.Group("args",
			slog.String("cmd", userInputCmd),
			slog.String("arg", userInputParam),
		),
	)

	// Register functions
	err = InitCommands(cmds)
	if err != nil {
		logger.Error("InitCommands", "error", err)
		os.Exit(1)
	}

	// Handle user input
	// arg1 - command <login|register>
	// arg2 - Params <username>
	switch userInputCmd {
	case "login":
		// Create Command struct
		loginCmd := Command{
			name: userInputCmd,
			args: []string{userInputParam},
		}

		// Run the command
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

		slog.Info("updated logged in user", "CurrentUserName", cfg.CurrentUserName)
	case "register":
		slog.Info("cmd: register")
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
		slog.Info("cmd: list users")
	default:
		logger.Error("invalid argument")
	}

}

func InitCommands(cmds Commands) error {
	err := cmds.register("login", loginHandler)
	if err != nil {
		return fmt.Errorf("cmds.register error: %v\n", err)
	}

	err = cmds.register("register", registerHandler)
	if err != nil {
		return fmt.Errorf("cmds.register error: %v\n", err)
	}

	err = cmds.register("users", usersHandler)
	if err != nil {
		return fmt.Errorf("cmds.register error: %v\n", err)
	}

	return nil
}

// This method registers a new handler function for a command name.
func (c *Commands) register(name string, f func(*State, Command) error) error {
	if _, ok := c.handlers[name]; ok {
		return errors.New("handler function already exists")
	}

	c.handlers[name] = f
	return nil
}

// his method runs a given command with the provided state if it exists.
func (c *Commands) run(s *State, cmd Command) error {
	logger.Debug(
		"cmd info",
		slog.Group("args",
			"cmd.name", cmd.name,
			"cmd.args", cmd.args,
		),
	)
	// _, ok := c.handlers[cmd.name]
	// if !ok {
	// 	return errors.New("invalid argument")
	// }

	switch cmd.name {
	case "login":
		err := loginHandler(s, cmd)
		if err != nil {
			return err
		}
	case "register":
		err := registerHandler(s, cmd)
		if err != nil {
			return err
		}
	case "users":
		err := usersHandler(s, cmd)
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
	slog.Debug("cfg output.",
		slog.Group("args",
			"cmd.name", c.name,
			"cmd.args", c.args,
		),
	)
	err := s.cfg.SetUser(c.args[0])
	if err != nil {
		return fmt.Errorf("unable to SetUser: %w\n", err)
	}
	logger.Info("User has been updated.", "user", c.args[0])

	return nil
}

// gator register <username>
func registerHandler(s *State, c Command) error {
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
	cmdParam := c.args[2]

	userParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmdParam,
	}

	// Check if user already exists, if so return
	userExists, err := s.db.GetUser(ctx, userParams.ID)
	if err != nil {
		return err
	}
	if userExists.Name != "" {
		return fmt.Errorf("user %q already exists in database", userExists.Name)
	}

	newUser, err := s.db.CreateUser(ctx, userParams)
	if err != nil {
		return err
	}

	fmt.Printf("user created and inserted into DB %+v\n", newUser)

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
		fmt.Printf("user: %+v\n", user)
	}

	return nil
}
