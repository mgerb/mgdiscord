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
	pausedItem          chan *audioItem
	playAudioInProgress bool
	skip                chan bool
	pause               chan bool
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

	if len(args) > 1 {
		timestamp, _ = util.ParseTimeStamp(args[1])
	} else {
		timestamp, _ = util.ParseTimeStampFromURL(args[0])
	}

	item := &audioItem{
		url:      args[0],
		opusData: make(chan []byte, 1000),
		dead:     false,
	}

	c.audioQueue <- item

	filePath, err := util.DownloadFromLink(args[0], config.Config.Timeout)

	if err != nil {
		return err
	}

	go util.WriteOpusData(filePath, 2, 960, 48000, timestamp, item)

	// wait for channel to at least have some audio data in it before start playing
	for {
		if len(item.opusData) > 0 {
			go c.playAudioInQueue()
			break
		}

		time.Sleep(time.Millisecond * 50)
	}

	return nil
}

// pulls audio items off audio channel until empty or paused
func (c *Connection) playAudioInQueue() {

	if c.playAudioInProgress {
		return
	}

	c.paused = false
	c.playAudioInProgress = true
	c.vc.Speaking(true)

outerloop:
	for len(c.audioQueue) > 0 || len(c.pausedItem) > 0 {
		var item *audioItem
		if len(c.pausedItem) > 0 {
			item = <-c.pausedItem
		} else {
			item = <-c.audioQueue
		}

	forloop:
		for len(item.opusData) > 0 {
			select {
			case c.vc.OpusSend <- <-item.opusData:
				break
			case <-c.skip:
				item.Cleanup()
				break forloop
			case <-c.pause:
				c.paused = true
				c.pausedItem <- item
				break outerloop
			}
		}
		time.Sleep(time.Millisecond * 100)
	}

	c.vc.Speaking(false)
	c.playAudioInProgress = false
}
