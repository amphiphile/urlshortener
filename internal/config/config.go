package config

import (
	"fmt"
	"os"
)

type Config struct {
	ServerUrlConfig ServerUrlConfig
	AppConfig       AppConfig
	DBConfig        DBConfig
}

type ServerUrlConfig struct {
	ServerAddress string
}

type AppConfig struct {
	BaseUrl string
}
type DBConfig struct {
	DBFileName string
}

func (c *ServerUrlConfig) String() string {
	return fmt.Sprintf(c.ServerAddress)

}
func (c *ServerUrlConfig) Set(flagValue string) error {
	c.ServerAddress = flagValue
	return nil
}

func (c *AppConfig) String() string {
	return fmt.Sprintf(c.BaseUrl)

}
func (c *AppConfig) Set(flagValue string) error {
	c.BaseUrl = flagValue
	return nil
}

func GetFromEnv(variable string, defaultValue string) string {
	if value, exists := os.LookupEnv(variable); exists {
		return value
	}
	return defaultValue
}
