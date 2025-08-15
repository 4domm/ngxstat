package parser

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/4domm/ngxstat/internal/domain"
)

var (
	pattern         = regexp.MustCompile("^(\\S+) (\\S+) (\\S+) \\[(.*?)] \"(\\S+) (\\S+) \\S+\" (\\d{3}) (\\d+)(?: \"(.*?)\" \"(.*?)\")?")
	NginxDateFormat = "02/Jan/2006:15:04:05 -0700"
)
var (
	ErrLogFormat = errors.New("неправильный формат лога")
	ErrLogData   = errors.New("лог содержит недопустимые данные")
)

type NginxParser struct {
}

func (np NginxParser) ParseLogLine(logLine string) (*domain.LogData, error) {
	data := strings.Split(logLine, "$")
	if len(data) != 2 {
		return nil, ErrLogFormat
	}

	matches := pattern.FindStringSubmatch(data[1])
	if matches == nil {
		return nil, ErrLogFormat
	}

	logData := &domain.LogData{
		Filename:     data[0],
		IPAddress:    ParseIPAddress(matches[1]),
		RemoteUser:   ParseRemoteUser(matches[3]),
		Timestamp:    ParseTimestamp(matches[4]),
		Method:       matches[5],
		Resource:     matches[6],
		StatusCode:   ParseStatusCode(matches[7]),
		ResponseSize: ParseResponseSize(matches[8]),
		Referer:      ParseOptionalField(matches[9]),
		UserAgent:    ParseOptionalField(matches[10]),
	}
	if !isValidRequiredFields(logData) || !isValidStatusCode(logData.StatusCode) {
		return nil, ErrLogData
	}

	return logData, nil
}

func ParseIPAddress(field string) string {
	if field != "-" {
		return field
	}

	return ""
}

func ParseRemoteUser(field string) string {
	if field != "-" {
		return field
	}

	return ""
}

func ParseTimestamp(field string) time.Time {
	timestamp, err := time.Parse(NginxDateFormat, field)
	if err != nil {
		return time.Time{}
	}

	return timestamp
}

func ParseStatusCode(field string) string {
	if field != "-" {
		return field
	}

	return ""
}

func ParseResponseSize(field string) int64 {
	size, err := strconv.Atoi(field)
	if err != nil {
		return 0
	}

	return int64(size)
}

func ParseOptionalField(field string) string {
	if field != "" && field != "-" {
		return field
	}

	return ""
}

func isValidRequiredFields(logData *domain.LogData) bool {
	return logData.IPAddress != "" &&
		!logData.Timestamp.IsZero() &&
		logData.Method != "" &&
		logData.Resource != "" &&
		logData.StatusCode != "" &&
		logData.ResponseSize >= 0
}

func isValidStatusCode(statusCode string) bool {
	strStatusCode, err := strconv.Atoi(statusCode)
	if err != nil {
		return false
	}

	return strStatusCode >= 100 && strStatusCode < 600
}

func GetFilterFunction(filterField domain.FilterField, filterValue string) func(*domain.LogData) bool {
	switch filterField {
	case domain.AGENT:
		return func(log *domain.LogData) bool {
			return log.UserAgent != "" && strings.EqualFold(log.UserAgent, filterValue)
		}
	case domain.METHOD:
		return func(log *domain.LogData) bool {
			return log.Method != "" && strings.EqualFold(log.Method, filterValue)
		}
	case domain.STATUS:
		return func(log *domain.LogData) bool { return log.StatusCode == filterValue }
	case domain.RESOURCE:
		return func(log *domain.LogData) bool {
			return log.Resource != "" && strings.EqualFold(log.Resource, filterValue)
		}
	case domain.REFERER:
		return func(log *domain.LogData) bool {
			return log.Referer != "" && strings.EqualFold(log.Referer, filterValue)
		}
	case domain.REMOTEUSER:
		return func(log *domain.LogData) bool {
			return log.RemoteUser != "" && strings.EqualFold(log.RemoteUser, filterValue)
		}
	case domain.SIZE:
		return func(log *domain.LogData) bool { return fmt.Sprintf("%d", log.ResponseSize) == filterValue }
	}

	return nil
}
