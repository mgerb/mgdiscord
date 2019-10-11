package connection

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
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

	// create new connection for guild if not already exists
	if connections[m.GuildID] == nil {
		connections[m.GuildID] = &Connection{
			paused:              true,
			audioQueue:          make(chan *audioItem, 10),
			pausedItem:          make(chan *audioItem, 1),
			playAudioInProgress: false,
			skip:                make(chan bool, 1),
			pause:               make(chan bool, 1),
			volume:              0.5,
		}
	}

	conn := connections[m.GuildID]

	if conn != nil {
		conn.handleMessage(s, m)
	}
}
