package bot

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// store API tokens
var (
	BotToken string
)

func Run() {
	// create a new discord session
	discord, err := discordgo.New("Bot " + BotToken)
	if err != nil {
		log.Fatal("Error: ", err)
	}

	// add message handler
	discord.AddHandler(newMessage)

	// open session
	discord.Open()
	defer discord.Close()

	// run until code is terminated
	fmt.Println("Bot is running...")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}

func newMessage(discord *discordgo.Session, message *discordgo.MessageCreate) {
	// ignore messages from the bot itself
	if message.Author.ID == discord.State.User.ID {
		return
	}

	// respond to messages
	switch {
	case strings.Contains(message.Content, "user"):
		discord.ChannelMessageSend(message.ChannelID, "Use '!user <email>'")
	case strings.Contains(message.Content, "!user"):
		user := getUser(message.Content)
		discord.ChannelMessageSendComplex(message.ChannelID, user)
	case strings.Contains(message.Content, "position"):
		discord.ChannelMessageSend(message.ChannelID, "Use '!position <email>'")
	case strings.Contains(message.Content, "!position"):
		position := getPosition(message.Content)
		discord.ChannelMessageSendComplex(message.ChannelID, position)
	case strings.Contains(message.Content, "competition"):
		competition := getCompetition(message.Content)
		discord.ChannelMessageSendComplex(message.ChannelID, competition)
	case strings.Contains(message.Content, "trade"):
		discord.ChannelMessageSend(message.ChannelID, 
			"Use '!buy <email> <team name> <yes/no> <amount of VEX to spend>'
			or '!sell <email> <team name> <yes/no> <number of shares to sell>'")
	case strings.Contains(message.Content, "!buy"):
		tradeRequest := handleBuy(message.Content)
		discord.ChannelMessageSend(message.ChannelID, tradeRequest)
	case strings.Contains(message.Content, "!sell"):
		sellRequest := handleSell(message.Content)
		discord.ChannelMessageSend(message.ChannelID, sellRequest)
	}
}
