package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
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
	// append slash if necessary
	if !strings.HasSuffix(output_path, "/") {
		output_path += "/"
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

				// prepare tmpfile for feed download
				tmpfile, err := ioutil.TempFile("", "pm4a2mp3-")
				if err != nil {
					throwerr("error creating tmpfile: %v", err)
				}
				//TODO  defer os.Remove(tmpfile.Name()) // clean up later
				fmt.Printf("we have a tmpfile: %s\n", tmpfile.Name()) //DEBUG

				// download feed
				media_url := ""
				err = DownloadFile(tmpfile.Name(), url)
				if err != nil {
					complain("error downloading feed: %v", err)
				} else {
					// prepare tmpfile for rewritten feed
					tmp_outfile, err := ioutil.TempFile("", "pm4a2mp3-new-")
					if err != nil {
						throwerr("error creating tmpfile 2: %v", err)
					}
					//TODO  defer os.Remove(tmp_outfile.Name()) // clean up later

					// compile regexp (TODO: not the best place)
					r, _ := regexp.Compile("<enclosure[^>]+url=\"([^\"]+m4a)\"")

					// parse Embedded Media from feed and rewrite
					s := bufio.NewScanner(tmpfile)
					for s.Scan() {
						line := s.Text()
						//fmt.Println(line) //DEBUG
						matchs := r.FindStringSubmatch(line)
						if len(matchs) > 1 {
							fmt.Println(matchs[1]) //DEBUG
							media_url = matchs[1]

							// TODO BROKEN
							outf_md5 := md5.Sum([]byte(media_url))
							fmt.Printf("md5 is: %x\n", outf_md5) //DEBUG
							outf_name := output_path + fmt.Sprintf("%x", string(outf_md5[:])) + ".mp3"
							fmt.Printf("outf is: %s\n", outf_name) //DEBUG

							// TODO: check if file already exists in our output_path
							// TODO: download to tmpfile and...
							// TODO: convert m4a to mp3	(in output_path)
							// TODO: save updated rss
						}

						// write line into feed tmp outfile
						tmp_outfile.WriteString(line + "\n")
					}
					// close tmp outfile
					tmp_outfile.Close()
				}

			}
		}
	}

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
