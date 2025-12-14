// getanime downloads torrent files from a RSS feed.
package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	parsetorrentname "github.com/lazyhacker/go-parse-torrent-name"
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
	sort       = flag.Bool("sort", false, "whether to sort the torrents.")
)

func main() {

	flag.Parse()
	if *sort {
		sortTorrents(*torrentdir)
		os.Exit(0)
	}

	if _, err := os.Stat(*torrentdir); os.IsNotExist(err) {
		log.Fatalf("Directory to save doesn't exist. %v", err)
	}

	if *rssurl == "" {
		flag.Usage()
		os.Exit(1)
	}

	resp, err := http.Get(*rssurl)

	if err != nil {
		log.Fatalf("Unable to get feed. %v", err)
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		log.Fatal(err)
	}

	var r RSS
	if err = xml.Unmarshal(content, &r); err != nil {
		log.Fatal(err)
	}

	for c, item := range r.Items.ItemList {
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

		downloadpath := filepath.Join(*torrentdir, item.Description+".torrent")
		log.Printf("%d Saving: %s\n", c, downloadpath)

		if _, err = os.Stat(downloadpath); os.IsNotExist(err) {
			if err = ioutil.WriteFile(downloadpath, torrent, 0644); err != nil {
				log.Printf("Unable to save %s\n", downloadpath)
			}
			continue
		}
		log.Printf("%d: %s already exists.\n", c, downloadpath)
	}
}

// sorttorrent the torrent files and move them into their own
// subdirectories based on the series name.
func sortTorrents(dir string) {
	f, err := os.Open(dir)
	if err != nil {
		log.Fatalf("Unable to open dir. %v\n", err)
	}
	defer f.Close()

	files, err := f.Readdir(0)
	if err != nil {
		log.Fatalf("Unable to read the directory. %v\n", err)
	}

	for _, file := range files {

		// Ignore directories and hidden files, only look at MKV files since that's what most fansubbers are using.
		if !file.IsDir() && !strings.HasPrefix(file.Name(), ".") && strings.HasSuffix(file.Name(), ".mkv") {

			// Use parser to look at the file name and extract the title, episode #, etc.
			tinfo, err := parsetorrentname.Parse(file.Name())
			//pretty print parsed torrent info in JSON format
			//j, _ := json.MarshalIndent(tinfo, "", "\t")
			//fmt.Printf("%v\n", string(j))
			if err != nil {
				log.Printf("Unable to parse file name, %v.  %v\n", file.Name(), err)
				continue
			}

			// Use series title for the directory name.
			sdir := filepath.Join(dir, tinfo.Title)
			err = os.Mkdir(sdir, 0755)
			if err != nil {
				log.Printf("Error message when creating directory. %v", err)
			}
			oldpath := filepath.Join(dir, file.Name())
			newpath := filepath.Join(sdir, convertName(tinfo))

			if _, err := os.Stat(newpath); err == nil {
				log.Printf("%v already exists! Skip moving.\n", newpath)
				continue
			}

			log.Printf("Moving file from %v to %v.\n", oldpath, newpath)
			if err = os.Rename(oldpath, newpath); err != nil {
				log.Printf("Error moving file from %v to %v. %v\n", oldpath, newpath, err)
			}
		}
	}
}

// convertName will change the name of the file to have episode number first in
// order to allow proper sorting in case the series episodes are from different
// groups.
func convertName(n *parsetorrentname.TorrentInfo) string {
	//j, _ := json.MarshalIndent(tinfo, "", "\t")
	//fmt.Printf("%v\n", string(j))
	return fmt.Sprintf("%02d - %v [%v] [%v].%v", n.Episode, n.Title, n.Website, n.Resolution, n.Container)

}
