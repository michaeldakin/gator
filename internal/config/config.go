package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

/*
* Export a Read function that reads the JSON file found at ~/.gatorconfig.json and returns a Config struct.
* It should read the file from the HOME directory, then decode the JSON string into a new Config struct.
* I used os.UserHomeDir to get the location of HOME.
 */
func Read() (Config, error) {
	jsonPath, err := getConfigFilePath()
	if err != nil {
		return Config{}, fmt.Errorf("failed to getConfigFilePath: %w\n", err)
	}
	// slog.Info("got json filepath", "filepath", jsonPath)

	file, err := os.ReadFile(jsonPath)
	if err != nil {
		return Config{}, fmt.Errorf("failed to read file: %w\n", err)
	}

	cfg := Config{}
	err = json.Unmarshal(file, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("failed to Unmarshal data: %w\n", err)
	}

	return cfg, nil
}

func getConfigFilePath() (string, error) {
	pwd, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	jsonPath := fmt.Sprintf("%s/%s", pwd, configFileName)
	_, err = os.Stat(jsonPath)
	if err != nil {
		return "", err
	}

	return jsonPath, nil
}

/*
* Export a SetUser method on the Config struct that writes the config struct
* to the JSON file after setting the current_user_name field.
 */
func (c Config) SetUser(name string) error {
	c.CurrentUserName = name
	// slog.Info("SetUser()", "DbUrl", c.DbUrl, "CurrentUserName", c.CurrentUserName)

	err := write(c)
	if err != nil {
		return err
	}
	return nil
}

func write(cfg Config) error {
	jsonPath, err := getConfigFilePath()
	if err != nil {
		return fmt.Errorf("failed to getConfigFilePath: %w\n", err)
	}

	// slog.Info("write()", "openfile", jsonPath)
	// file, err := os.OpenFile(jsonPath, os.O_RDWR, 0644)
	// if err != nil {
	// 	return fmt.Errorf("failed to read file: %w\n", err)
	// }
	// defer file.Close()

	// encoder := json.NewEncoder(file)
	//    encoder.SetIndent("", "    ")
	//    err = encoder.Encode(cfg)
	jsonData, err := json.Marshal(cfg)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer

	err = json.Indent(&buf, jsonData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode data to file: %w\n", err)
	}

	err = os.WriteFile(jsonPath, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}
