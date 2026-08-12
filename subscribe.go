package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/hikkanon512/golib26/asylum"
)

type threadIdentifier struct {
	id    uint
	board string
}

const subscribeList = "./subscribe.txt"

const (
	routineAlreadyExist = iota + 1
	routineAccepted
	routineIsNew
)

func SubscribeModeEnable(directory string) {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := subscribeModeEnable(directory); err != nil {
			log.Fatalln(err)
		}
	}()

	<-exit
}

func subscribeModeEnable(directory string) error {
	currentState := make(map[threadIdentifier]byte)
	for {
		fileData, err := os.ReadFile(subscribeList)
		if err != nil {
			return err
		}
		lines := strings.Split(string(fileData), "\n")
		for _, line := range lines {
			fields := strings.Split(strings.TrimSpace(line), " ")
			if len(fields) < 2 {
				continue
			}

			var ident threadIdentifier
			ident.board = fields[1]
			id, parseErr := strconv.ParseUint(fields[0], 10, 64)
			if parseErr != nil {
				ident.board = fields[0]
				id, parseErr = strconv.ParseUint(fields[1], 10, 64)
				if parseErr != nil {
					continue
				}
			}
			ident.id = uint(id)

			if _, ok := currentState[ident]; ok {
				currentState[ident] = routineAccepted
				continue
			}
			currentState[ident] = routineIsNew
		}
		for ident, state := range currentState {
			if state == routineAlreadyExist {
				delete(currentState, ident)
			}
		}
		for ident, state := range currentState {
			if state == routineAccepted {
				currentState[ident] = routineAlreadyExist
			}
		}

		stateUpdated := false
		for ident, state := range currentState {
			if state != routineIsNew {
				continue
			}
			go subscribeThread(ident.board, ident.id, directory, currentState)
			currentState[ident] = routineAlreadyExist
			stateUpdated = true
		}
		if stateUpdated {
			log.Printf("new routines was created\n")
		}
		time.Sleep(time.Minute)
	}
}

func subscribeThread(section string, threadID uint, directory string, identsMap map[threadIdentifier]byte) {
	for {
		if _, ok := identsMap[threadIdentifier{
			board: section,
			id:    threadID,
		}]; !ok {
			log.Printf("thread /%s/%d/ was deleted from update list\n", section, threadID)
			return
		}

		if err := asylum.DownloadThread(section, threadID, directory); err != nil && err != asylum.ErrThreadNotFound {
			log.Printf("ERROR: on thread /%s/%d/ has error %s\n", section, threadID, err)
			return
		} else if err == asylum.ErrThreadNotFound {
			return
		}
		log.Printf("thread /%s/%d/ was updated\n", section, threadID)
		time.Sleep(updatePause)
	}
}
