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

	"github.com/xfrr/goffmpeg/transcoder"
	"gopkg.in/ini.v1"
)

var output_path string
var output_url string

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

	// init output url from .ini
	output_url = cfg.Section("default").Key("output_url").String()
	if output_url == "" {
		throwerr("No output url defined: %v", err)
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
							outf_name := fmt.Sprintf("%x", string(outf_md5[:])) + ".mp3"
							outf_fname := output_path + outf_name
							fmt.Printf("outf is: %s\n", outf_name) //DEBUG

							// check if file already exists in our output_path
							// TODO: we should still rewrite the feed, even if the file already exists
							if !fileExists(outf_name) {
								tmp_download, err := ioutil.TempFile("", "pm4a2mp3-download-")
								if err != nil {
									throwerr("error creating tmpfile 3: %v", err)
								}
								//DEBUG defer os.Remove(tmp_download.Name())

								//download to tmpfile and
								err = DownloadFile(tmp_download.Name(), media_url)
								if err != nil {
									complain("error downloading media: %v", err)
								} else {
									//DEBUG defer os.Remove(tmpfile.Name()) //cleanup tmpfile
									// successful download, let's convert to mp3
									// ffmpeg -i bits-2021-02-21-3mslDxiZ.m4a -c:a libmp3lame -q:a 4 bits.mp3

									fmt.Println("Download done, starting transcoding") //DEBUG
									trans := new(transcoder.Transcoder)
									err := trans.Initialize(tmp_download.Name(), outf_fname)
									trans.MediaFile().SetSkipVideo(true)

									if err != nil {
										complain("error transcoding: %v", err)
									}
									done := trans.Run(false)
									err = <-done

									if err != nil {
										complain("error transcoding 2: %v", err)
										os.Remove(outf_fname) // cleanup broken output file
									} else {
										// TODO replace enclosure in output feed
										line = fmt.Sprintf("<enclosure url=\"%s%s\" type=\"audio/mp3\"/>", output_url, outf_name)
									}
								}

							} else { //DEBUG
								fmt.Println("outf already exists") // DEBUG
							} //DEBUG
						}

						// write line into feed tmp outfile
						tmp_outfile.WriteString(line + "\n")
					}
					// close tmp outfile
					tmp_outfile.Close()

					// save updated rss
					copyFile(tmp_outfile.Name(), output_path+"feed.rss")
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

// Thanks: https://golangcode.com/check-if-a-file-exists/
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Thanks: https://opensource.com/article/18/6/copying-files-go
func copyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
