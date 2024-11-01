package config

const configFileName = ".gatorconfig.json"

type Config struct {
	DbUrl           string
	CurrentUserName string
}

func getConfigFilePath() (string, error) {
    return "", nil
}

func Read() Config {
	// read ~/.gatorconfig.json
	// os.UserHomeDir + configFileName
    // json.Unmarshal / decode into Config
	return Config{}
}

func write(cfg Config) error {
    return nil
}

func (c *Config) SetUser(name string) {
    // Take in `user` name
    // Set c.CurrentUserName = name
    // Marshal data to json
    // Write file
}

