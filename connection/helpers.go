package connection

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

func getVoiceChannelID(s *discordgo.Session, m *discordgo.MessageCreate) (string, error) {

	c, err := s.State.Channel(m.ChannelID)

	if err != nil {
		return "", err
	}

	guild, err := s.State.Guild(c.GuildID)

	if err != nil {
		return "", err
	}

	for _, vc := range guild.VoiceStates {
		if vc.UserID == m.Author.ID {
			return vc.ChannelID, nil
		}
	}

	return "", errors.New("voice channel not found")
}
