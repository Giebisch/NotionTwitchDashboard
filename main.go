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
	notion_page_id  string
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

		env_vars := `CLIENT_ID=
CLIENT_SECRET=
TWITCH_CHANNELS=
NOTION_SECRET=
NOTION_PAGE_ID=
`
		file.WriteString(env_vars)
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
		notion_page_id:  env_vars[4],
	}
	return &config, nil
}

func main() {
	config, err := initialize()
	if err != nil {
		return
	}

	var wg = &sync.WaitGroup{}
	wg.Add(1)
	go Twitch(config, wg)
	wg.Wait()
}
