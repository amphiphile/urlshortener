package app

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strings"
	"testing"
)

var (
	originalURL = "https://test.ru"
	shortURLId  string
)

var urlHandler = &URLHandler{
	Storage: &URLStorage{
		BaseURL:    "http://localhost:8080/",
		DBFileName: "db-test.json",
	},
}

func TestURLHandler_handleShrinkURL(t *testing.T) {

	type want struct {
		statusCode  int
		contentType string
	}

	tests := []struct {
		name        string
		url         string
		method      string
		originalURL string
		contentType string
		want        want
	}{
		{
			name:        "post test #1: good text/plain",
			url:         "/",
			method:      http.MethodPost,
			originalURL: originalURL,
			contentType: "text/plain",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "text/plain",
			},
		},
		{
			name:        "post test #2: good application/json",
			url:         "/api/shorten",
			method:      http.MethodPost,
			originalURL: originalURL,
			contentType: "application/json",
			want: want{
				statusCode:  http.StatusCreated,
				contentType: "application/json",
			},
		},
		{
			name:        "post test #3: bad request url",
			url:         "/bad/bad",
			method:      http.MethodPost,
			originalURL: originalURL,
			contentType: "application/json",
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()

			if tt.contentType == "application/json" {

				router.POST("/api/shorten", urlHandler.HandleShrinkURLJSON)

				requestBody, _ := json.Marshal(shrinkRequest{
					URL: tt.originalURL,
				})

				request := httptest.NewRequest(tt.method, tt.url, bytes.NewReader(requestBody))
				request.Header.Set("Content-Type", tt.contentType)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, request)
				response := w.Result()
				defer response.Body.Close()

				assert.Equal(t, tt.want.statusCode, response.StatusCode)

				if response.StatusCode == http.StatusCreated {
					assert.True(t, strings.HasPrefix(response.Header.Get("Content-Type"), tt.want.contentType)) //FIXME: charset

					responseBody, err := io.ReadAll(response.Body)
					require.NoError(t, err)

					var responseData shrinkResult

					err = json.Unmarshal(responseBody, &responseData)
					require.NoError(t, err)

					_, urlParseErr := url.Parse(responseData.Result)
					require.NoError(t, urlParseErr)

					shortURLId = path.Base(responseData.Result) //FIXME: вызывать тесты последовательно
				}
			} else if tt.contentType == "text/plain" {
				router.POST("/", urlHandler.HandleShrinkURLText)

				requestBody := tt.originalURL

				request := httptest.NewRequest(tt.method, tt.url, strings.NewReader(requestBody))
				request.Header.Set("Content-Type", tt.contentType)

				w := httptest.NewRecorder()
				router.ServeHTTP(w, request)

				response := w.Result()
				defer response.Body.Close()

				assert.Equal(t, tt.want.statusCode, response.StatusCode)

				if response.StatusCode == http.StatusCreated {
					assert.True(t, strings.HasPrefix(response.Header.Get("Content-Type"), tt.want.contentType)) //FIXME: charset

					defer response.Body.Close()

					resBody, err := io.ReadAll(response.Body)
					require.NoError(t, err)

					_, urlParseErr := url.Parse(string(resBody))
					require.NoError(t, urlParseErr)
				}

			}
		})
	}
}

func TestURLHandler_handleUnwrapURL(t *testing.T) {

	type want struct {
		statusCode int
		location   string
	}

	tests := []struct {
		name   string
		url    string
		method string
		want   want
	}{
		{
			name:   "get test #1: good",
			url:    "/" + shortURLId,
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				location:   originalURL,
			},
		},
		{
			name:   "get test #2: not existing url",
			url:    "/bad",
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()
			router.GET("/:id", urlHandler.HandleUnwrapURL)
			request := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, request)

			response := w.Result()
			defer response.Body.Close()

			assert.Equal(t, tt.want.statusCode, response.StatusCode)

			if response.StatusCode == http.StatusTemporaryRedirect {
				assert.Equal(t, tt.want.location, response.Header.Get("Location"))
			}
		})
	}
}
