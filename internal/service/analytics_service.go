package service

import (
	"strconv"
	"sync"
	"time"

	"github.com/4domm/ngxstat/internal/domain"

	"github.com/4domm/ngxstat/internal/infrastructure/parser"
	"github.com/HdrHistogram/hdrhistogram-go"
)

const (
	MaxHistogramValue              = 100000000
	MinHistogramValue              = 0
	NumberOfSignificantValueDigits = 3
	StartServerErrorCode           = 500
	EndServerErrorCode             = 599
	TopN                           = 3
)

type Reader interface {
	ReadLines(*domain.InputConfig) (chan string, error)
}

type AnalyticsService struct {
	AnalysisResult *domain.AnalysisResult
	Histogram      *hdrhistogram.Histogram
	LogParser      parser.LogParser
	Reader         Reader
	mu             sync.Mutex
	filesUsed      map[string]struct{}
}

func NewAnalyticsService(logParser parser.LogParser, readers Reader) *AnalyticsService {
	return &AnalyticsService{
		LogParser:      logParser,
		Reader:         readers,
		Histogram:      hdrhistogram.New(MinHistogramValue, MaxHistogramValue, NumberOfSignificantValueDigits),
		AnalysisResult: domain.NewAnalysisResult(),
		filesUsed:      make(map[string]struct{}),
	}
}

func (s *AnalyticsService) Process(inputConfig *domain.InputConfig) (*domain.AnalysisResult, error) {
	lines, err := s.Reader.ReadLines(inputConfig)
	if err != nil {
		return nil, err
	}

	logData := s.parseAndFilter(lines, inputConfig)
	s.runAnalyticsWorkers(logData)
	s.AnalysisResult.ProcessAll(TopN, s.Histogram, inputConfig.From, inputConfig.To)

	return s.AnalysisResult, nil
}

func (s *AnalyticsService) parseAndFilter(
	lines <-chan string,
	inputConfig *domain.InputConfig,
) <-chan *domain.LogData {
	filterFunction := parser.GetFilterFunction(inputConfig.FilterField, inputConfig.FilterValue)
	logData := make(chan *domain.LogData)

	go func() {
		defer close(logData)

		for line := range lines {
			parsedData, err := s.LogParser.ParseLogLine(line)
			if err == nil && parsedData != nil && s.IsTailoredForTimeRange(parsedData, inputConfig.From, inputConfig.To) {
				if filterFunction == nil {
					logData <- parsedData
				} else if filterFunction(parsedData) {
					logData <- parsedData
				}
			}
		}
	}()

	return logData
}

func (s *AnalyticsService) runAnalyticsWorkers(logData <-chan *domain.LogData) {
	const numWorkers = 8

	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for data := range logData {
				s.mu.Lock()
				s.UpdateAnalytics(data)

				s.UpdateFiles(data.Filename)

				s.mu.Unlock()
			}
		}()
	}

	wg.Wait()
}

func (s *AnalyticsService) IsTailoredForTimeRange(logData *domain.LogData, from, to time.Time) bool {
	if !from.IsZero() && logData.Timestamp.Before(from) {
		return false
	}

	if !to.IsZero() && logData.Timestamp.After(to) {
		return false
	}

	return true
}
func (s *AnalyticsService) UpdateAnalytics(logData *domain.LogData) {
	s.AnalysisResult.TotalResponseSize += logData.ResponseSize
	s.AnalysisResult.TotalRequests++

	if s.IsServerErrorStatus(logData) {
		s.AnalysisResult.TotalServerErrorsLogs++
	}

	if logData.Referer != "" {
		s.AnalysisResult.MostFrequentReferrers[logData.Referer]++
	}

	s.AnalysisResult.MostRequestedResources[logData.Resource]++
	s.AnalysisResult.MostFrequentStatusCodes[logData.StatusCode]++
	err := s.Histogram.RecordValue(logData.ResponseSize)

	if err != nil {
		return
	}
}

func (s *AnalyticsService) UpdateFiles(name string) {
	if _, ok := s.filesUsed[name]; !ok {
		s.AnalysisResult.Filenames = append(s.AnalysisResult.Filenames, name)
		s.filesUsed[name] = struct{}{}
	}
}

func (s *AnalyticsService) IsServerErrorStatus(logData *domain.LogData) bool {
	strStatusCode, _ := strconv.Atoi(logData.StatusCode)
	return strStatusCode >= StartServerErrorCode && strStatusCode <= EndServerErrorCode
}
