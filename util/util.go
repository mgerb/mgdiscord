package util

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

// GetSha1 -
func GetSha1(str string) string {
	hasher := sha1.New()
	hasher.Write([]byte(str))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

// FileExists -
func FileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// ExecuteCommand - execute command and log to stdout
//
// - timeout - in seconds
func ExecuteCommand(name string, timeout int, arg ...string) error {
	cmd := exec.Command(name, arg...)

	var stdBuffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdBuffer)

	cmd.Stdout = mw
	cmd.Stderr = mw

	timeoutchan := make(chan bool)
	donechan := make(chan error)

	go func() {
		// Execute the command
		if err := cmd.Run(); err != nil {
			log.Println(err)
			donechan <- errors.New(stdBuffer.String())
		} else {
			donechan <- nil
		}
	}()

	go func() {
		time.Sleep(time.Second * time.Duration(timeout))
		timeoutchan <- true
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	select {
	case <-timeoutchan:
		return errors.New("Timeout:\n" + stdBuffer.String())
	case err := <-donechan:
		return err
	}
}

// FindFullFilePath - search a directory for a file that starts with specified name
func FindFullFilePath(dir, name string) (string, error) {

	files, err := FindMatchingFiles(dir, name)

	if err != nil || len(files) < 1 {
		return "", err
	}

	return files[0], nil
}

// FindMatchingFiles -
func FindMatchingFiles(dir, name string) ([]string, error) {
	output := []string{}

	files, err := ioutil.ReadDir(dir)

	if err != nil {
		return output, err
	}

	for _, f := range files {
		fileName := f.Name()
		if strings.HasPrefix(fileName, name) {
			output = append(output, path.Join(dir, fileName))
		}
	}

	return output, nil
}

// DeleteFile -
func DeleteFile(path string) error {
	return os.Remove(path)
}

// MakeDirIfNotExists -
func MakeDirIfNotExists(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
}

// ParseTimeStampFromURL - parse from youtube url for example
func ParseTimeStampFromURL(urlString string) (string, error) {
	u, err := url.Parse(urlString)

	if err != nil {
		return "", err

	}

	timeQuery := u.Query().Get("start")

	if timeQuery == "" {
		timeQuery = u.Query().Get("t")
	}

	if timeQuery == "" {
		return "", errors.New("unable to parse timestamp from url")
	}

	return ParseTimeStamp(timeQuery)
}

// ParseTimeStamp - parse a youtube formatted timestamp
//
// - format - 1h2m12s
//
// - assume seconds if number is an Int
func ParseTimeStamp(timestamp string) (string, error) {

	_, err := strconv.Atoi(timestamp)

	// if we can parse an int it means the time stamp is in seconds
	// add "s" onto the timestamp so we can parse in the next step
	if err == nil {
		timestamp = timestamp + "s"
	}

	dur, err := time.ParseDuration(timestamp)

	if err != nil {
		return "", err
	}

	zeroTime := time.Date(0, 0, 0, 0, 0, 0, 0, time.Now().Location())

	return zeroTime.Add(dur).Format("15:04:05"), nil
}

// IsURL - check if string is valid url
func IsURL(urlString string) bool {
	_, err := url.ParseRequestURI(urlString)
	return err == nil
}
