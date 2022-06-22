package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PageChildren struct {
	Child []struct {
		Object    string
		Id        string
		Paragraph struct {
			Rich_text []struct {
				Text struct {
					Content string
				}
			}
		}
		Image struct {
			Type     string
			External struct {
				Url string
			}
		}
	} `json:"results"`
}

func httpRequestWithBody(config *Config, requestType string, url string, content string) string {
	client := &http.Client{}
	req, _ := http.NewRequest(requestType, url, bytes.NewBuffer([]byte(content)))
	req.Header = http.Header{
		"Authorization":  {"Bearer " + config.notion_secret},
		"Content-Type":   {"application/json"},
		"Notion-Version": {"2022-06-28"},
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
	return string(body)
}

func getPageChildren(config *Config) PageChildren {
	url := "https://api.notion.com/v1/blocks/" + config.notion_page_id + "/children"
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header = http.Header{
		"Authorization":  {"Bearer " + config.notion_secret},
		"Notion-Version": {"2022-06-28"},
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

	var pageChildren PageChildren
	json.Unmarshal([]byte(string(body)), &pageChildren)
	return pageChildren
}

func deleteBlock(config *Config, id string) {
	url := "https://api.notion.com/v1/blocks/" + id
	httpRequestWithBody(config, http.MethodDelete, url, "")
}

func updateChilds(config *Config, pC *PageChildren, streams *StreamInfoAnswer, users *UserInfoAnswer) {
	for i, stream := range streams.Stream {

		// text
		var formatStreams string
		formatStreams = formatStreams + fmt.Sprintf("%s streams: %s\\n", stream.User_name, stream.Title)
		formatStreams = formatStreams + fmt.Sprintf("%d live viewers\\n", stream.Viewer_count)
		content := `{
			"paragraph": {
	
				"rich_text": [{ 
			
				  "text": { "content": "` + formatStreams + `" } 
			
				  }]
			  }
		}`
		block_id := pC.Child[i*2].Id
		url := "https://api.notion.com/v1/blocks/" + block_id
		httpRequestWithBody(config, http.MethodPatch, url, content)

		// image
		image := strings.Replace(strings.Replace(stream.Thumbnail_url, "{width}", "800", -1), "{height}", "450", -1)
		image = image + "?" + strconv.FormatInt(time.Now().Unix(), 10)
		content = `{
			"image": { 
		
				"external": { "url": "` + image + `" } 
		
				}
		}`
		block_id = pC.Child[(i*2)+1].Id
		url = "https://api.notion.com/v1/blocks/" + block_id
		httpRequestWithBody(config, http.MethodPatch, url, content)
		fmt.Println("Updated notion page")
	}
}

func Notion(config *Config, streams *StreamInfoAnswer, users *UserInfoAnswer) {
	pageChildren := getPageChildren(config)
	numberChildren := len(pageChildren.Child)
	// delete unnecessary blocks
	if numberChildren > 2*len(streams.Stream) {
		for i := 1; ; i++ {
			if numberChildren <= 2*len(streams.Stream) {
				break
			}
			deleteBlock(config, pageChildren.Child[len(pageChildren.Child)-i].Id)
			numberChildren -= 1
		}
	}
	// add needed blocks
	for numberChildren < 2*len(streams.Stream) {
		url := "https://api.notion.com/v1/blocks/" + config.notion_page_id + "/children"
		err := httpRequestWithBody(config, http.MethodPatch, url, `{
			"children": [
				{
					"object":	"block",
					"type":		"paragraph",
					"paragraph": {
						"rich_text": []
					}
				},
				{
					"object":	"block",
					"type":		"image",
					"image": {
						"external": {
							"url": "https://static.twitchcdn.net/assets/favicon-32-e29e246c157142c94346.png"
						}
					}
				}
			]}
		`)
		fmt.Println(err)
		numberChildren += 2
	}
	pageChildren = getPageChildren(config)
	updateChilds(config, &pageChildren, streams, users)
}
