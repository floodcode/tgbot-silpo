package main

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/floodcode/tgbot"
)

var (
	bot        tgbot.TelegramBot
	botUser    tgbot.User
	foresights []string
)

type botConfig struct {
	Token string `json:"token"`
}

func main() {
	rand.Seed(time.Now().Unix())

	content, err := ioutil.ReadFile("foresights")
	checkError(err)

	foresights = strings.Split(string(content), "\n")

	configData, err := ioutil.ReadFile("config.json")
	checkError(err)

	var config botConfig
	err = json.Unmarshal(configData, &config)
	checkError(err)

	bot, err = tgbot.New(config.Token)
	checkError(err)

	botUser, err = bot.GetMe()
	checkError(err)

	bot.Poll(tgbot.PollConfig{
		Delay:    100,
		Callback: updatesCallback,
	})
}

func updatesCallback(updates []tgbot.Update) {
	for _, update := range updates {
		if update.Message == nil || len(update.Message.Text) == 0 {
			continue
		}

		processTextMessage(update.Message)
	}
}

func processTextMessage(message *tgbot.Message) {
	var cmdMatch, _ = regexp.Compile(`^\/([a-zA-Z_]+)(?:@` + botUser.Username + `)?(?:\s(.+)|)$`)
	match := cmdMatch.FindStringSubmatch(message.Text)

	if match == nil {
		return
	}

	command := strings.ToLower(match[1])

	if command == "silpo" {
		sendForesight(message)
	}
}

func sendForesight(message *tgbot.Message) {
	foresight := foresights[rand.Intn(len(foresights))]

	bot.SendMessage(tgbot.SendMessageConfig{
		ChatID:    tgbot.ChatID(message.Chat.ID),
		Text:      foresight,
		ParseMode: tgbot.ParseModeMarkdown(),
	})
}

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}
