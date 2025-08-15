package generator

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"github.com/4domm/ngxstat/internal/domain"
	"github.com/4domm/ngxstat/internal/infrastructure/parser"
)

type MarkdownReportGenerator struct {
	writer FileWriter
}

func NewMarkdownReportGenerator(writer FileWriter) *MarkdownReportGenerator {
	return &MarkdownReportGenerator{writer: writer}
}
func (mrg MarkdownReportGenerator) GenerateReport(result *domain.AnalysisResult) {
	file, err := mrg.writer.CreateFile(mrg.GetFilePath())
	if err != nil {
		mrg.GenerateExceptionReport(mrg.GetErrorFilePath(), "Error writing to file")
		return
	}

	defer file.Close()

	writer := bufio.NewWriter(file)

	defer writer.Flush()

	mrg.writeGeneralInfo(writer, result)
	mrg.writeRequestedResources(writer, result)
	mrg.writeResponseCodes(writer, result)
	mrg.writeAdditionalInfo(writer, result)
}

func (mrg MarkdownReportGenerator) writeGeneralInfo(writer *bufio.Writer, result *domain.AnalysisResult) {
	mrg.writeLine(writer, mrg.getGeneralInfoHeader())
	mrg.writeLine(writer, mrg.formatLine("Metric", "Value"))
	mrg.writeLine(writer, mrg.formatLine("---", "---"))
	mrg.writeLine(writer, mrg.formatLine("Number of Files", len(result.Filenames)))
	mrg.writeLine(writer, mrg.getListFiles(result.Filenames))
	mrg.writeLine(writer, mrg.formatLine("Start Date", mrg.emptyTimeConverter(result.From)))
	mrg.writeLine(writer, mrg.formatLine("End Date", mrg.emptyTimeConverter(result.To)))
	mrg.writeLine(writer, mrg.formatLine("Total Requests", result.TotalRequests))
	mrg.writeLine(writer, mrg.formatLine("Average Response Size", result.AverageResponseSize))
	mrg.writeLine(writer, mrg.formatLine("95th Percentile Response Size", result.Percentile95ResponseSize))
	mrg.writeLine(writer, "")
}

func (mrg MarkdownReportGenerator) writeRequestedResources(writer *bufio.Writer, result *domain.AnalysisResult) {
	mrg.writeLine(writer, mrg.getRequestedResourcesHeader())
	mrg.writeLine(writer, mrg.formatLine("Resource", "Count"))
	mrg.writeLine(writer, mrg.formatLine("---", "---"))

	for key, value := range result.MostRequestedResources {
		mrg.writeLine(writer, mrg.formatLine(key, value))
	}

	mrg.writeLine(writer, "")
}

func (mrg MarkdownReportGenerator) writeResponseCodes(writer *bufio.Writer, result *domain.AnalysisResult) {
	mrg.writeLine(writer, mrg.getResponseCodesHeader())
	mrg.writeLine(writer, mrg.formatLine("Code", "Count"))
	mrg.writeLine(writer, mrg.formatLine("---", "---"))

	for key, value := range result.MostFrequentStatusCodes {
		mrg.writeLine(writer, mrg.formatLine(key, value))
	}

	mrg.writeLine(writer, "")
}

func (mrg MarkdownReportGenerator) getListFiles(filenames []string) string {
	var builder strings.Builder
	for _, filename := range filenames {
		builder.WriteString(mrg.formatLine("- File", filename) + "\n")
	}

	return builder.String()[:len(builder.String())-1]
}

func (mrg MarkdownReportGenerator) writeAdditionalInfo(writer *bufio.Writer, result *domain.AnalysisResult) {
	mrg.writeLine(writer, mrg.getAdditionalInfoHeader())
	mrg.writeLine(writer, mrg.formatLine("Metric", "Value"))
	mrg.writeLine(writer, mrg.formatLine("---", "---"))
	mrg.writeLine(writer, mrg.formatLine("Server Errors (5xx)", result.TotalServerErrorsLogs))
	mrg.writeLine(writer, "")

	mrg.writeMostFrequentReferrers(writer, result)
}

func (mrg MarkdownReportGenerator) writeMostFrequentReferrers(writer *bufio.Writer, result *domain.AnalysisResult) {
	mrg.writeLine(writer, mrg.getReferrersHeader())
	mrg.writeLine(writer, mrg.formatLine("Referrer", "Count"))
	mrg.writeLine(writer, mrg.formatLine("---", "---"))

	for key, value := range result.MostFrequentReferrers {
		mrg.writeLine(writer, mrg.formatLine(key, value))
	}

	mrg.writeLine(writer, "")
}

func (mrg MarkdownReportGenerator) GenerateExceptionReport(filePath, message string) {
	file, err := mrg.writer.CreateFile(filePath)
	if err != nil {
		fmt.Printf("Error writing exception report: %s\n", err.Error())
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	mrg.writeLine(writer, mrg.getExceptionHeader())
	mrg.writeLine(writer, message)
}

func (mrg MarkdownReportGenerator) formatLine(paramName string, value interface{}) string {
	return fmt.Sprintf("| %-21s | %21v |", paramName, value)
}

func (mrg MarkdownReportGenerator) writeLine(writer *bufio.Writer, line string) {
	_, _ = writer.WriteString(line + "\n")
}

func (mrg MarkdownReportGenerator) emptyTimeConverter(value time.Time) string {
	if value.IsZero() {
		return "-"
	}

	return value.Format(parser.NginxDateFormat)
}

func (mrg MarkdownReportGenerator) GetFilePath() string {
	return "report.md"
}

func (mrg MarkdownReportGenerator) GetErrorFilePath() string {
	return "error.md"
}

func (mrg MarkdownReportGenerator) getGeneralInfoHeader() string {
	return "#### Общая информация \n"
}

func (mrg MarkdownReportGenerator) getRequestedResourcesHeader() string {
	return "#### Запрашиваемые ресурсы\n\n"
}

func (mrg MarkdownReportGenerator) getResponseCodesHeader() string {
	return "\n#### Коды ответа\n"
}

func (mrg MarkdownReportGenerator) getExceptionHeader() string {
	return "### Произошла ошибка: "
}

func (mrg MarkdownReportGenerator) getReferrersHeader() string {
	return "### Ссылающиеся ресурсы\n\n"
}

func (mrg MarkdownReportGenerator) getAdditionalInfoHeader() string {
	return "### Дополнительная информация: \n"
}
