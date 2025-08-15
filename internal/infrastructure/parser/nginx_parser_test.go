package parser_test

import (
	"testing"
	"time"

	"github.com/4domm/ngxstat/internal/infrastructure/parser"
	"github.com/stretchr/testify/assert"
)

func TestParseLogLine(t *testing.T) {
	nginxParser := parser.NginxParser{}

	t.Run("Valid Log Line", func(t *testing.T) {
		logLine := `$127.0.0.1 - admin [10/Oct/2023:13:55:36 +0000] "GET /index.html HTTP/1.1" 200 1234 "http://example.com" "Mozilla/5.0"`
		logData, err := nginxParser.ParseLogLine(logLine)

		assert.NoError(t, err)
		assert.NotNil(t, logData)
		assert.Equal(t, "127.0.0.1", logData.IPAddress)
		assert.Equal(t, "admin", logData.RemoteUser)
		assert.Equal(t, "GET", logData.Method)
		assert.Equal(t, "/index.html", logData.Resource)
		assert.Equal(t, "200", logData.StatusCode)
		assert.Equal(t, int64(1234), logData.ResponseSize)
		assert.Equal(t, "http://example.com", logData.Referer)
		assert.Equal(t, "Mozilla/5.0", logData.UserAgent)

		expectedTime, _ := time.Parse(parser.NginxDateFormat, "10/Oct/2023:13:55:36 +0000")
		assert.Equal(t, expectedTime, logData.Timestamp)
	})

	t.Run("Log Line Missing Optional Fields", func(t *testing.T) {
		logLine := `$192.168.1.1 - - [11/Nov/2023:10:10:10 +0000] "POST /submit HTTP/1.1" 201 0`
		logData, err := nginxParser.ParseLogLine(logLine)

		assert.NoError(t, err)
		assert.NotNil(t, logData)
		assert.Equal(t, "192.168.1.1", logData.IPAddress)
		assert.Equal(t, "", logData.RemoteUser)
		assert.Equal(t, "POST", logData.Method)
		assert.Equal(t, "/submit", logData.Resource)
		assert.Equal(t, "201", logData.StatusCode)
		assert.Equal(t, int64(0), logData.ResponseSize)
		assert.Equal(t, "", logData.Referer)
		assert.Equal(t, "", logData.UserAgent)

		expectedTime, _ := time.Parse(parser.NginxDateFormat, "11/Nov/2023:10:10:10 +0000")
		assert.Equal(t, expectedTime, logData.Timestamp)
	})

	t.Run("Invalid Log Line Format", func(t *testing.T) {
		logLine := `INVALID LOG FORMAT`
		logData, err := nginxParser.ParseLogLine(logLine)

		assert.Error(t, err)
		assert.Equal(t, parser.ErrLogFormat, err)
		assert.Nil(t, logData)
	})

	t.Run("Invalid Timestamp Format", func(t *testing.T) {
		logLine := `$127.0.0.1 - admin [INVALID_TIMESTAMP] "GET /index.html HTTP/1.1" 200 1234`
		logData, err := nginxParser.ParseLogLine(logLine)

		assert.Error(t, err)
		assert.Equal(t, parser.ErrLogData, err)
		assert.Nil(t, logData)
	})

	t.Run("Missing Required Fields", func(t *testing.T) {
		logLine := `- - - [-] "-" "-" - -`
		logData, err := nginxParser.ParseLogLine(logLine)

		assert.Error(t, err)
		assert.Equal(t, parser.ErrLogFormat, err)
		assert.Nil(t, logData)
	})

	t.Run("Invalid Status Code", func(t *testing.T) {
		logLine := `$127.0.0.1 - admin [10/Oct/2023:13:55:36 +0000] "GET /index.html HTTP/1.1" 9999 1234`
		logData, err := nginxParser.ParseLogLine(logLine)

		assert.Error(t, err)
		assert.Equal(t, parser.ErrLogFormat, err)
		assert.Nil(t, logData)
	})

	t.Run("Negative Response Size", func(t *testing.T) {
		logLine := `$127.0.0.1 - admin [10/Oct/2023:13:55:36 +0000] "GET /index.html HTTP/1.1" 200 -1234`
		logData, err := nginxParser.ParseLogLine(logLine)

		assert.Error(t, err)
		assert.Equal(t, parser.ErrLogFormat, err)
		assert.Nil(t, logData)
	})
}
