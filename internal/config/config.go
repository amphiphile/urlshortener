package config

import (
	"fmt"
	"os"
)

type Config struct {
	ServerURLConfig ServerURLConfig
	AppConfig       AppConfig
	DBConfig        DBConfig
}

type ServerURLConfig struct {
	ServerAddress string
}

type AppConfig struct {
	BaseURL string
}
type DBConfig struct {
	DBFileName string
}

func (c *ServerURLConfig) String() string {
	return fmt.Sprintf(c.ServerAddress)

}
func (c *ServerURLConfig) Set(flagValue string) error {
	c.ServerAddress = flagValue
	return nil
}

func (c *AppConfig) String() string {
	return fmt.Sprintf(c.BaseURL)

}
func (c *AppConfig) Set(flagValue string) error {
	c.BaseURL = flagValue
	return nil
}

func GetFromEnv(variable string, defaultValue string) string {
	if value, exists := os.LookupEnv(variable); exists {
		return value
	}
	return defaultValue
}
