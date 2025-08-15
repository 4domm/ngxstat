package app

import (
	"github.com/4domm/ngxstat/internal/domain"
	"github.com/4domm/ngxstat/internal/infrastructure/generator"
	"github.com/4domm/ngxstat/internal/infrastructure/parser"
	"github.com/4domm/ngxstat/internal/service"
)

type ReportGenerator interface {
	GenerateReport(result *domain.AnalysisResult)

	GenerateExceptionReport(filePath string, message string)
	GetErrorFilePath() string
}

type Application struct {
	InputConfig      *domain.InputConfig
	NginxLogParser   parser.LogParser
	AnalyticsService *service.AnalyticsService
	FileWriter       generator.FileWriter
	Generators       map[string]ReportGenerator
}

func NewApplication(
	generators map[string]ReportGenerator,
	inputConfig *domain.InputConfig,
	nginxParser parser.LogParser,
	analyticsService *service.AnalyticsService,
	writer generator.FileWriter,
) Application {
	return Application{
		Generators:       generators,
		InputConfig:      inputConfig,
		NginxLogParser:   nginxParser,
		AnalyticsService: analyticsService,
		FileWriter:       writer,
	}
}
func (a *Application) Run() {
	reportGenerator := a.Generators[a.InputConfig.OutputFormat]
	res, err := a.AnalyticsService.Process(a.InputConfig)

	if err != nil {
		reportGenerator.GenerateExceptionReport(reportGenerator.GetErrorFilePath(), err.Error())
		return
	}

	reportGenerator.GenerateReport(res)
}
