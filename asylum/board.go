package asylum

import (
	"encoding/json"
	"net/http"
)

type Board struct {
	BumpLimit        uint     `json:"bump_limit"`
	Category         string   `json:"category"`
	DefaultName      string   `json:"default_name"`
	EnableDices      bool     `json:"enable_dices"`
	EnableFlags      bool     `json:"enable_flags"`
	EnableIcons      bool     `json:"enable_icons"`
	EnableLikes      bool     `json:"enable_likes"`
	EnableNames      bool     `json:"enable_names"`
	EnablePosting    bool     `json:"enable_posting"`
	EnableSage       bool     `json:"enable_sage"`
	EnableShield     bool     `json:"enable_shield"`
	EnableSubject    bool     `json:"enable_subject"`
	EnableThreadTags bool     `json:"enable_thread_tags"`
	EnableTrips      bool     `json:"enable_trips"`
	FileTypes        []string `json:"file_types"`
	ID               string   `json:"id"`
	Info             string   `json:"info"`
	InfoOuter        string   `json:"info_outer"`
	MaxComment       int      `json:"max_comment"`
	MaxFilesSize     int      `json:"max_files_size"`
	MaxPages         int      `json:"max_pages"`
	Name             string   `json:"name"`
	ThreadsPerPage   int      `json:"threads_per_page"`
}

type Thread struct {
	Message
	MessagesCount uint `json:"posts_count"`
}

func GetBoardCatalog(section string) (*Board, []Thread, error) {
	threadURL := siteURL + `/` + section + `/catalog.json`

	req, err := http.Get(threadURL)
	if err != nil {
		return nil, nil, err
	}
	defer req.Body.Close()

	var resp struct {
		Board   Board    `json:"board"`
		Threads []Thread `json:"threads"`
	}

	if err = json.NewDecoder(req.Body).Decode(&resp); err != nil {
		return nil, nil, err
	}

	return &resp.Board, resp.Threads, nil
}
