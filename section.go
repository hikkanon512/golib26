package main

import (
	"github.com/hikkanon512/golib26/asylum"
	"log"
)

func DumpABoard(section string, directory string) error {
	bod, threads, err := asylum.GetBoardCatalog(section)
	if err != nil {
		return err
	}
	log.Printf("Are saving %d threads\n", len(threads))
	for _, thread := range threads {
		// fmt.Println(thread.Name, thread.Num, thread.Op, thread.Subject)
		if err = asylum.DownloadThread(section, thread.Num, directory); err != nil {
			return err
		}
		log.Printf("THREADDNNO: %d\n", thread.Num)

	}

	return asylum.SaveBoardMeta(*bod, directory+"/bod.json")
}
