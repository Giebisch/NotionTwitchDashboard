package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

type AccessToken struct {
	Access_token string
	Expires_in   int
	Token_type   string
}

type UserInfoAnswer struct {
	Data []struct {
		Id                string
		Login             string
		Display_name      string
		Offline_image_url string
	} `json:"data"`
}

type StreamInfoAnswer struct {
	Stream []struct {
		Game_id       string
		Id            string
		Title         string
		Thumbnail_url string
		User_id       string
		User_name     string
		Viewer_count  int
	} `json:"data"`
}

func getAccessToken(config *Config) AccessToken {
	url := "https://id.twitch.tv/oauth2/token"

	postBody, _ := json.Marshal(map[string]string{
		"client_id":     config.client_id,
		"client_secret": config.client_secret,
		"grant_type":    "client_credentials",
	})
	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post(url, "application/json", responseBody)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var accessToken AccessToken
	json.Unmarshal([]byte(string(body)), &accessToken)
	return accessToken
}

func getUserInfo(config *Config, aT *AccessToken) UserInfoAnswer {
	params := strings.Replace(config.twitch_channels, ",", "&login=", -1)
	url := "https://api.twitch.tv/helix/users?login=" + params
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header = http.Header{
		"Authorization": {"Bearer " + aT.Access_token},
		"Client-Id":     {config.client_id},
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var users UserInfoAnswer
	json.Unmarshal([]byte(string(body)), &users)
	return users
}

func getStreamInfo(config *Config, aT *AccessToken, users *UserInfoAnswer) StreamInfoAnswer {
	var user_ids []string
	for _, user := range users.Data {
		user_ids = append(user_ids, user.Id)
	}
	params := strings.Join(user_ids, "&user_id=")
	url := "https://api.twitch.tv/helix/streams?user_id=" + params
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header = http.Header{
		"Authorization": {"Bearer " + aT.Access_token},
		"Client-Id":     {config.client_id},
	}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var streamInfos StreamInfoAnswer
	json.Unmarshal([]byte(string(body)), &streamInfos)
	return streamInfos
}

func Twitch(config *Config, wg *sync.WaitGroup) {
	defer wg.Done()
	accessToken := getAccessToken(config)
	for {
		users := getUserInfo(config, &accessToken)
		streams := getStreamInfo(config, &accessToken, &users)

		Notion(config, &streams, &users)
		time.Sleep(time.Second * 45)
	}
}
