package main

import (
	"fmt"
	"strings"

	"github.com/4domm/ngxstat/internal/infrastructure/client"
	"github.com/4domm/ngxstat/internal/infrastructure/generator"
	"github.com/4domm/ngxstat/internal/infrastructure/reader"
	"github.com/4domm/ngxstat/internal/service"

	"github.com/4domm/ngxstat/internal/app"
	"github.com/4domm/ngxstat/internal/domain"
	"github.com/4domm/ngxstat/internal/infrastructure/parser"
)

func main() {
	config, err := client.ParseFlags()
	if err != nil {
		fmt.Printf("error %v\n", err)
		return
	}

	nginxParser := parser.NginxParser{}

	var linesReader service.Reader
	if strings.HasPrefix(config.Path, "http://") || strings.HasPrefix(config.Path, "https://") {
		linesReader = &reader.URLReader{}
	} else {
		linesReader = &reader.FileReader{}
	}

	analyticsService := service.NewAnalyticsService(nginxParser, linesReader)
	writer := generator.FileWriter{}
	markdownReportGen := generator.NewMarkdownReportGenerator(writer)
	adocReportGen := generator.NewAdocReportGenerator(writer)
	generators := map[string]app.ReportGenerator{
		domain.ADOC:     adocReportGen,
		domain.MARKDOWN: markdownReportGen,
		"":              markdownReportGen,
	}
	application := app.NewApplication(generators, config, nginxParser, analyticsService, writer)

	application.Run()
}
