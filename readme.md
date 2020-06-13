# MGDiscord
https://cloud.docker.com/repository/docker/mgerb/mgdiscord

The discord bot that plays everything. This bot is similar to other music bots, but it uses youtube-dl
to play audio from a wide variety of sources. [See the list of supported sites here](https://ytdl-org.github.io/youtube-dl/supportedsites.html).

## Features

- queue up audio from a wide variety of sites
- start audio track from specific timestamp (see Commands)
- support for timestamp in Youtube URL
- skip/pause/resume tracks
- set volume

## Commands

- play \<url\> \<optional timestamp\> | \<youtube search query\> - add a valid URL or youtube search to the play queue
  - timestamp - start an audio track at a specific timestamp
  - e.g. !play https://www.youtube.com/watch?v=oHg5SJYRHA0 1m13s
  - NOTE: a Youtube URL with timestamp in the query string will start playing at that time
- skip - skip to the next in queue
- pause - pause current playing track
- resume - resume playing current track
- volume \<1 - 100\> - set the volume 1 through 100


## Run with Docker Compose

I highly suggest running this with Docker as I am not currently compiling binaries.
This bot depends on youtube-dl, which must be provided by the host. This was previously
provided by the container, but youtube-dl needs to be updated too much for this now.

```
version: "3"

services:
  mgdiscord:
    image: mgerb/mgdiscord:latest
    volumes:
      - /usr/local/bin/youtube-dl:/usr/local/bin/youtube-dl:ro
    env_file:
      - .env
    environment:
      - BOT_PREFIX=${BOT_PREFIX}
      - TOKEN=${TOKEN}
      - TIMEOUT=${TIMEOUT}
```

### Config

- BOT_PREFIX - the prefix to your bot commands e.g. "!play rick roll"
- TOKEN - [How to get your bot token](https://github.com/reactiflux/discord-irc/wiki/Creating-a-discord-bot-&-getting-a-token)
- TIMEOUT - how long before the bot times out downloading a media file