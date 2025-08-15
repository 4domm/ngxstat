package service_test

import (
	"testing"
	"time"

	"github.com/4domm/ngxstat/internal/infrastructure/client"

	"github.com/4domm/ngxstat/internal/domain"
	"github.com/4domm/ngxstat/internal/infrastructure/parser"
	"github.com/4domm/ngxstat/internal/service"
	"github.com/stretchr/testify/assert"
)

func TestIsTailoredForTimeRange(t *testing.T) {
	analyticsService := &service.AnalyticsService{}
	tests := []struct {
		name     string
		logData  *domain.LogData
		from     time.Time
		to       time.Time
		expected bool
	}{
		{
			name: "Timestamp within range",
			logData: &domain.LogData{
				Timestamp: parser.ParseTimestamp("17/May/2015:22:05:06 +0000"),
			},
			from:     mustParseDate("2015-01-16"),
			to:       mustParseDate("2016-10-18T09:00:00+0000"),
			expected: true,
		},
		{
			name: "Timestamp before range",
			logData: &domain.LogData{
				Timestamp: parser.ParseTimestamp("17/May/2015:07:59:59 +0000"),
			},
			from:     mustParseDate("2015-05-17T08:00:00-0000"),
			to:       mustParseDate("2015-05-17T09:00:00-0000"),
			expected: false,
		},
		{
			name: "Timestamp after range",
			logData: &domain.LogData{
				Timestamp: parser.ParseTimestamp("17/May/2015:09:00:01 +0000"),
			},
			from:     mustParseDate("2015-05-17T08:00:00-0000"),
			to:       mustParseDate("2015-05-17T09:00:00-0000"),
			expected: false,
		},
		{
			name: "No 'from' range, timestamp before 'to'",
			logData: &domain.LogData{
				Timestamp: parser.ParseTimestamp("17/May/2015:08:30:00 +0000"),
			},
			from:     time.Time{},
			to:       mustParseDate("2015-05-17T09:00:00-0000"),
			expected: true,
		},
		{
			name: "No 'to' range, timestamp after 'from'",
			logData: &domain.LogData{
				Timestamp: parser.ParseTimestamp("17/May/2015:08:30:00 +0000"),
			},
			from:     mustParseDate("2015-05-17T08:00:00-0000"),
			to:       time.Time{},
			expected: true,
		},
		{
			name: "No range limits, always true",
			logData: &domain.LogData{
				Timestamp: parser.ParseTimestamp("17/May/2015:08:30:00 +0000"),
			},
			from:     time.Time{},
			to:       time.Time{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyticsService.IsTailoredForTimeRange(tt.logData, tt.from, tt.to)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func mustParseDate(value string) time.Time {
	t, err := client.ParseDate(value)
	if err != nil {
		panic("Failed to parse timestamp in test: " + err.Error())
	}

	return t
}

func TestAnalyticsService_UpdateAnalytics(t *testing.T) {
	analyticsService := service.NewAnalyticsService(nil, nil)
	logData := []*domain.LogData{
		{
			Timestamp:    parser.ParseTimestamp("2023-01-01T00:00:00+0000"),
			ResponseSize: 500,
			Referer:      "https://example.com",
			Resource:     "/index.html",
			StatusCode:   "200",
		},
		{
			Timestamp:    parser.ParseTimestamp("2023-01-01T00:01:00+0000"),
			ResponseSize: 700,
			Referer:      "https://example.com",
			Resource:     "/about.html",
			StatusCode:   "500",
		},
	}

	expectedResult := &domain.AnalysisResult{
		TotalRequests:            2,
		TotalResponseSize:        1200,
		TotalServerErrorsLogs:    1,
		MostRequestedResources:   map[string]int64{"/index.html": 1, "/about.html": 1},
		MostFrequentReferrers:    map[string]int64{"https://example.com": 2},
		MostFrequentStatusCodes:  map[string]int64{"200": 1, "500": 1},
		Percentile95ResponseSize: 700,
		AverageResponseSize:      600,
	}

	for _, log := range logData {
		analyticsService.UpdateAnalytics(log)
	}

	analyticsService.AnalysisResult.ProcessAll(3, analyticsService.Histogram, time.Time{}, time.Time{})

	assertAnalysisResult(t, expectedResult, analyticsService.AnalysisResult)
}

func assertAnalysisResult(t *testing.T, expected, actual *domain.AnalysisResult) {
	assert.Equal(t, expected.TotalRequests, actual.TotalRequests, "TotalRequests mismatch")
	assert.Equal(t, expected.TotalResponseSize, actual.TotalResponseSize, "TotalResponseSize mismatch")
	assert.Equal(t, expected.TotalServerErrorsLogs, actual.TotalServerErrorsLogs, "TotalServerErrorsLogs mismatch")
	assert.Equal(t, expected.AverageResponseSize, actual.AverageResponseSize, "AverageResponseSize mismatch")
	assert.Equal(t, expected.Percentile95ResponseSize, actual.Percentile95ResponseSize, "Percentile95ResponseSize mismatch")

	assertTopNMaps(t, "MostRequestedResources", expected.MostRequestedResources, actual.MostRequestedResources)
	assertTopNMaps(t, "MostFrequentReferrers", expected.MostFrequentReferrers, actual.MostFrequentReferrers)
	assertTopNMaps(t, "MostFrequentStatusCodes", expected.MostFrequentStatusCodes, actual.MostFrequentStatusCodes)
}

func assertTopNMaps(t *testing.T, name string, expected, actual map[string]int64) {
	assert.Equal(t, len(expected), len(actual), "%s size mismatch", name)

	for key, expectedValue := range expected {
		actualValue, ok := actual[key]
		assert.True(t, ok, "Expected key %s in %s not found", key, name)
		assert.Equal(t, expectedValue, actualValue, "Value mismatch for key %s in %s", key, name)
	}
}
