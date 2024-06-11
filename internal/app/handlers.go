package app

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
)

type UrlShrinkerUnwrapper interface {
	ShrinkUrl(url string) (string, error)
	UnwrapUrl(id string) (string, error)
}

type UrlHandler struct {
	Storage UrlShrinkerUnwrapper
}

func (h *UrlHandler) RequestHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.ShrinkUrlHandler(w, r)
	case http.MethodGet:
		h.UnwrapUrlHandler(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusBadRequest) //FIXME
	}
}

func (h *UrlHandler) ShrinkUrlHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	originalUrl := string(body[:])
	id, _ := h.Storage.ShrinkUrl(originalUrl)

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte("http://localhost:8080/" + id))
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

}

func (h *UrlHandler) UnwrapUrlHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.Trim(r.URL.Path, "/")
	originalUrl, err := h.Storage.UnwrapUrl(id)
	if err != nil {
		http.Error(w, "Requested url not found", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", originalUrl)
	w.WriteHeader(http.StatusTemporaryRedirect)

}

type UrlStorage struct {
	DBFileName string
}
type urlsMap map[string]string

func (u *UrlStorage) ShrinkUrl(url string) (string, error) {
	urls, _ := u.readFromDB()

	id := encode(url)
	urls[id] = url

	err := u.writeToDB(urls)
	if err != nil {
		panic(err)
	}

	return id, nil
}

func (u *UrlStorage) UnwrapUrl(id string) (string, error) {
	urls, _ := u.readFromDB()

	url, ok := urls[id]
	if !ok {
		return "", errors.New("URL not found")
	}
	return url, nil
}

func (u *UrlStorage) readFromDB() (urlsMap, error) {

	urls := make(urlsMap)

	fileInfo, err := os.Stat(u.DBFileName)
	if os.IsNotExist(err) {
		return urls, err
	} else if fileInfo.Size() == 0 {
		return urls, err
	} else {
		urlsString, err := os.ReadFile(u.DBFileName)
		if err != nil {
			return urls, err
		}

		err = json.Unmarshal(urlsString, &urls)
		if err != nil {
			return urls, err
		}

	}

	return urls, nil
}
func (u *UrlStorage) writeToDB(urls urlsMap) error {

	urlsJson, err := json.Marshal(urls)
	if err != nil {
		return err
	}

	err = os.WriteFile(u.DBFileName, urlsJson, 0644)

	if err != nil {
		return err
	}

	return nil
}
