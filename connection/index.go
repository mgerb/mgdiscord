package connection

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/mgerb/mgdiscord/config"
)

// keep our own connection state for each guild
// guild id mapped to connection object
var connections = map[string]*Connection{}

// Start -
func Start(token string) {

	dg, err := discordgo.New("Bot " + token)

	if err != nil {
		log.Fatal(err)
	}

	// do connection setup here
	dg.AddHandler(mainHandler)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

func mainHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	if connections[m.GuildID] == nil {
		connections[m.GuildID] = &Connection{}
	}

	conn := connections[m.GuildID]

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// TODO:
	if strings.HasPrefix(m.Content, config.Config.BotPrefix) {

		content := strings.TrimPrefix(m.Content, config.Config.BotPrefix)

		var err error

		if strings.HasPrefix(content, "play") {
			conn.joinUsersChannel(s, m)
			args := strings.Split(strings.Trim(strings.TrimPrefix(content, "play"), " \n"), " ")
			err = conn.playAudio(args)
		}

		if err != nil {
			log.Println(err)
			sendMessage(s, m, err.Error())
		}
	}
}

func sendMessage(s *discordgo.Session, m *discordgo.MessageCreate, content string) {
	s.ChannelMessageSend(m.ChannelID, "```\n"+content+"\n```")
}
