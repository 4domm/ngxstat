package domain

import "time"

type LogData struct {
	Filename     string
	Timestamp    time.Time
	IPAddress    string
	RemoteUser   string
	Method       string
	Resource     string
	StatusCode   string
	ResponseSize int64
	Referer      string
	UserAgent    string
}
