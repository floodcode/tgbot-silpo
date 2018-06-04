package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/floodcode/tbf"
	"github.com/floodcode/tgbot"
)

const (
	configPath         = "config.json"
	foresightsPath     = "foresights"
	userForesightsPath = "user-foresights.json"
)

var (
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
	Delay int    `json:"delay"`
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
	loadData()
	saveData()

	configData, err := ioutil.ReadFile(configPath)
	checkError(err)

	var config BotConfig
	err = json.Unmarshal(configData, &config)
	checkError(err)

	bot, err := tbf.New(config.Token)
	checkError(err)

	bot.AddRoute("silpo", silpoAction)

	err = bot.Poll(tbf.PollConfig{
		Delay: config.Delay,
	})
}

func silpoAction(req tbf.Request) {
	foresight := getForesight(req.Message.From)
	req.QuickMessageMD(fmt.Sprintf("_Ваше передбачення на сьогодні:_\n*%s*", foresight))
}

func loadForesights() {
	content, err := ioutil.ReadFile(foresightsPath)
	checkError(err)

	foresights = strings.Split(string(content), "\n")
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
	ioutil.WriteFile(userForesightsPath, jsonData, 0644)
	userForesightsMutex.Unlock()
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

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}
