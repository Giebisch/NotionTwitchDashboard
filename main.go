package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

type Config struct {
	client_id       string
	client_secret   string
	twitch_channels string
	notion_secret   string
}

func initialize() (*Config, error) {
	file, err := os.Open("config.txt")
	// no configuration file exists
	if errors.Is(err, os.ErrNotExist) {
		file, err := os.Create("config.txt")
		if err != nil {
			panic(err)
		}
		defer file.Close()

		_, _ = file.WriteString("CLIENT_ID=\n")
		_, _ = file.WriteString("CLIENT_SECRET=\n")
		_, _ = file.WriteString("TWITCH_CHANNELS=\n")
		_, _ = file.WriteString("NOTION_SECRET=\n")

		fmt.Println("Please add credentials in 'config.txt' first.")
		return nil, errors.New("no credentials")
	}
	defer file.Close()

	// config file exists: read variables
	var env_vars []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		env_vars = append(env_vars, strings.Split(scanner.Text(), "=")[1])
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	config := Config{
		client_id:       env_vars[0],
		client_secret:   env_vars[1],
		twitch_channels: env_vars[2],
		notion_secret:   env_vars[3],
	}
	return &config, nil
}

func main() {
	config, err := initialize()
	if err != nil {
		panic(err)
	}

	var wg = &sync.WaitGroup{}
	wg.Add(1)
	go Twitch(config, wg)
	wg.Wait()
}
