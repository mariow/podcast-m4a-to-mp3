# podcast-m4a-to-mp3

## Why, oh, why?

Podcasts, I spent way too much time on them. And sometimes I'm happy to skip parts of them. Luckily chapter support has made it's way into most podcast players. Unfortunately I find that there are several podcasts I like a lot but that only publish M4A files and it seems that M4A chapters are not supported on Android. To me it seems more than reasonable to set up a script on on old dedicated server to re-encode all of those podcasts to MP3 so that my player correctly displays the chapters. If you think this is ridiculous your are probably right, please move along, nothing to see here.

## HOWTO

### Prerequisites
1. You need to be unreasonably upset about podcasts that are only available in m4a and have no chapter support on Android
2. You cannot be offended by the horrible naming of this script
3. You need some webspace to re-host those podcasts
4. Golang
5. FFMpeg with mp3lame

### Installation

1. `go fetch`
2. `go build podcast-m4a-to-mp3.go`
3. copy the sample .ini to podcast-m4a-to-mp3.ini in the same directory that the script will be run from
4. add the subscriptions you need 
5. set up a cronjob to run
6. Subscribe to the new feed

### .ini documentation and example

```ini
[default]
output_path = "/path/for/output/files/"
output_url = "https://url.that.points/to-the/files-above"


[podcast dummy]
url="https://original.podcasturl/feed.rss"
```
This would:
- fetch the feed from `original.podcasturl/feed.rss`
- re-encode all of the files and save the output in `/path/for/output/files/`
- skip all files that already exist in the output folder
- produce a copy of the rss feed in `/path/for/output/files/` with the name `dummy.rss` (the name of the .ini section)
- assume that all of the file in `/path/for/output/files/` are reachable via `https://url.that.points/to-the/files-above...`

## TODO

- [x] subscribe to feed
- [x] parse all enclosure urls
- [x] build database of urls
- [x] convert everything m4a to mp3
- [x] edit feed to point to new mp3 files
- ~~[ ] make feed name configurable and produce one feed per podcast~~
- [ ] Lockfile
- [ ] remove outdated mp3 files
- [ ] Checks. Safeguards. Seatbelts. 

## DISCLAIMER

It's not beautful. The name is horrible. The code ... well. It doesn't have enough checks against destroying things. But it servers my purpose. Use with care. I'm happy to accept pull requests for improvements though, it should be easy enough to improve this.
