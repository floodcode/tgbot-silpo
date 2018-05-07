package main

import (
	"encoding/json"
	"fmt"
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
	foresights = []string{}
	userMap    = map[int]int{}
	lastDay    int
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
	foresightMessage := fmt.Sprintf("_Ваше передбачення на сьогодні:_\n*%s*", getForesight(message.From))

	bot.SendMessage(tgbot.SendMessageConfig{
		ChatID:    tgbot.ChatID(message.Chat.ID),
		Text:      foresightMessage,
		ParseMode: tgbot.ParseModeMarkdown(),
	})
}

func getForesight(user *tgbot.User) string {
	currentDay := time.Now().Day()
	if currentDay != lastDay {
		userMap = map[int]int{}
		lastDay = currentDay
	}

	foresightIndex, ok := userMap[user.ID]
	if ok {
		return foresights[foresightIndex]
	}

	randomIndex := rand.Intn(len(foresights))
	userMap[user.ID] = randomIndex

	return foresights[randomIndex]
}

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}
