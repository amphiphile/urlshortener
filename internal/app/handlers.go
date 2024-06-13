package app

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
)

type UrlShrinkerUnwrapper interface {
	ShrinkUrl(url string) (string, error)
	UnwrapUrl(id string) (string, error)
}

type UrlHandler struct {
	Storage UrlShrinkerUnwrapper
}

func (h *UrlHandler) ShrinkUrlHandler(c *gin.Context) {
	if c.Request.URL.Path != "/" {
		http.Error(c.Writer, "Bad request", http.StatusBadRequest)
		return
	}
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		http.Error(c.Writer, "Bad request", http.StatusBadRequest)
		return
	}
	originalUrl := string(body[:])
	id, _ := h.Storage.ShrinkUrl(originalUrl)

	c.Writer.Header().Set("content-type", "text/plain")
	c.Writer.WriteHeader(http.StatusCreated)

	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	_, err = c.Writer.Write([]byte(scheme + "://" + c.Request.Host + c.Request.URL.Path + id))
	if err != nil {
		http.Error(c.Writer, "Bad request", http.StatusBadRequest)
		return
	}

}

func (h *UrlHandler) UnwrapUrlHandler(c *gin.Context) {
	id := c.Param("id")
	originalUrl, err := h.Storage.UnwrapUrl(id)
	if err != nil {
		http.Error(c.Writer, "Requested url not found", http.StatusBadRequest)
		return
	}
	c.Writer.Header().Set("Location", originalUrl)
	c.Writer.WriteHeader(http.StatusTemporaryRedirect)

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
