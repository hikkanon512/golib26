package asylum

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
)

type Message struct {
	Banned    int    `json:"banned"`
	Board     string `json:"board"`
	Closed    uint   `json:"closed"`
	Comment   string `json:"comment"` // message
	Date      string `json:"date"`
	Email     string `json:"email"`
	Endless   uint   `json:"endless"`
	Files     []File `json:"files"`
	LastHit   uint   `json:"lasthit"`
	Name      string `json:"name"`
	Num       uint   `json:"num"`
	Op        uint   `json:"op"`
	Parent    uint   `json:"parent"`
	Sticky    uint   `json:"sticky"`
	Subject   string `json:"subject"` // subject
	Tags      string `json:"tags"`
	Timestamp uint   `json:"timestamp"`
	Trip      string `json:"trip"`
	Views     uint   `json:"views"`
}

var (
	ErrThreadNotFound = errors.New("thread not found")
	ErrNoThreads      = errors.New("no threads in response")
	ErrUnexpected     = errors.New("some error")
)

func GetMessages(section string, threadID uint) ([]Message, error) {
	threadURL := siteURL + `/` + section + `/res/` + strconv.FormatUint(uint64(threadID), 10) + `.json`

	request, err := http.Get(threadURL)
	if err != nil {
		return nil, err
	}
	defer request.Body.Close()

	if request.StatusCode != http.StatusOK {
		if request.StatusCode == http.StatusNotFound {
			return nil, ErrThreadNotFound
		}
		return nil, fmt.Errorf("site returned %d: %w", request.StatusCode, ErrUnexpected)
	}

	type Content struct {
		Messages []Message `json:"posts"`
	}
	var resp struct {
		CurrentThread    uint      `json:"current_thread"`
		FilesCount       uint      `json:"files_count"`
		IsBoard          bool      `json:"is_board"`
		IsClosed         uint      `json:"is_closed"`
		IsIndex          bool      `json:"is_index"`
		MaxNum           uint      `json:"max_num"`
		PostsCount       uint      `json:"posts_count"`
		ThreadFirstImage string    `json:"thread_first_image"`
		Contents         []Content `json:"threads"`
	}
	if err := json.NewDecoder(request.Body).Decode(&resp); err != nil {
		return nil, err
	}
	if len(resp.Contents) == 0 {
		return nil, ErrNoThreads
	}
	return resp.Contents[0].Messages, nil
}

var ErrFatal = errors.New("error is fatal")
var DirNamePrefix = "thread"

func DownloadThread(section string, threadID uint, directory string) error {
	const errEncounterLim = 8

	messages, err := GetMessages(section, threadID)
	if err != nil {
		return err
	}

	directory += fmt.Sprintf(`\%s_%s_%d`, DirNamePrefix, section, threadID)
	if _, err = os.Stat(directory); err != nil && os.IsNotExist(err) {
		if err = os.Mkdir(directory, modeOwningNonExec); err != nil {
			return fmt.Errorf("sorry %w: %s", ErrFatal, err)
		}
	}
	if err = os.Mkdir(directory+`\thumbs`, modeOwningNonExec); err != nil && !os.IsExist(err) {
		return err
	} else if err = os.Mkdir(directory+`\files`, modeOwningNonExec); err != nil && !os.IsExist(err) {
		return err
	}

	threadFile, err := os.OpenFile(directory+`\thread.json`, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, modeOwningNonExec)
	if err != nil {
		return err
	} else if err = json.NewEncoder(threadFile).Encode(&messages); err != nil {
		threadFile.Close()
		return err
	} else if err = threadFile.Close(); err != nil {
		return err
	}

	// Проверка сохраняемых программой данных на предмет соответствия законам РФ может быть осуществлена лишь в ручном режиме.
	// Автор программы НЕ НЕСЁТ ОТВЕТСТВЕННОСТЬ за загружаемые ею из Интернета данные.
	errEncounter := 0
	for _, message := range messages {
		for _, file := range message.Files {
			err = loadFileIfNotExist(siteURL+file.Path, directory+`\files\`+file.Name)
			if thumbErr := loadFileIfNotExist(siteURL+file.Thumbnail, directory+`\thumbs\`+path.Base(file.Thumbnail)); thumbErr != nil && err == nil {
				err = thumbErr
			}
			if err != nil {
				errEncounter++
				// Limit for single thread.
				if errEncounter >= errEncounterLim {
					return err
				}
			}
		}
	}
	return nil
}
