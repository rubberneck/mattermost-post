package main

import (
	"encoding/json"
	"fmt"
	"github.com/crgimenes/goconfig"
	_ "github.com/crgimenes/goconfig/json"
	"github.com/mattermost/mattermost-server/model"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type configMMPost struct {
	Server   string `json:"server"`
	Pat      string `json:"pat"`
	Team     string `json:"team"`
	Channel  string `json:"channel"`
	MaxLines int    `json:"maxlines" cfgDefault:"50"`
	Syntax   string `json:"syntax"`
	Filename string `json:"filename"`
}

var client *model.Client4
var channel *model.Channel

func main() {

	config := configMMPost{}

	if _, err := os.Stat(os.Getenv("HOME") + "/.config/mmpost"); os.IsNotExist(err) {
		fmt.Println()
		_ = os.MkdirAll(os.Getenv("HOME")+"/.config/mmpost", 0755)
		configFileJson, _ := json.MarshalIndent(config, "", "\t")
		_ = ioutil.WriteFile(os.Getenv("HOME")+"/.config/mmpost/config.json", configFileJson, 0644)
		fmt.Println("Created config File " + os.Getenv("HOME") + "/.config/mmpost/config.json")
		os.Exit(0)
	}

	goconfig.Path = os.Getenv("HOME") + "/.config/mmpost/"
	goconfig.File = "config.json"
	goconfig.FileRequired = true
	goconfig.PrefixEnv = "MMPOST"
	err := goconfig.Parse(&config)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pipeInfo, _ := os.Stdin.Stat()
	if (pipeInfo.Mode() & os.ModeCharDevice) == os.ModeCharDevice {
		fmt.Println("The command is intended to work with pipes.")
		fmt.Println("Usage:")
		fmt.Println("  cat yourfile.txt | mmpost")
		os.Exit(1)
	}

	stdinReader := io.Reader(os.Stdin)
	messageBytes, err := ioutil.ReadAll(stdinReader)

	if err != nil {
		fmt.Println("ERROR: Reading all from stdin")
		fmt.Println(err)
		os.Exit(1)
	}

	messageString := string(messageBytes)

	if cap(messageBytes) <= 0 {
		fmt.Println("ERROR: pipe was empty")
		os.Exit(1)
	}

	client = model.NewAPIv4Client(config.Server)
	client.AuthType = "BEARER"
	client.AuthToken = config.Pat

	if rchannel, resp := client.GetChannelByNameForTeamName(config.Channel, config.Team, ""); resp.Error != nil {
		fmt.Println("We failed to get the channel")
		fmt.Println(resp.Error)
		os.Exit(1)
	} else {
		channel = rchannel
	}

	if strings.Count(messageString, "\n") <= config.MaxLines && cap(messageBytes) <= 16300 {
		post := &model.Post{}
		post.ChannelId = channel.Id
		post.Message = "```" + config.Syntax + "\n" + messageString + "```\n"

		if _, resp := client.CreatePost(post); resp.Error != nil {
			fmt.Println("We failed to send a message to the channel")
			fmt.Println(resp.Error)
			os.Exit(1)
		}
	} else {
		if config.Filename != "" {
			fileUploadResponse, response := client.UploadFile(messageBytes, channel.Id, config.Filename)
			if response.Error != nil {
				fmt.Println("ERROR: Failed to upload file.")
				fmt.Println(response.Error)
				os.Exit(1)
			}

			post := &model.Post{}
			post.ChannelId = channel.Id
			post.FileIds = []string{fileUploadResponse.FileInfos[0].Id}
			if _, resp := client.CreatePost(post); resp.Error != nil {
				fmt.Println("We failed to send a message to the channel")
				fmt.Println(resp.Error)
				os.Exit(1)
			}
		} else {
			fmt.Println("ERROR: text longer than " + fmt.Sprint(config.MaxLines) + " lines or larger than 16300 bytes requires --filename")
			os.Exit(1)
		}
	}
}
