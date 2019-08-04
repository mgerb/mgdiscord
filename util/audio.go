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

// initialize temp file paths
func init() {
	MakeDirIfNotExists(fileFolder)
}

// GetOpusFromLink - use youtube dl to extract opus data from url
//
// - timestamp - format 00:00:00
func GetOpusFromLink(url, timestamp string) ([][]byte, error) {

	urlHash := GetSha1(url)
	fullFileName, err := FindFullFilePath(fileFolder, urlHash)

	if err != nil {
		return nil, err
	}

	// download and extract audio if file doesn't already exist
	if !FileExists(fullFileName) {
		err := ExecuteCommand("youtube-dl", 30, "-o", path.Join(fileFolder, urlHash)+".%(ext)s", url)

		if err != nil {
			cleanupFailedFiles(fileFolder, urlHash)
			return nil, err
		}

		fullFileName, err = FindFullFilePath(fileFolder, urlHash)

		if err != nil {
			return nil, err
		}
	}

	data, err := GetFileOpusData(fullFileName, 2, 960, 48000, timestamp)

	return data, err
}

// GetFileOpusData - uses ffmpeg to convert any audio
// file to opus data ready to send to discord
//
// - channels - 2
//
// - opusFrameSize - 960
//
// - sampleRate - 48000
//
// - timestamp - format 00:00:00 - start at specified time
func GetFileOpusData(filePath string, channels, opusFrameSize, sampleRate int, timestamp string) ([][]byte, error) {

	args := []string{"-i", filePath, "-f", "s16le", "-acodec", "pcm_s16le", "-ar", strconv.Itoa(sampleRate), "-ac", strconv.Itoa(channels)}

	if timestamp != "" {
		args = append(args, "-ss", timestamp)
	}

	args = append(args, "pipe:1")

	cmd := exec.Command("ffmpeg", args...)

	cmdout, err := cmd.StdoutPipe()

	if err != nil {
		return nil, err
	}

	pcmdata := bufio.NewReader(cmdout)

	err = cmd.Start()

	if err != nil {
		return nil, err
	}

	// create encoder to convert audio to opus codec
	opusEncoder, err := opus.NewEncoder(sampleRate, channels, opus.AppVoIP)

	if err != nil {
		return nil, err
	}

	opusOutput := make([][]byte, 0)

	for {
		// read pcm data from ffmpeg stdout
		audiobuf := make([]int16, opusFrameSize*channels)
		err = binary.Read(pcmdata, binary.LittleEndian, &audiobuf)

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return opusOutput, nil
		}

		if err != nil {
			return nil, err
		}

		// convert raw pcm to opus
		opus := make([]byte, 1000)
		n, err := opusEncoder.Encode(audiobuf, opus)

		if err != nil {
			return nil, err
		}

		// append bytes to output
		opusOutput = append(opusOutput, opus[:n])
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
