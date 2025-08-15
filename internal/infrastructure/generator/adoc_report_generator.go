package generator

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"github.com/4domm/ngxstat/internal/domain"
	"github.com/4domm/ngxstat/internal/infrastructure/parser"
)

type AdocReportGenerator struct {
	writer FileWriter
}

func NewAdocReportGenerator(writer FileWriter) *AdocReportGenerator {
	return &AdocReportGenerator{writer: writer}
}
func (arg *AdocReportGenerator) GenerateReport(result *domain.AnalysisResult) {
	file, err := arg.writer.CreateFile(arg.GetFilePath())
	if err != nil {
		arg.GenerateExceptionReport(arg.GetErrorFilePath(), "Ошибка записи в файл")
		return
	}

	defer file.Close()

	writer := bufio.NewWriter(file)

	defer writer.Flush()

	arg.writeGeneralInfo(writer, result)
	arg.writeRequestedResources(writer, result)
	arg.writeResponseCodes(writer, result)
	arg.writeAdditionalInfo(writer, result)
}

func (arg *AdocReportGenerator) writeGeneralInfo(writer *bufio.Writer, result *domain.AnalysisResult) {
	arg.writeLine(writer, arg.getGeneralInfoHeader())
	arg.writeLine(writer, arg.formatLine("Метрика", "Значение"))
	arg.writeLine(writer, arg.formatLine("---------------------", "---------------------"))

	arg.writeLine(writer, arg.formatLine("Количество файлов", len(result.Filenames)))
	arg.writeLine(writer, arg.getListFiles(result.Filenames))

	arg.writeLine(writer, arg.formatLine("Начальная дата", arg.emptyTimeConverter(result.From)))
	arg.writeLine(writer, arg.formatLine("Конечная дата", arg.emptyTimeConverter(result.To)))
	arg.writeLine(writer, arg.formatLine("Количество запросов", result.TotalRequests))
	arg.writeLine(writer, arg.formatLine("Средний размер ответа", result.AverageResponseSize))
	arg.writeLine(writer, arg.formatLine("95p размера ответа", result.Percentile95ResponseSize))
	arg.writeLine(writer, "")
}

func (arg *AdocReportGenerator) writeRequestedResources(writer *bufio.Writer, result *domain.AnalysisResult) {
	arg.writeLine(writer, arg.getRequestedResourcesHeader())
	arg.writeLine(writer, arg.formatLine("Ресурс", "Количество"))
	arg.writeLine(writer, arg.formatLine("---------------------", "---------------------"))

	for key, value := range result.MostRequestedResources {
		arg.writeLine(writer, arg.formatLine(key, value))
	}

	arg.writeLine(writer, "")
}

func (arg *AdocReportGenerator) writeResponseCodes(writer *bufio.Writer, result *domain.AnalysisResult) {
	arg.writeLine(writer, arg.getResponseCodesHeader())
	arg.writeLine(writer, arg.formatLine("Код", "Количество"))
	arg.writeLine(writer, arg.formatLine("---------------------", "---------------------"))

	for key, value := range result.MostFrequentStatusCodes {
		arg.writeLine(writer, arg.formatLine(key, value))
	}

	arg.writeLine(writer, "")
}

func (arg *AdocReportGenerator) writeAdditionalInfo(writer *bufio.Writer, result *domain.AnalysisResult) {
	arg.writeLine(writer, arg.getAdditionalInfoHeader())
	arg.writeLine(writer, arg.formatLine("Метрика", "Значение"))
	arg.writeLine(writer, arg.formatLine("---------------------", "---------------------"))
	arg.writeLine(writer, arg.formatLine("Кол-во отказов (5xx)", result.TotalServerErrorsLogs))
	arg.writeLine(writer, "")

	arg.writeMostFrequentReferrers(writer, result)
}

func (arg *AdocReportGenerator) writeMostFrequentReferrers(writer *bufio.Writer, result *domain.AnalysisResult) {
	arg.writeLine(writer, arg.getRefereesHeader())
	arg.writeLine(writer, arg.formatLine("Реферер", "Количество"))
	arg.writeLine(writer, arg.formatLine("---------------------", "---------------------"))

	for key, value := range result.MostFrequentReferrers {
		arg.writeLine(writer, arg.formatLine(key, value))
	}

	arg.writeLine(writer, "")
}

func (arg *AdocReportGenerator) GenerateExceptionReport(filePath, message string) {
	file, err := arg.writer.CreateFile(filePath)
	if err != nil {
		fmt.Printf("Ошибка при записи в файл об ошибке: %s\n", err.Error())
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	arg.writeLine(writer, arg.getExceptionHeader())
	arg.writeLine(writer, message)
}

func (arg *AdocReportGenerator) getListFiles(filenames []string) string {
	var builder strings.Builder
	for _, filename := range filenames {
		builder.WriteString(arg.formatLine("Файл:", filename) + "\n")
	}

	return builder.String()[:len(builder.String())-1]
}

func (arg *AdocReportGenerator) formatLine(paramName string, value interface{}) string {
	return fmt.Sprintf("| %-21s | %21v |\n", paramName, value)
}

func (arg *AdocReportGenerator) writeLine(writer *bufio.Writer, line string) {
	_, _ = writer.WriteString(line + "\n")
}

func (arg *AdocReportGenerator) emptyTimeConverter(value time.Time) string {
	if value.IsZero() {
		return "-"
	}

	return value.Format(parser.NginxDateFormat)
}

func (arg *AdocReportGenerator) GetFilePath() string {
	return "report.adoc"
}

func (arg *AdocReportGenerator) GetErrorFilePath() string {
	return "error.adoc"
}
func (arg *AdocReportGenerator) getGeneralInfoHeader() string {
	return "=== Общая информация\n\n"
}
func (arg *AdocReportGenerator) getRequestedResourcesHeader() string {
	return "=== Запрашиваемые ресурсы\n\n"
}

func (arg *AdocReportGenerator) getResponseCodesHeader() string {
	return "=== Коды ответа\n\n"
}

func (arg *AdocReportGenerator) getExceptionHeader() string {
	return "== Произошла ошибка\n\n"
}

func (arg *AdocReportGenerator) getRefereesHeader() string {
	return "=== Ссылающиеся ресурсы\n\n"
}

func (arg *AdocReportGenerator) getAdditionalInfoHeader() string {
	return "=== Доп. метрики: \n\n"
}
