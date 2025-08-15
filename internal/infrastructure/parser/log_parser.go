package parser

import "github.com/4domm/ngxstat/internal/domain"

type LogParser interface {
	ParseLogLine(string) (*domain.LogData, error)
}
