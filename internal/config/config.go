package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Config struct {
	ServerUrlConfig ServerUrlConfig
	AppConfig       AppConfig
	DBFileName      string
}

type ServerUrlConfig struct {
	ServerHost string
	ServerPort int
}

type AppConfig struct {
	BaseUrl string
}

func (c ServerUrlConfig) String() string {
	return fmt.Sprintf("%s:%d", c.ServerHost, c.ServerPort)

}
func (c ServerUrlConfig) Set(flagValue string) error {
	if flagValue == "" {
		flagValue = "localhost:8080"
	}
	hp := strings.Split(flagValue, ":")
	if len(hp) != 2 {
		return errors.New("need address in a form host:port")
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}
	c.ServerHost = hp[0]
	c.ServerPort = port
	return nil
}

func (c AppConfig) String() string {
	return fmt.Sprintf(c.BaseUrl)

}
func (c AppConfig) Set(flagValue string) error {
	if flagValue == "" {
		flagValue = "http://localhost:8080/"
	}
	if !strings.HasSuffix(flagValue, "/") {
		flagValue += "/"
	}
	c.BaseUrl = flagValue
	return nil
}
