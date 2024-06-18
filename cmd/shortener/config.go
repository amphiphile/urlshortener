package main

import (
	"flag"
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

func parseConfig(cfg *Config) error {
	flag.Var(&cfg.ServerURLConfig, "a", "HTTP server startup address")
	flag.Var(&cfg.AppConfig, "b", "Base address of the shortened URL")
	flag.Parse()
	if serverAddress := cfg.ServerURLConfig.String(); serverAddress == "" {
		if err := cfg.ServerURLConfig.Set(GetFromEnv("SERVER_ADDRESS", "localhost:8080")); err != nil {
			return err
		}
	}
	if baseURL := cfg.AppConfig.String(); baseURL == "" {
		if err := cfg.AppConfig.Set(GetFromEnv("BASE_URL", "http://localhost:8080/")); err != nil {
			return err
		}
	}

	if dbFileName := cfg.DBConfig.String(); dbFileName == "" {
		if err := cfg.DBConfig.Set("db.json"); err != nil {
			return err
		}
	}
	return nil
}

func GetFromEnv(variable string, defaultValue string) string {
	if value, exists := os.LookupEnv(variable); exists {
		return value
	}
	return defaultValue
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

func (c *DBConfig) String() string {
	return fmt.Sprintf(c.DBFileName)

}
func (c *DBConfig) Set(flagValue string) error {
	c.DBFileName = flagValue
	return nil
}
