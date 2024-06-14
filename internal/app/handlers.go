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

type URLShrinkerUnwrapper interface {
	ShrinkURL(url string) (string, error)
	UnwrapURL(id string) (string, error)
}

type URLHandler struct {
	BaseURL string
	Storage URLShrinkerUnwrapper
}

type shrinkRequest struct {
	URL string `json:"url"`
}
type shrinkResult struct {
	Result string `json:"result"`
}

func (h *URLHandler) ShrinkURLJSONHandler(c *gin.Context) {
	contentType := c.Request.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") { //FIXME: charset

		var request shrinkRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		originalURL := request.URL
		id, _ := h.Storage.ShrinkURL(originalURL)

		response := shrinkResult{
			Result: h.BaseURL + id,
		}
		c.JSON(http.StatusCreated, response)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unsupported content type: %s", contentType)})
		return
	}

}

func (h *URLHandler) ShrinkURLTextHandler(c *gin.Context) {
	contentType := c.Request.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "text/plain") { //FIXME: charset
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}
		originalURL := string(body[:])
		id, _ := h.Storage.ShrinkURL(originalURL)

		c.String(http.StatusCreated, h.BaseURL+id)

	} else {
		c.String(http.StatusBadRequest, fmt.Sprintf("Unsupported content type: %s", contentType))
		return
	}

}

func (h *URLHandler) UnwrapURLHandler(c *gin.Context) {
	id := c.Param("id")
	originalURL, err := h.Storage.UnwrapURL(id)
	if err != nil {
		http.Error(c.Writer, "Requested url not found", http.StatusBadRequest)
		return
	}
	c.Writer.Header().Set("Location", originalURL)
	c.Writer.WriteHeader(http.StatusTemporaryRedirect)

}

type URLStorage struct {
	DBFileName string
}
type urlsMap map[string]string

func (u *URLStorage) ShrinkURL(url string) (string, error) {
	urls, _ := u.readFromDB()

	id := encode(url)
	urls[id] = url

	err := u.writeToDB(urls)
	if err != nil {
		panic(err)
	}

	return id, nil
}

func (u *URLStorage) UnwrapURL(id string) (string, error) {
	urls, _ := u.readFromDB()

	url, ok := urls[id]
	if !ok {
		return "", errors.New("URL not found")
	}
	return url, nil
}

func (u *URLStorage) readFromDB() (urlsMap, error) {

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
func (u *URLStorage) writeToDB(urls urlsMap) error {

	urlsJSON, err := json.Marshal(urls)
	if err != nil {
		return err
	}

	err = os.WriteFile(u.DBFileName, urlsJSON, 0644)

	if err != nil {
		return err
	}

	return nil
}
