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
	vc                  *discordgo.VoiceConnection
	paused              bool
	audioQueue          chan *audioItem
	playAudioInProgress bool
	skip                chan bool
	pause               chan bool
}

type audioItem struct {
	url      string
	opusData [][]byte
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

func (c *Connection) queueAudio(args []string) error {
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

	item := &audioItem{
		opusData: opus,
		url:      args[0],
	}

	c.audioQueue <- item

	go c.playAudioInQueue()

	return nil
}

func (c *Connection) playAudioInQueue() {

	if c.playAudioInProgress {
		return
	}

	c.playAudioInProgress = true
	c.vc.Speaking(true)

	for len(c.audioQueue) > 0 {
		item := <-c.audioQueue

	forloop:
		for _, o := range item.opusData {
			select {
			case c.vc.OpusSend <- o:
				break
			case <-c.skip:
				break forloop
			}
		}
		time.Sleep(time.Millisecond * 100)
	}

	c.vc.Speaking(false)
	c.playAudioInProgress = false
}
