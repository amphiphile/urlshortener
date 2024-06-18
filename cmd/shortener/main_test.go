package main

import (
	"bytes"
	"encoding/json"
	"github.com/amphiphile/urlshortener/internal/app"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
func TestShortener(t *testing.T) {
	suite.Run(t, new(ShortenerTestSuite))
}

type ShortenerTestSuite struct {
	suite.Suite
	urlHandler *app.URLHandler
	router     *gin.Engine

	urlPairs map[string]string
}

func (s *ShortenerTestSuite) SetupSuite() {
	s.urlHandler = &app.URLHandler{
		Storage: &app.URLStorage{
			BaseURL:    "http://localhost:8080/",
			DBFileName: "db-test.json",
		},
	}
	s.router = setupRouter(*s.urlHandler)

	s.urlPairs = make(map[string]string)

}

func (s *ShortenerTestSuite) TestShortener() {
	s.Run("shrink text", func() {
		originalURL := "https://test-text.ru"

		request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalURL))
		request.Header.Set("Content-Type", gin.MIMEPlain)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, request)

		response := w.Result()
		defer response.Body.Close()

		s.Require().Equal(http.StatusCreated, response.StatusCode)

		s.Assert().True(strings.HasPrefix(response.Header.Get("Content-Type"), gin.MIMEPlain))

		responseBody, err := io.ReadAll(response.Body)
		s.Require().NoError(err)

		s.urlPairs[originalURL] = string(responseBody)
	})
	s.Run("shrink json", func() {
		originalURL := "https://test-json.com"

		requestBody, _ := json.Marshal(app.ShrinkRequest{
			URL: originalURL,
		})

		request := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(requestBody))
		request.Header.Set("Content-Type", gin.MIMEJSON)

		w := httptest.NewRecorder()
		s.router.ServeHTTP(w, request)
		response := w.Result()
		defer response.Body.Close()

		s.Require().Equal(http.StatusCreated, response.StatusCode)

		s.Assert().True(strings.HasPrefix(response.Header.Get("Content-Type"), gin.MIMEJSON))

		responseBody, err := io.ReadAll(response.Body)
		s.Require().NoError(err)

		var responseData app.ShrinkResult

		err = json.Unmarshal(responseBody, &responseData)
		s.Require().NoError(err)

		s.urlPairs[originalURL] = responseData.Result

	})
	s.Run("unwrap", func() {
		for originalURL, shortenURL := range s.urlPairs {

			shortURLId := path.Base(shortenURL)

			request := httptest.NewRequest(http.MethodGet, "/"+shortURLId, nil)

			w := httptest.NewRecorder()
			s.router.ServeHTTP(w, request)

			response := w.Result()
			defer response.Body.Close()

			s.Require().Equal(http.StatusTemporaryRedirect, response.StatusCode)

			s.Assert().Equal(originalURL, response.Header.Get("Location"))
		}
	})
}
