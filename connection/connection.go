package connection

import (
	"errors"
	"log"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mgerb/mgdiscord/config"
	"github.com/mgerb/mgdiscord/util"
)

// Connection -
type Connection struct {
	vc *discordgo.VoiceConnection
}

// join users channel that sent the message
func (c *Connection) joinUsersChannel(s *discordgo.Session, m *discordgo.MessageCreate) error {
	id, err := getVoiceChannelID(s, m)

	if err != nil {
		log.Println(err)
		return err
	}

	// return if voice connection is already in channel
	if c.vc != nil && c.vc.ChannelID == id {
		return nil
	}

	c.vc, err = s.ChannelVoiceJoin(m.GuildID, id, false, false)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (c *Connection) playAudio(args []string) error {

	if c.vc == nil || len(args) == 0 || args[0] == "" {
		return errors.New("Invalid arguments: " + config.Config.BotPrefix + "play <url> <timestamp>")
	}

	var timestamp string
	var err error

	if len(args) > 1 {
		timestamp, err = util.ParseTimeStamp(args[1])
	} else {
		timestamp, err = util.ParseTimeStampFromURL(args[0])
	}

	opus, err := util.GetOpusFromLink(args[0], timestamp)

	if err != nil {
		return err
	}

	c.vc.Speaking(true)

	for _, o := range opus {
		c.vc.OpusSend <- o
	}

	time.Sleep(time.Millisecond * 100)

	c.vc.Speaking(false)

	return nil
}
