package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"gopkg.in/ini.v1"
)

var output_path string

func main() {
	cfg, err := ini.Load("podcast-m4a-to-mp3.ini")
	if err != nil {
		throwerr("Fail to read file: %v", err)
	}

	// init output path from .ini
	output_path = cfg.Section("default").Key("output_path").String()
	if output_path == "" {
		throwerr("No output path defined: %v", err)
	}
	fmt.Printf("Output path defined: %s\n", output_path) //DEBUG
	// TODO: validate output path

	// TODO: iterate over all podcast sections
	secs := cfg.SectionStrings()
	for _, sec := range secs {
		//fmt.Printf("%s\n", sec) //DEBUG

		if strings.HasPrefix(sec, "podcast ") {
			fmt.Printf("Is Podcast: ") // DEBUG

			podcast_name := strings.TrimPrefix(sec, "podcast ")
			fmt.Printf("%s\n", podcast_name) // DEBUG

			url := cfg.Section(sec).Key("url").String()
			if url != "" {
				fmt.Printf("URL is: %s\n", url) // DEBUG

				//TODO: validate URL

				// download to tmpfile
				tmpfile, err := ioutil.TempFile("", "pm4a2mp3-")
				if err != nil {
					throwerr("error creating tmpfile: %v", err)
				}
				//TODO  defer os.Remove(tmpfile.Name()) // clean up later
				fmt.Printf("we have a tmpfile: %s\n", tmpfile.Name()) //DEBUG

				// download feed
				err = DownloadFile(tmpfile.Name(), url)
				if err != nil {
					complain("error downloading feed: %v", err)
				} else {
					// parse feed
				}

			}
		}
	}

	// TODO: parse Embedded Media from feed
	// TODO: skip if the file is not m4a
	// TODO: check if file already exists in our output_path
	// TODO: download to tmpfile and...
	// TODO: convert m4a to mp3	(in output_path)
	// TODO: write updated rss

	//	keys := cfg.Section(sec).KeyStrings()
	//	for _, key := range keys {
	//		fmt.Printf("%s=%s\n", key, cfg.Section(sec).Key(key).String())
}

// Complain about an error, don't fail
func complain(msg string, err error) {
	fmt.Printf(msg+"\n", err)
}

// Throw error and exit
func throwerr(msg string, err error) {
	complain(msg, err)
	os.Exit(1)
}

// Thanks: https://golangcode.com/download-a-file-from-a-url/
func DownloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
