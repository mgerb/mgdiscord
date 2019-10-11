package connection

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/mgerb/mgdiscord/config"
	"github.com/mgerb/mgdiscord/util"
)

const (
	channels   = 2
	frameSize  = 960
	sampleRate = 48000
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
	volume              float64
}

func (c *Connection) handleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	if strings.HasPrefix(m.Content, config.Config.BotPrefix) {

		allArgs := strings.Split(strings.TrimPrefix(m.Content, config.Config.BotPrefix), " ")

		if len(allArgs) == 0 {
			return
		}

		command := allArgs[0]
		args := allArgs[1:]
		var err error

		switch command {
		case "skip":
			if c.playAudioInProgress {
				c.skip <- true
			}
			// if paused already - remove item in paused queue
			if c.paused && len(c.pausedItem) > 0 {
				item := <-c.pausedItem
				item.Cleanup()
				c.playAudioInQueue()
			}
			break

		case "pause":
			if !c.paused {
				c.pause <- true
			}
			break

		case "resume":
			if c.paused {
				c.playAudioInQueue()
			}
			break

		case "play":
			err = c.queueAudio(s, m, args)

			break

		case "volume":
			err = c.setVolume(s, m, args)
			break
		}

		if err != nil {
			log.Println(err)
			c.sendMessage(s, m, err.Error())
		}
	}
}

func (c *Connection) setVolume(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {

	errString1 := "Value must be between 0 and 100"

	confirmMessage := func() {
		c.sendMessage(s, m, "Volume: "+strconv.Itoa(int(c.volume*100)))
	}

	if len(args) == 0 || args[0] == "" {
		confirmMessage()
		return nil
	}

	volume, err := strconv.Atoi(args[0])

	if err != nil || volume < 0 || volume > 100 {
		return errors.New(errString1)
	}

	c.volume = float64(volume) / float64(100)

	confirmMessage()

	return nil
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

func (c *Connection) queueAudio(s *discordgo.Session, m *discordgo.MessageCreate, args []string) error {

	if len(args) == 0 || args[0] == "" {
		return errors.New("Invalid arguments: " + config.Config.BotPrefix + "play <url> <timestamp>")
	}

	err := c.joinUsersChannel(s, m)

	if err != nil {
		return err
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

	go util.WriteOpusData(filePath, channels, frameSize, sampleRate, timestamp, c.volume, item)

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

func (c *Connection) sendMessage(s *discordgo.Session, m *discordgo.MessageCreate, content string) {
	s.ChannelMessageSend(m.ChannelID, "```\n"+content+"\n```")
}
