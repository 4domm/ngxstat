package domain

import (
	"sort"
	"time"

	"github.com/HdrHistogram/hdrhistogram-go"
)

var (
	PERCENTILE = 95.0
)

type AnalysisResult struct {
	MostRequestedResources   map[string]int64
	MostFrequentStatusCodes  map[string]int64
	MostFrequentReferrers    map[string]int64
	TotalResponseSize        int64
	TotalRequests            int64
	TotalServerErrorsLogs    int64
	AverageResponseSize      float64
	Percentile95ResponseSize int64
	From                     time.Time
	To                       time.Time
	Filenames                []string
}

func NewAnalysisResult() *AnalysisResult {
	return &AnalysisResult{
		MostRequestedResources:  make(map[string]int64),
		MostFrequentStatusCodes: make(map[string]int64),
		MostFrequentReferrers:   make(map[string]int64),
	}
}
func (ar *AnalysisResult) ProcessAll(topN int, histogram *hdrhistogram.Histogram, from, to time.Time) {
	ar.CountAverageResponseSize()
	ar.GetTopRequestedResources(topN)
	ar.getTopReferrers(topN)
	ar.From = from
	ar.To = to
	ar.getTopFrequentStatusCodes(topN)
	ar.getPercentile(histogram)
}
func (ar *AnalysisResult) CountAverageResponseSize() {
	if ar.TotalRequests > 0 {
		ar.AverageResponseSize = float64(ar.TotalResponseSize) / float64(ar.TotalRequests)
	}
}

func (ar *AnalysisResult) GetTopRequestedResources(topN int) {
	ar.MostRequestedResources = ar.getTopN(ar.MostRequestedResources, topN)
}

func (ar *AnalysisResult) getTopFrequentStatusCodes(topN int) {
	ar.MostFrequentStatusCodes = ar.getTopN(ar.MostFrequentStatusCodes, topN)
}

func (ar *AnalysisResult) getTopReferrers(topN int) {
	ar.MostFrequentReferrers = ar.getTopN(ar.MostFrequentReferrers, topN)
}

func (ar *AnalysisResult) getTopN(countMap map[string]int64, topN int) map[string]int64 {
	type entry struct {
		Key   string
		Value int64
	}

	entries := make([]entry, 0, len(countMap))

	for k, v := range countMap {
		entries = append(entries, entry{Key: k, Value: v})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Value > entries[j].Value
	})

	topNMap := make(map[string]int64)
	for i := 0; i < topN && i < len(entries); i++ {
		topNMap[entries[i].Key] = entries[i].Value
	}

	return topNMap
}

func (ar *AnalysisResult) getPercentile(histogram *hdrhistogram.Histogram) {
	ar.Percentile95ResponseSize = histogram.ValueAtPercentile(PERCENTILE)
}
