// getanime reads the RSS feed for torrent files and download them to a
// directory.
package main

import (
	"encoding/xml"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Items   Items    `xml:"channel"`
}
type Items struct {
	XMLName  xml.Name `xml:"channel"`
	ItemList []Item   `xml:"item"`
}
type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Guid        string `xml:"guid"`
}

var (
	torrentdir = flag.String("savedir", "./", "Path to save torrent files.")
	rssurl     = flag.String("rss", "", "RSS URL for latest torrents.")
)

func main() {

	flag.Parse()
	if *rssurl == "" {
		flag.Usage()
		os.Exit(1)
	}

	if _, err := os.Stat(*torrentdir); os.IsNotExist(err) {
		log.Fatalf("Directory to save doesn't exist. %v", err)
	}

	resp, err := http.Get(*rssurl)

	if err != nil {
		log.Fatalf("Unable to get feed. %v", err)
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)

	var i RSS

	if err != nil {
		log.Fatal(err)
	}
	err = xml.Unmarshal(content, &i)
	if err != nil {
		log.Fatal(err)
	}

	for c, item := range i.Items.ItemList {
		log.Printf("%d Title: %s\n", c, item.Title)
		log.Printf("%d Link: %s\n", c, item.Link)

		download, err := http.Get(item.Link)
		defer download.Body.Close()
		if err != nil {
			log.Printf("%d: Unable to download %s\n", c, item.Title)
			break
		}

		torrent, err := ioutil.ReadAll(download.Body)

		if err != nil {
			log.Printf("%d: Error parsing body of %s\n", c, item.Title)
			break
		}

		downloadpath := *torrentdir + "/" + item.Description + ".torrent"
		log.Printf("%d Saving: %s\n", c, downloadpath)

		if _, err = os.Stat(downloadpath); os.IsNotExist(err) {
			err = ioutil.WriteFile(downloadpath, torrent, 0644)
			if err != nil {
				log.Printf("Unable to save %s\n", downloadpath)
			}
		} else {
			log.Printf("%d: %s already exists.\n", c, downloadpath)
		}
	}
}
