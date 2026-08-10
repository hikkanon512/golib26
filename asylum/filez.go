package asylum

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
)

const siteURL = `https://2ch.su`

func (m Message) diff(v Message) bool {
	if len(m.Files) != len(v.Files) {
		return true
	}
	for i := range m.Files {
		if m.Files[i].Name != v.Files[i].Name || m.Files[i].Path != v.Files[i].Path ||
			m.Files[i].MD5 != v.Files[i].MD5 || m.Files[i].Thumbnail != v.Files[i].Thumbnail ||
			m.Files[i].Size != v.Files[i].Size || m.Files[i].FullName != v.Files[i].FullName {
			return true
		}
	}
	return m.Trip != v.Trip || m.Timestamp != v.Timestamp || m.Tags != v.Tags ||
		m.Subject != v.Subject || m.Sticky != v.Sticky || m.Parent != v.Parent ||
		m.Op != v.Op || m.Num != v.Num || m.Name != v.Name ||
		m.Email != v.Email || m.Date != v.Date || m.Comment != v.Comment ||
		m.Closed != v.Closed || m.Board != v.Board || m.Banned != v.Banned
}

type File struct {
	DisplayName string `json:"displayname"`
	FullName    string `json:"fullname"`
	Height      uint   `json:"height"`
	MD5         string `json:"md5"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	Size        uint   `json:"size"`
	Thumbnail   string `json:"thumbnail"`
	TnHeight    uint   `json:"tn_height"`
	TnWidth     uint   `json:"tn_width"`
	Type        uint   `json:"type"`
	Width       uint   `json:"width"`
}

// owner: rw
// group: r
const modeOwningNonExec = fs.FileMode(0640)

func loadFileIfNotExist(url, path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, modeOwningNonExec)
	if err != nil {
		return err
	}
	defer file.Close()

	request, err := http.Get(url)
	if err != nil {
		return err
	}
	defer request.Body.Close()

	if request.StatusCode != http.StatusOK {
		if request.StatusCode == http.StatusNotFound {
			return ErrThreadNotFound
		}
		return fmt.Errorf("site returned %d: %w", request.StatusCode, ErrUnexpected)
	}

	if _, err = file.ReadFrom(request.Body); err != nil {
		return err
	}
	return nil
}
