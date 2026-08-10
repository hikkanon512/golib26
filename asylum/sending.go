package asylum

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strconv"
)

type PostingOptions struct {
	Board     string // Доска.
	WithSage  bool   // Не бамп.
	Email     string // Email-адрес.
	Subject   string // Тема.
	Icon      string // Номер иконки в некоторых (прим. /re/) разделах.
	UserCode  string // Связано с пасскодом.
	Challenge string
	Headers   map[string]string // Заголовки из браузера.
	// В заголовки передаётся структура с полями, пример:
	// "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/150.0.0.0 Safari/537.36"
	// "Sec-Fetch-Dest": "empty"
	// "Sec-Fetch-Mode": "cors"
	// "Sec-Fetch-Site": "same-origin"
	// "X-Requested-With": "XMLHttpRequest"
}

func (po *PostingOptions) getIcon() string {
	if po.Icon == "" {
		return "-1"
	}
	return po.Icon
}

func SendMessage(threadID int, message string, filePaths []string, options *PostingOptions) (int, error) {
	const skipMagic = `\skip\`

	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	sage := skipMagic
	email := options.Email
	if options.WithSage {
		sage = "on"
		email = "sage"
	}
	fields := [][2]string{
		{"task", "post"},
		{"board", options.Board},
		{"thread", strconv.Itoa(threadID)},
		{"usercode", options.UserCode},
		{"code", ""},
		{"captcha_type", "emoji_captcha"},
		{"subject", options.Subject},
		{"comment", message},
		{"icon", options.getIcon()},
		{"emoji_captcha_id", ""},
		{"makaka_id", ""},
		{"makaka_answer", ""},
		{"2ch_challenge", options.Challenge},
		{"email", email},
		{"sage", sage},
	}

	for _, field := range fields {
		if field[1] == skipMagic {
			continue
		} else if err := writer.WriteField(field[0], field[1]); err != nil {
			return 0, err
		}
	}

	for _, filePath := range filePaths {
		_, fileName := filepath.Split(filePath)
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			return 0, err
		}

		header := make(textproto.MIMEHeader)
		header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file[]"; filename="%s"`, fileName))
		header.Set("Content-Type", mime.TypeByExtension(filepath.Ext(fileName)))
		requestPart, err := writer.CreatePart(header)
		if err != nil {
			return 0, err
		}
		if _, err = requestPart.Write(fileContent); err != nil {
			return 0, err
		}
	}

	if err := writer.Close(); err != nil {
		return 0, err
	}

	req, err := http.NewRequest(http.MethodPost, siteURL+"/user/posting?nc=1", &buffer)
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	for h, v := range options.Headers {
		req.Header.Set(h, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("site returned %d response code", resp.StatusCode)
	}

	var answer SendMessageResponse
	if err = json.NewDecoder(resp.Body).Decode(&answer); err != nil {
		return 0, err
	} else if answer.Result != 1 {
		// code 1 is used for success result
		return 0, fmt.Errorf("site returned error %d: %s", answer.Error.Code, answer.Error.Message)
	}
	return answer.Num, nil
}

type SendMessageResponse struct {
	Num    int `json:"num"`
	Result int `json:"result"`
	Error  struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}
