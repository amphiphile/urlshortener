package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
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
	BaseUrl string
	Storage UrlShrinkerUnwrapper
}

type shrinkRequest struct {
	Url string `json:"url"`
}
type shrinkResult struct {
	Result string `json:"result"`
}

func (h *UrlHandler) ShrinkUrlJsonHandler(c *gin.Context) {
	contentType := c.Request.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") { //FIXME: charset

		var request shrinkRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		originalUrl := request.Url
		id, _ := h.Storage.ShrinkUrl(originalUrl)

		response := shrinkResult{
			Result: h.BaseUrl + id,
		}
		c.JSON(http.StatusCreated, response)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unsupported content type: %s", contentType)})
		return
	}

}

func (h *UrlHandler) ShrinkUrlTextHandler(c *gin.Context) {
	contentType := c.Request.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "text/plain") { //FIXME: charset

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		originalUrl := string(body[:])
		id, _ := h.Storage.ShrinkUrl(originalUrl)

		c.String(http.StatusCreated, h.BaseUrl+id)

	} else {
		c.String(http.StatusBadRequest, fmt.Sprintf("Unsupported content type: %s", contentType))
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
