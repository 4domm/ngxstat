package generator_test

import (
	"os"
	"testing"
	"time"

	"github.com/4domm/ngxstat/internal/infrastructure/client"
	"github.com/4domm/ngxstat/internal/infrastructure/generator"

	"github.com/4domm/ngxstat/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestMarkdownReportGenerator_GenerateReport(t *testing.T) {
	fileWriter := generator.FileWriter{}

	reportGenerator := generator.NewMarkdownReportGenerator(fileWriter)
	defer os.Remove(reportGenerator.GetFilePath())

	result := &domain.AnalysisResult{
		Filenames:                []string{"file1.log", "file2.log"},
		TotalRequests:            10,
		TotalResponseSize:        5000,
		AverageResponseSize:      500,
		Percentile95ResponseSize: 800,
		MostRequestedResources:   map[string]int64{"/index.html": 5, "/about.html": 3},
		MostFrequentStatusCodes:  map[string]int64{"200": 8, "500": 2},
		MostFrequentReferrers:    map[string]int64{"https://example.com": 7},
		TotalServerErrorsLogs:    2,
		From:                     parseTestTime("2023-01-01T00:00:00+0000"),
		To:                       parseTestTime("2023-01-01T23:59:59+0000"),
	}

	reportGenerator.GenerateReport(result)

	output, err := os.ReadFile(reportGenerator.GetFilePath())
	assert.NoError(t, err, "failed to read generated file")

	resStr := string(output)

	assertContains(t, resStr, "#### Общая информация ")
	assertContains(t, resStr, "| Metric                |                 Value |")
	assertContains(t, resStr, "| ---                   |                   --- |")
	assertContains(t, resStr, "| Number of Files       |                     2 |")
	assertContains(t, resStr, "| - File                |             file1.log |")
	assertContains(t, resStr, "| - File                |             file2.log |")
	assertContains(t, resStr, "| Start Date            | 01/Jan/2023:00:00:00 +0000 |")
	assertContains(t, resStr, "| End Date              | 01/Jan/2023:23:59:59 +0000 |")
	assertContains(t, resStr, "| Total Requests        |                    10 |")
	assertContains(t, resStr, "| Average Response Size |                   500 |")
	assertContains(t, resStr, "| 95th Percentile Response Size |                   800 |")
	assertContains(t, resStr, "#### Запрашиваемые ресурсы")
	assertContains(t, resStr, "| Resource              |                 Count |")
	assertContains(t, resStr, "| /index.html           |                     5 |")
	assertContains(t, resStr, "| /about.html           |                     3 |")
	assertContains(t, resStr, "#### Коды ответа")
	assertContains(t, resStr, "| Code                  |                 Count |")
	assertContains(t, resStr, "| 500                   |                     2 |")
	assertContains(t, resStr, "| 200                   |                     8 |")
	assertContains(t, resStr, "### Дополнительная информация:")
	assertContains(t, resStr, "| Metric                |                 Value ")
	assertContains(t, resStr, "| Server Errors (5xx)   |                     2 |")
	assertContains(t, resStr, "### Ссылающиеся ресурсы")
	assertContains(t, resStr, "| Referrer              |                 Count |")
	assertContains(t, resStr, "| https://example.com   |                     7 |")
}

func TestAdocReportGenerator_GenerateReport(t *testing.T) {
	fileWriter := generator.FileWriter{}

	reportGenerator := generator.NewAdocReportGenerator(fileWriter)
	defer os.Remove(reportGenerator.GetFilePath())

	result := &domain.AnalysisResult{
		Filenames:                []string{"file1.log", "file2.log"},
		TotalRequests:            10,
		TotalResponseSize:        5000,
		AverageResponseSize:      500,
		Percentile95ResponseSize: 800,
		MostRequestedResources:   map[string]int64{"/index.html": 5, "/about.html": 3},
		MostFrequentStatusCodes:  map[string]int64{"200": 8, "500": 2},
		MostFrequentReferrers:    map[string]int64{"https://example.com": 7},
		TotalServerErrorsLogs:    2,
		From:                     parseTestTime("2023-01-01T00:00:00+0000"),
		To:                       parseTestTime("2023-01-01T23:59:59+0000"),
	}

	reportGenerator.GenerateReport(result)

	output, err := os.ReadFile(reportGenerator.GetFilePath())
	assert.NoError(t, err, "failed to read generated file")

	resStr := string(output)

	assertContains(t, resStr, "=== Общая информация")
	assertContains(t, resStr, "| Метрика               |              Значение |")
	assertContains(t, resStr, "| --------------------- | --------------------- |")
	assertContains(t, resStr, "| Количество файлов     |                     2 |")
	assertContains(t, resStr, "| Файл:                 |             file1.log |")
	assertContains(t, resStr, "| Файл:                 |             file2.log |")
	assertContains(t, resStr, "| Начальная дата        | 01/Jan/2023:00:00:00 +0000 |")
	assertContains(t, resStr, "| Конечная дата         | 01/Jan/2023:23:59:59 +0000 |")
	assertContains(t, resStr, "| Количество запросов   |                    10 |")
	assertContains(t, resStr, "| Средний размер ответа |                   500 |")
	assertContains(t, resStr, "| 95p размера ответа    |                   800 |")
	assertContains(t, resStr, "=== Запрашиваемые ресурсы")
	assertContains(t, resStr, "| Ресурс                |            Количество |")
	assertContains(t, resStr, "| /index.html           |                     5 |")
	assertContains(t, resStr, "| /about.html           |                     3 |")
	assertContains(t, resStr, "=== Коды ответа")
	assertContains(t, resStr, "| Код                   |            Количество |")
	assertContains(t, resStr, "| 200                   |                     8 |")
	assertContains(t, resStr, "| 500                   |                     2 |")
	assertContains(t, resStr, "=== Доп. метрики:")
	assertContains(t, resStr, "| Метрика               |              Значение |")
	assertContains(t, resStr, "| Кол-во отказов (5xx)  |                     2 |")
	assertContains(t, resStr, "=== Ссылающиеся ресурсы")
	assertContains(t, resStr, "| Реферер               |            Количество |")
	assertContains(t, resStr, "| https://example.com   |                     7 |")
}

func parseTestTime(value string) time.Time {
	parsed, err := client.ParseDate(value)
	if err != nil {
		panic("Failed to parse time: " + err.Error())
	}

	return parsed
}

func assertContains(t *testing.T, output, substring string) {
	assert.Contains(t, output, substring, "Expected output to contain %q", substring)
}
