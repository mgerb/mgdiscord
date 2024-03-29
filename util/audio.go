package util

import (
	"bufio"
	"encoding/binary"
	"io"
	"os/exec"
	"path"
	"strconv"

	"github.com/hraban/opus"
)

const fileFolder = "youtube-dl-cache"

// OpusWritable -
type OpusWritable interface {
	OpusChan() chan []byte
	IsClosed() bool
}

// initialize temp file paths
func init() {
	MakeDirIfNotExists(fileFolder)
}

// DownloadMedia - use youtube-dl to download file - returns file path
//
// - queryString - youtube-dl search query
//
// - timeout - in seconds
func DownloadMedia(queryString string, timeout int) (string, error) {

	urlHash := GetSha1(queryString)
	fullFileName, err := FindFullFilePath(fileFolder, urlHash)

	if err != nil {
		return "", err
	}

	// download and extract audio if file doesn't already exist
	if !FileExists(fullFileName) {
		err := ExecuteCommand("yt-dlp", timeout, "-x", "--audio-format", "mp3", "-o", path.Join(fileFolder, urlHash)+".%(ext)s", queryString)

		if err != nil {
			cleanupFailedFiles(fileFolder, urlHash)
			return "", err
		}

		fullFileName, err = FindFullFilePath(fileFolder, urlHash)

		if err != nil {
			return "", err
		}
	}

	return fullFileName, nil
}

// WriteOpusData - uses ffmpeg to convert any audio
// file to opus data ready to send to discord
//
// - writes to OpusWritable
//
// - channels - 2
//
// - opusFrameSize - 960
//
// - sampleRate - 48000
//
// - volume - between 0.0 and 1.0
//
// - timestamp - format 00:00:00 - start at specified time
func WriteOpusData(filePath string, channels, opusFrameSize, sampleRate int, timestamp string, volume float64, opusWriteable OpusWritable) error {

	args := []string{"-i", filePath, "-f", "s16le", "-acodec", "pcm_s16le", "-ar", strconv.Itoa(sampleRate), "-ac", strconv.Itoa(channels)}

	if timestamp != "" {
		args = append(args, "-ss", timestamp)
	}

	args = append(args, "-filter:a", "volume="+strconv.FormatFloat(volume, 'f', 2, 64))
	args = append(args, "pipe:1")

	cmd := exec.Command("ffmpeg", args...)

	cmdout, err := cmd.StdoutPipe()

	if err != nil {
		return err
	}

	pcmdata := bufio.NewReader(cmdout)

	err = cmd.Start()

	if err != nil {
		return err
	}

	// create encoder to convert audio to opus codec
	opusEncoder, err := opus.NewEncoder(sampleRate, channels, opus.AppVoIP)

	if err != nil {
		return err
	}

	for {
		// read pcm data from ffmpeg stdout
		audiobuf := make([]int16, opusFrameSize*channels)
		err = binary.Read(pcmdata, binary.LittleEndian, &audiobuf)

		if err == io.EOF || err == io.ErrUnexpectedEOF || opusWriteable.IsClosed() {
			return nil
		}

		if err != nil {
			return err
		}

		// convert raw pcm to opus
		opus := make([]byte, 1000)
		n, err := opusEncoder.Encode(audiobuf, opus)

		if err != nil {
			return err
		}

		opusWriteable.OpusChan() <- opus[:n]
	}
}

// delete files that start with name
func cleanupFailedFiles(dir, urlHash string) error {
	files, err := FindMatchingFiles(dir, urlHash)

	if err != nil {
		return err
	}

	for _, f := range files {
		err := DeleteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}
