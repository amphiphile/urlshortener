package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/url"
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
		result, _ := h.Storage.ShrinkURL(request.URL)

		c.JSON(http.StatusCreated, shrinkResult{
			Result: result,
		})
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
		result, _ := h.Storage.ShrinkURL(string(body[:]))

		c.String(http.StatusCreated, result)

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
	BaseURL    string
}
type urlsMap map[string]string

func (u *URLStorage) ShrinkURL(originalURL string) (string, error) {
	urls, _ := u.readFromDB()

	id := encode(originalURL)
	urls[id] = originalURL

	err := u.writeToDB(urls)
	if err != nil {
		return "", err
	}

	result, err := url.JoinPath(u.BaseURL, id)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (u *URLStorage) UnwrapURL(id string) (string, error) {
	urls, _ := u.readFromDB()

	originalURL, ok := urls[id]
	if !ok {
		return "", errors.New("URL not found")
	}
	return originalURL, nil
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
