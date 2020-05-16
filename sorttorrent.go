package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	parsetorrentname "github.com/lazyhacker/go-parse-torrent-name"
)

var (
	dir = flag.String("dir", "", "directory with torrent files")
)

func main() {
	flag.Parse()

	f, err := os.Open(*dir)
	if err != nil {
		log.Fatalf("Unable to open dir. %v\n", err)
	}

	files, err := f.Readdir(0)
	if err != nil {
		log.Fatalf("Unable to read the directory. %v\n", err)
	}

	for _, file := range files {
		if !file.IsDir() && !strings.HasPrefix(file.Name(), ".") && strings.HasSuffix(file.Name(), ".mkv") {
			tinfo, err := parsetorrentname.Parse(file.Name())
			if err != nil {
				log.Printf("Unable to parse file name, %v.  %v\n", file.Name(), err)
				continue
			}

			sdir := filepath.Join(*dir, tinfo.Title)
			err = os.Mkdir(sdir, 0755)
			if err != nil {
				log.Printf("Error message when creating directory. %v", err)
			}
			oldpath := filepath.Join(*dir, file.Name())
			newpath := filepath.Join(sdir, file.Name())

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
