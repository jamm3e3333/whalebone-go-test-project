package gin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	ginpkg "github.com/gin-gonic/gin"
	"github.com/jamm3e3333/whalebone-go-test-project/pkg/logger"
	"github.com/rs/zerolog/log"
)

type RequestResponseLog struct {
	Request  *Request  `json:"request"`
	Response *Response `json:"response"`
	Message  string    `json:"message"`
}

type Request struct {
	Method     string          `json:"method"`
	URI        string          `json:"uri"`
	Body       json.RawMessage `json:"body"`
	RemoteAddr string          `json:"remote_addr"`
	UserAgent  string          `json:"user_agent"`
}

type Response struct {
	Status int             `json:"status"`
	Body   json.RawMessage `json:"body"`
}

type bodyWriter struct {
	ginpkg.ResponseWriter
	bodyBuf *bytes.Buffer
}

func (w *bodyWriter) Write(b []byte) (int, error) {
	w.bodyBuf.Write(b)
	return w.ResponseWriter.Write(b)
}

type LoggerMiddlewareConfig struct {
	ignoredPaths map[string]bool
}

func NewLoggerMiddlewareConfig(
	ignoredPaths []string,
) *LoggerMiddlewareConfig {
	l := &LoggerMiddlewareConfig{
		ignoredPaths: map[string]bool{},
	}

	for _, p := range ignoredPaths {
		l.ignoredPaths[p] = true
	}

	return l
}

func LoggerMiddleware(config *LoggerMiddlewareConfig, lg logger.Logger) ginpkg.HandlerFunc {
	return func(c *ginpkg.Context) {
		var requestBodyBytes []byte
		var requestBodyBytesLogging *bytes.Buffer
		var responseBodyWriter *bodyWriter

		if _, ok := config.ignoredPaths[c.FullPath()]; ok {
			c.Next()
			return
		}

		if c.Request.Body != nil {
			requestBodyBytes, _ = io.ReadAll(c.Request.Body)
			requestBodyBytesLogging = bytes.NewBuffer(requestBodyBytes)
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBodyBytes))

		responseBodyWriter = &bodyWriter{
			bodyBuf:        bytes.NewBufferString(""),
			ResponseWriter: c.Writer}
		c.Writer = responseBodyWriter

		c.Next()

		var requestBody map[string]any
		if requestBodyBytesLogging != nil && requestBodyBytesLogging.Len() > 0 {
			if err := json.Unmarshal(requestBodyBytesLogging.Bytes(), &requestBody); err != nil {
				log.Error().Msg(fmt.Sprintf("middleware.logger: error unmarshalling request body: %s", err))
			}
		}
		requestMapMetadata := map[string]any{
			"method":      c.Request.Method,
			"uri":         c.Request.RequestURI,
			"body":        requestBody,
			"remote_addr": c.Request.RemoteAddr,
			"user_agent":  c.Request.UserAgent(),
		}
		var responseBody map[string]any
		if len(responseBodyWriter.bodyBuf.Bytes()) > 0 {
			if err := json.Unmarshal(responseBodyWriter.bodyBuf.Bytes(), &responseBody); err != nil {
				log.Error().Msg(fmt.Sprintf("middleware.logger: error unmarshalling response body: %s", err))
			}
		}
		responseMapMetadata := map[string]any{
			"status": c.Writer.Status(),
			"body":   responseBody,
		}
		lg.InfoWithMetadata("HTTP Request / Response", map[string]any{
			"request":  requestMapMetadata,
			"response": responseMapMetadata,
		})
	}
}
