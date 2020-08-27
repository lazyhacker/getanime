// sorttorrent looks at the files in a directory and moves them into their own
// subdirectories based on the series name.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	parsetorrentname "github.com/lazyhacker/go-parse-torrent-name"
)

var (
	dir = flag.String("dir", ".", "directory with torrent files")
)

func main() {
	flag.Parse()

	f, err := os.Open(*dir)
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
			sdir := filepath.Join(*dir, tinfo.Title)
			err = os.Mkdir(sdir, 0755)
			if err != nil {
				log.Printf("Error message when creating directory. %v", err)
			}
			oldpath := filepath.Join(*dir, file.Name())
			newpath := filepath.Join(sdir, convertName(tinfo))

			if _, err := os.Stat(newpath); err == nil {
				log.Printf("%v already exists! Skip moving.\n", newpath)
				continue
			}

			log.Printf("Moving file from %v to %v.\n", oldpath, newpath)
			err = os.Rename(oldpath, newpath)
			if err != nil {
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
	return fmt.Sprintf("%d - %v [%v] [%v].%v", n.Episode, n.Title, n.Website, n.Resolution, n.Container)

}
