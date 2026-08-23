package main

import (
	"flag"
	"log"
	"time"

	"github.com/hikkanon512/golib26/asylum"
)

var (
	dirNamePrefix     string
	saveIntoDir       string
	threadNum         uint
	section           string
	toSubscribe       bool
	updatePause       time.Duration
	withSubscribeList bool
)

func main() {
	flag.StringVar(&section, "b", "b", "dvach board (b, 2d, zog, se etc)")
	flag.UintVar(&threadNum, "t", 0, "dvach thread number")
	flag.StringVar(&saveIntoDir, "d", "./", "to-save directory path")
	flag.StringVar(&asylum.DirNamePrefix, "p", asylum.DirNamePrefix, "thread folder-name prefix")
	flag.BoolVar(&toSubscribe, "s", false, "subscribe for thread updating or just download")
	flag.DurationVar(&updatePause, "u", 5*time.Minute, "pause between updates")
	flag.BoolVar(&withSubscribeList, "l", false, "subscribe mode with list in file './subscribe.txt'")
	flag.Parse()

	if withSubscribeList && threadNum != 0 {
		log.Fatalln("thread number 't' and subscribe mode 'l' are incompatible")
	} else if threadNum == 0 && !withSubscribeList {
		log.Fatalln("flag 't' must be provided necessarily")
	} else if withSubscribeList {
		SubscribeModeEnable(saveIntoDir)
		return
	}

	if toSubscribe {
		log.Printf("thread /%s/%d/ subscribed for started\n", section, threadNum)
		if err := SubscribeThread(section, threadNum, saveIntoDir); err != nil {
			log.Fatalf("fatal error: %s\n", err)
		}
		log.Printf("thread /%s/%d/ was ended\n", section, threadNum)
	} else {
		log.Printf("thread /%s/%d/ downloading started\n", section, threadNum)
		if err := asylum.DownloadThread(section, threadNum, saveIntoDir); err != nil {
			log.Fatalf("fatal error: %s\n", err)
		}
		log.Printf("thread /%s/%d/ was downloaded\n", section, threadNum)
	}
}

func SubscribeThread(section string, threadID uint, directory string) error {
	for {
		if err := asylum.DownloadThread(section, threadID, directory); err != nil && err != asylum.ErrThreadNotFound {
			return err
		} else if err == asylum.ErrThreadNotFound {
			return nil
		}
		log.Printf("thread /%s/%d/ was updated\n", section, threadID)
		time.Sleep(updatePause)
	}
}
