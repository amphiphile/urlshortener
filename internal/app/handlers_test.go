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
	originalUrl = "https://test.ru"
	shortUrlId  string
)

var urlHandler = &UrlHandler{
	BaseUrl: "http://localhost:8080/",
	Storage: &UrlStorage{
		DBFileName: "db.json",
	},
}

func TestUrlHandler_shrinkUrlHandler(t *testing.T) {

	type want struct {
		statusCode  int
		contentType string
	}

	tests := []struct {
		name        string
		url         string
		method      string
		originalUrl string
		contentType string
		want        want
	}{
		{
			name:        "post test #1: good text/plain",
			url:         "/",
			method:      http.MethodPost,
			originalUrl: originalUrl,
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
			originalUrl: originalUrl,
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
			originalUrl: originalUrl,
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

				router.POST("/api/shorten", urlHandler.ShrinkUrlJsonHandler)

				requestBody, _ := json.Marshal(shrinkRequest{
					Url: tt.originalUrl,
				})

				request := httptest.NewRequest(tt.method, tt.url, bytes.NewReader(requestBody))
				request.Header.Set("Content-Type", tt.contentType)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, request)
				response := w.Result()

				assert.Equal(t, tt.want.statusCode, response.StatusCode)

				if response.StatusCode == http.StatusCreated {
					assert.True(t, strings.HasPrefix(response.Header.Get("Content-Type"), tt.want.contentType)) //FIXME: charset

					defer func(Body io.ReadCloser) {
						err := Body.Close()
						require.NoError(t, err)
					}(response.Body)

					responseBody, err := io.ReadAll(response.Body)
					require.NoError(t, err)

					var responseData shrinkResult

					err = json.Unmarshal(responseBody, &responseData)
					require.NoError(t, err)

					_, urlParseErr := url.Parse(responseData.Result)
					require.NoError(t, urlParseErr)

					shortUrlId = path.Base(responseData.Result) //FIXME: вызывать тесты последовательно
				}
			} else if tt.contentType == "text/plain" {
				router.POST("/", urlHandler.ShrinkUrlTextHandler)

				requestBody := tt.originalUrl

				request := httptest.NewRequest(tt.method, tt.url, strings.NewReader(requestBody))
				request.Header.Set("Content-Type", tt.contentType)

				w := httptest.NewRecorder()
				router.ServeHTTP(w, request)

				response := w.Result()

				assert.Equal(t, tt.want.statusCode, response.StatusCode)

				if response.StatusCode == http.StatusCreated {
					assert.True(t, strings.HasPrefix(response.Header.Get("Content-Type"), tt.want.contentType)) //FIXME: charset

					defer func(Body io.ReadCloser) {
						err := Body.Close()
						require.NoError(t, err)
					}(response.Body)

					resBody, err := io.ReadAll(response.Body)
					require.NoError(t, err)

					_, urlParseErr := url.Parse(string(resBody))
					require.NoError(t, urlParseErr)
				}

			}
		})
	}
}

func TestUrlHandler_unwrapUrlHandler(t *testing.T) {

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
			url:    "/" + shortUrlId,
			method: http.MethodGet,
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				location:   originalUrl,
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
			router.GET("/:id", urlHandler.UnwrapUrlHandler)
			request := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, request)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			if result.StatusCode == http.StatusTemporaryRedirect {
				assert.Equal(t, tt.want.location, result.Header.Get("Location"))
			}
		})
	}
}
