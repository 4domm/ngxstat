package client

import (
	"errors"
	"flag"
	"fmt"
	"slices"
	"time"

	"github.com/4domm/ngxstat/internal/domain"
)

const (
	DateFormatWithTime = "2006-01-02T15:04:05-0700"
	DateFormatNoTime   = "2006-01-02"
)

func ParseFlags() (*domain.InputConfig, error) {
	var path, outputFormat, filterValue string

	var filterField string

	flag.StringVar(&path, "path", "", "Путь к лог-файлам или URL")
	flag.StringVar(&outputFormat, "format", "", "Формат вывода (adoc или markdown)")
	flag.StringVar(&filterField, "filter-field", "", "Поле для фильтрации")
	flag.StringVar(&filterValue, "filter-value", "", "Значение для фильтрации")

	var fromStr, toStr string

	flag.StringVar(&fromStr, "from", "", "Начало временного диапазона в формате ISO8601")
	flag.StringVar(&toStr, "to", "", "Конец временного диапазона в формате ISO8601")

	flag.Parse()

	if path == "" {
		return nil, errors.New("use path flag to define path to files")
	}

	var from, to time.Time

	var err error
	from, err = ParseDate(fromStr)

	if err != nil {
		return nil, err
	}

	to, err = ParseDate(toStr)
	if err != nil {
		return nil, err
	}

	if outputFormat != "" && !(outputFormat == domain.MARKDOWN || outputFormat == domain.ADOC) {
		return nil, &domain.InvalidOutputFormatError{Format: outputFormat}
	}

	if filterField != "" && filterValue == "" {
		return nil, &domain.MissingFilterValueError{}
	}

	if (filterField == "" && filterValue != "") || (filterField != "" && filterValue == "") {
		return nil, &domain.InvalidFilterCombinationError{}
	}

	if !slices.Contains(domain.FilterFields, domain.FilterField(filterField)) {
		return nil, fmt.Errorf("не поддерживается фильтрация по данному полю, варианты:%v", domain.FilterFields)
	}

	return &domain.InputConfig{
			FilterField: domain.FilterField(filterField),
			FilterValue: filterValue, From: from,
			To:           to,
			OutputFormat: outputFormat,
			Path:         path},
		nil
}

func ParseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, nil
	}

	parsedTime, err := time.Parse(DateFormatWithTime, dateStr)
	if err == nil {
		return parsedTime, nil
	}

	parsedTime, err = time.Parse(DateFormatNoTime, dateStr)
	if err == nil {
		return parsedTime, nil
	}

	return time.Time{}, fmt.Errorf("неверный формат даты: %v", err)
}
