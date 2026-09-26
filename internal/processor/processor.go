package processor

import (
	"bufio"
	"fmt"
	"log-handler/internal/parser"
	"os"
	"sort"
)

func ReadLogFile(filePath string) ([]parser.LogEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return []parser.LogEntry{}, err
	}
	defer file.Close()

	entries := make([]parser.LogEntry, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if logLine, err := parser.ParseLogLine(line); err == nil {
			entries = append(entries, logLine)
		} else {
			fmt.Printf("error parse log: %s\n", err)
		}
	}

	return entries, nil
}

func CorrelateRequests(entries []parser.LogEntry) map[string][]parser.LogEntry {
	grouped := make(map[string][]parser.LogEntry)
	withoutReqID := make([]parser.LogEntry, 0)

	for _, entry := range entries {
		if entry.RequestID != "" {
			grouped[entry.RequestID] = append(grouped[entry.RequestID], entry)
		} else {
			withoutReqID = append(withoutReqID, entry)
		}
	}

	return grouped
}

func DetectFailedRequests(correlatedRequests map[string][]parser.LogEntry) []string {
	res := make([]string, 0)
	for k, v := range correlatedRequests {
		_, b := FindFirstFailure(v)
		if b {
			res = append(res, k)
		}
	}

	return res
}

func FindFirstFailure(requestEntries []parser.LogEntry) (parser.LogEntry, bool) {
	res := SortTimelineByTimestamp(requestEntries)

	for _, entry := range res {
		if entry.Level == "WARN" || entry.Level == "ERROR" {
			return entry, true
		}
	}

	return parser.LogEntry{}, false
}

func SortTimelineByTimestamp(entries []parser.LogEntry) []parser.LogEntry {
	res := make([]parser.LogEntry, len(entries))
	copy(res, entries)
	sort.Slice(res, func(i, j int) bool {
		return res[i].Timestamp.Before(res[j].Timestamp)
	})
	return res
}
