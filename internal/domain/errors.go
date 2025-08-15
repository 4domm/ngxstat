package domain

import (
	"errors"
	"fmt"
)

var ErrDownload = errors.New("failed to download file")
var ErrFinding = errors.New("no files")

type InvalidOutputFormatError struct {
	Format string
}

func (e *InvalidOutputFormatError) Error() string {
	return fmt.Sprintf("Неверный формат вывода: %s. Используйте %s или %s.", e.Format, MARKDOWN, ADOC)
}

type MissingFilterValueError struct {
}

func (e *MissingFilterValueError) Error() string {
	return "Для фильтрации по полю необходимо указать значение фильтра."
}

type InvalidFilterCombinationError struct {
}

func (e *InvalidFilterCombinationError) Error() string {
	return "Для фильтрации необходимо указать оба параметра: --filter-field и --filter-value."
}
