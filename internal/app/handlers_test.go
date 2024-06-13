package app

import (
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
	Storage: &UrlStorage{
		DBFileName: "db_test.json",
	},
}

func TestUrlHandler_shrinkUrlHandler(t *testing.T) {

	type want struct {
		statusCode   int
		contentType  string
		isCorrectUrl bool
	}

	tests := []struct {
		name        string
		url         string
		method      string
		requestBody string
		want        want
	}{
		{
			name:        "post test #1: good",
			url:         "/",
			method:      http.MethodPost,
			requestBody: originalUrl,
			want: want{
				statusCode:   http.StatusCreated,
				contentType:  "text/plain",
				isCorrectUrl: true,
			},
		},
		{
			name:        "post test #2: bad request url",
			url:         "/bad/bad",
			method:      http.MethodPost,
			requestBody: originalUrl,
			want: want{
				statusCode: http.StatusNotFound,
			},
		},
	}
	for it, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()
			router.POST("/", urlHandler.ShrinkUrlHandler)
			request := httptest.NewRequest(tt.method, tt.url, strings.NewReader(tt.requestBody))
			w := httptest.NewRecorder()
			router.ServeHTTP(w, request)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)

			if result.StatusCode == http.StatusCreated {
				assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))
			}
			defer result.Body.Close()
			resBody, _ := io.ReadAll(result.Body)

			_, urlParseErr := url.Parse(string(resBody))
			if tt.want.isCorrectUrl {
				require.NoError(t, urlParseErr)
				if it == 0 { //FIXME
					shortUrlId = path.Base(string(resBody))
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
