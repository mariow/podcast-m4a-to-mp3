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
	// TODO: validate output path

	// iterate over all podcast sections
	secs := cfg.SectionStrings()
	for _, sec := range secs {
		if strings.HasPrefix(sec, "podcast ") {
			//REMOVEME: podcast_name := strings.TrimPrefix(sec, "podcast ")

			url := cfg.Section(sec).Key("url").String()
			if url != "" {
				//TODO: validate URL

				// prepare tmpfile for feed download
				tmpfile, err := ioutil.TempFile("", "pm4a2mp3-")
				if err != nil {
					throwerr("error creating tmpfile: %v", err)
				}
				defer os.Remove(tmpfile.Name()) // clean up later

				// download feed
				media_url := ""
				err = downloadFile(tmpfile.Name(), url)
				if err != nil {
					complain("error downloading feed: %v", err)
					continue
				}
				// prepare tmpfile for rewritten feed
				tmp_outfile, err := ioutil.TempFile("", "pm4a2mp3-new-")
				if err != nil {
					throwerr("error creating tmpfile 2: %v", err)
				}
				defer os.Remove(tmp_outfile.Name()) // clean up later

				// compile regexp (TODO: not the best place)
				r, _ := regexp.Compile("<enclosure[^>]+url=\"([^\"]+m4a)\"")

				// parse Embedded Media from feed and rewrite
				s := bufio.NewScanner(tmpfile)
				for s.Scan() {
					line := s.Text()

					matchs := r.FindStringSubmatch(line)
					if len(matchs) > 1 {
						media_url = matchs[1]

						// md5 sum of the original url is our cache key and name for the new file
						outf_md5 := md5.Sum([]byte(media_url))
						outf_name := fmt.Sprintf("%x", string(outf_md5[:])) + ".mp3"
						outf_fname := output_path + outf_name

						// check if file already exists in our output_path
						if !fileExists(outf_fname) {
							tmp_download, err := ioutil.TempFile("", "pm4a2mp3-download-")
							if err != nil {
								throwerr("error creating tmpfile 3: %v", err)
							}
							defer os.Remove(tmp_download.Name())

							//download to tmpfile and
							err = downloadFile(tmp_download.Name(), media_url)
							if err != nil {
								complain("error downloading media: %v", err)
							} else {
								defer os.Remove(tmpfile.Name()) //cleanup tmpfile
								// successful download, let's convert to mp3
								// ffmpeg -i bits-2021-02-21-3mslDxiZ.m4a -c:a libmp3lame -q:a 4 bits.mp3

								err := transcodeFile(tmp_download.Name(), outf_name)
								if err != nil {
									complain("error transcoding: %v", err)
									os.Remove(outf_fname) // cleanup broken output file
									outf_name = ""        // empty outf_name will keep us from rewriting the output feed
								}
							}
						}

						// rewrite enclosure
						if outf_name != "" {
							line = fmt.Sprintf("<enclosure url=\"%s%s\" type=\"audio/mp3\"/>", output_url, outf_name)
						}
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

func transcodeFile(input_file string, output_file string) error {
	trans := new(transcoder.Transcoder)
	err := trans.Initialize(input_file, output_file)
	trans.MediaFile().SetSkipVideo(true)

	if err != nil {
		return err
	}
	done := trans.Run(false)
	err = <-done

	return err
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
func downloadFile(filepath string, url string) error {

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
