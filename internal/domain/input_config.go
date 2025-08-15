package domain

import (
	"time"
)

type FilterField string

const (
	AGENT      FilterField = "agent"
	METHOD     FilterField = "method"
	STATUS     FilterField = "status"
	RESOURCE   FilterField = "resource"
	REFERER    FilterField = "referer"
	REMOTEUSER FilterField = "remote_user"
	SIZE       FilterField = "size"
	ADOC                   = "adoc"
	MARKDOWN               = "markdown"
)

var FilterFields = []FilterField{AGENT, METHOD, STATUS, RESOURCE, REFERER, REMOTEUSER, SIZE, ""}

type InputConfig struct {
	Path         string
	From         time.Time
	To           time.Time
	OutputFormat string
	FilterField  FilterField
	FilterValue  string
}
