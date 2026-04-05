package main

import (
	"vexasholdem-bot/bot"
	"log"
	"os"
)

func main() {
	// load environment variables
	botToken, ok := os.LookupEnv("BOT_TOKEN")
	if !ok {
		log.Fatal("BOT_TOKEN not found")
	}

	openWeatherToken, ok := os.LookupEnv("OPENWEATHER_TOKEN")
	if !ok {
		log.Fatal("OPENWEATHER_TOKEN not found")
	}

	// start the bot
	bot.BotToken = botToken
	bot.OpenWeatherToken = openWeatherToken
	bot.Run()
}
