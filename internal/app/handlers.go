package app

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/url"
	"os"
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

func (h *URLHandler) HandleShrinkURLJSON(c *gin.Context) {

	var request shrinkRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.Storage.ShrinkURL(request.URL)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, shrinkResult{
		Result: result,
	})
	return

}

func (h *URLHandler) HandleShrinkURLText(c *gin.Context) {

	//FIXME: разобраться с bind для plaintext
	if c.ContentType() != binding.MIMEPlain {
		c.Abort()
		c.String(http.StatusBadRequest, "expected plain text")
		return
	}
	defer c.Request.Body.Close()

	request, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Abort()
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.Storage.ShrinkURL(string(request))
	if err != nil {
		c.Abort()
		c.String(http.StatusBadRequest, err.Error())
		return

	}

	c.String(http.StatusCreated, result)
	return

}

func (h *URLHandler) HandleUnwrapURL(c *gin.Context) {
	id := c.Param("id")
	originalURL, err := h.Storage.UnwrapURL(id)
	if err != nil {
		c.Abort()
		c.String(http.StatusBadRequest, "Requested url not found")
		return
	}
	c.Redirect(http.StatusTemporaryRedirect, originalURL)
}

type URLStorage struct {
	DBFileName string
	BaseURL    string
}
type urlsMap map[string]string

func (u *URLStorage) ShrinkURL(originalURL string) (string, error) {
	urls, err := u.readFromDB()
	if err != nil {
		return "", err

	}

	id := uuid.New().String()
	urls[id] = originalURL

	err = u.writeToDB(urls)
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
	urls, err := u.readFromDB()
	if err != nil {
		return "", err
	}

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
		f, err := os.Create(u.DBFileName)
		defer f.Close()
		if err != nil {
			return urls, err
		}
		return urls, nil
	} else if fileInfo.Size() == 0 {
		return urls, nil
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
