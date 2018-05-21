package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/floodcode/tgbot"
)

const (
	configPath         = "config.json"
	foresightsPath     = "foresights"
	userForesightsPath = "user-foresights.json"
)

var (
	bot                 tgbot.TelegramBot
	botUser             tgbot.User
	botConfig           BotConfig
	foresights          = []string{}
	userForesights      = UserForesights{UserMapping: UserMapping{}}
	userForesightsMutex = &sync.Mutex{}
)

type UserForesights struct {
	LastDay     int         `json:"last_day"`
	UserMapping UserMapping `json:"user_mapping"`
}

type BotConfig struct {
	Token string `json:"token"`
}

type UserMapping map[int]int

func (m UserMapping) MarshalJSON() ([]byte, error) {
	resultMap := map[string]string{}
	for key, value := range m {
		stringKey := strconv.Itoa(key)
		stringValue := strconv.Itoa(value)
		resultMap[stringKey] = stringValue
	}

	return json.Marshal(resultMap)
}

func (m UserMapping) UnmarshalJSON(b []byte) error {
	mapping := map[string]string{}
	err := json.Unmarshal(b, &mapping)
	if err != nil {
		return err
	}

	for key, value := range mapping {
		numericKey, err := strconv.Atoi(key)
		if err != nil {
			return err
		}

		numericValue, err := strconv.Atoi(value)
		if err != nil {
			return err
		}

		m[numericKey] = numericValue
	}

	return nil
}

func main() {
	// rand is used to renerate random foresights
	rand.Seed(time.Now().Unix())

	loadForesights()
	loadConfig()
	loadData()
	saveData()
	startBot()
}

func startBot() {
	var err error
	bot, err = tgbot.New(botConfig.Token)
	checkError(err)

	botUser, err = bot.GetMe()
	checkError(err)

	err = bot.Poll(tgbot.PollConfig{
		Delay:    250,
		Callback: updatesCallback,
	})

	checkError(err)
}

func loadForesights() {
	content, err := ioutil.ReadFile(foresightsPath)
	checkError(err)

	foresights = strings.Split(string(content), "\n")
}

func loadConfig() {
	configData, err := ioutil.ReadFile(configPath)
	checkError(err)

	err = json.Unmarshal(configData, &botConfig)
	checkError(err)
}

func loadData() {
	if _, err := os.Stat(userForesightsPath); os.IsNotExist(err) {
		saveData()
	}

	configData, err := ioutil.ReadFile(userForesightsPath)
	checkError(err)

	err = json.Unmarshal(configData, &userForesights)
	checkError(err)
}

func saveData() {
	userForesightsMutex.Lock()
	jsonData, _ := json.Marshal(userForesights)
	err := ioutil.WriteFile(userForesightsPath, jsonData, 0644)
	logError(err)
	userForesightsMutex.Unlock()
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

	_, err := bot.SendMessage(tgbot.SendMessageConfig{
		ChatID:    tgbot.ChatID(message.Chat.ID),
		Text:      foresightMessage,
		ParseMode: tgbot.ParseModeMarkdown(),
	})

	logError(err)
}

func getForesight(user *tgbot.User) string {
	currentDay := time.Now().Day()
	if currentDay != userForesights.LastDay {
		userForesights.UserMapping = map[int]int{}
		userForesights.LastDay = currentDay
		saveData()
	}

	foresightIndex, ok := userForesights.UserMapping[user.ID]
	if ok {
		return foresights[foresightIndex]
	}

	randomIndex := rand.Intn(len(foresights))
	userForesights.UserMapping[user.ID] = randomIndex
	saveData()

	return foresights[randomIndex]
}

func logError(e error) {
	if e != nil {
		log.Println(e)
	}
}

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}
