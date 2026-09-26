package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

type LogEntry struct {
	Timestamp time.Time
	Level     string
	Service   string
	Message   string
	RequestID string
	UserID    string
}

type AnalysisResult struct {
	TotalEntriesProcessed int                   `json:"total_entries_processed"`
	FailedRequestsFound   int                   `json:"failed_requests_found"`
	ProcessingTimeSeconds float64               `json:"processing_time_seconds"`
	FailedRequests        []FailedRequestReport `json:"failed_requests"`
}

type Config struct {
	InputDir   string
	OutputFile string
}

type FailedRequestReport struct {
	RequestID      string   `json:"request_id"`
	FailingService string   `json:"failing_service"`
	ErrorMessage   string   `json:"error_message"`
	Timeline       []string `json:"timeline"`
}

func ParseLogLine(line string) (LogEntry, error) {
	log := LogEntry{}

	iso8601Regex := regexp.MustCompile(
		`^(\S+)`,
	)
	t := iso8601Regex.FindString(line)
	if t != "" {
		parse, err := time.Parse(time.RFC3339, t)
		if err != nil {
			return LogEntry{}, errors.New("invalid log: timestamp not found")
		}
		log.Timestamp = parse
	} else {
		return LogEntry{}, errors.New("invalid log: timestamp not found")
	}

	re := regexp.MustCompile(`\[([A-Z]+)]`)
	levels := re.FindStringSubmatch(line)
	if len(levels) > 1 {
		log.Level = levels[1]
	} else {
		return LogEntry{}, errors.New("invalid log: level not found")
	}

	re = regexp.MustCompile("([a-z-]+):")
	srv := re.FindStringSubmatch(line)
	if len(srv) > 1 {
		log.Service = srv[1]
	} else {
		return LogEntry{}, errors.New("invalid log: service not found")
	}

	re = regexp.MustCompile("[a-z-]+: (.*)")
	msg := re.FindStringSubmatch(line)
	if len(msg) > 1 {
		log.Message = msg[1]
	} else {
		return LogEntry{}, errors.New("invalid log: massage not found")
	}

	re = regexp.MustCompile(`request_id=([a-zA-Z0-9_]+)`)
	matches := re.FindStringSubmatch(line)

	if len(matches) > 1 {
		log.RequestID = matches[1]
	} else {
		return LogEntry{}, errors.New("invalid log: reqID not found")
	}

	re = regexp.MustCompile(`user_id=([0-9]+)`)
	matches = re.FindStringSubmatch(line)

	if len(matches) > 1 {
		log.UserID = matches[1]
	} else {
		return LogEntry{}, errors.New("invalid log: user id not found")
	}

	return log, nil
}

func ReadLogFile(filePath string) ([]LogEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return []LogEntry{}, err
	}
	defer file.Close()

	entries := make([]LogEntry, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if logLine, err := ParseLogLine(line); err == nil {
			entries = append(entries, logLine)
		} else {
			fmt.Printf("error parse log: %s\n", err)
		}
	}

	return entries, nil
}

func ScanLogDirectory(dirPath string) ([]string, error) {
	paths := make([]string, 0)
	if err := filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".log" {
			paths = append(paths, path)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return paths, nil
}

func ProcessMultipleFiles(filePaths []string) ([]LogEntry, error) {
	entries := make([]LogEntry, 0)

	for _, path := range filePaths {
		if logs, err := ReadLogFile(path); err != nil {
			fmt.Printf("error parse log by path %s: %v\n", path, err)
		} else {
			entries = append(entries, logs...)
		}
	}

	return entries, nil
}

func CorrelateRequests(entries []LogEntry) map[string][]LogEntry {
	grouped := make(map[string][]LogEntry)
	withoutReqID := make([]LogEntry, 0)

	for _, entry := range entries {
		if entry.RequestID != "" {
			grouped[entry.RequestID] = append(grouped[entry.RequestID], entry)
		} else {
			withoutReqID = append(withoutReqID, entry)
		}
	}

	return grouped
}

func DetectFailedRequests(correlatedRequests map[string][]LogEntry) []string {
	res := make([]string, 0)
	for k, v := range correlatedRequests {
		_, b := FindFirstFailure(v)
		if b {
			res = append(res, k)
		}
	}

	return res
}

func FindFirstFailure(requestEntries []LogEntry) (LogEntry, bool) {
	res := SortTimelineByTimestamp(requestEntries)

	for _, entry := range res {
		if entry.Level == "WARN" || entry.Level == "ERROR" {
			return entry, true
		}
	}

	return LogEntry{}, false
}

func SortTimelineByTimestamp(entries []LogEntry) []LogEntry {
	res := make([]LogEntry, len(entries))
	copy(res, entries)
	sort.Slice(res, func(i, j int) bool {
		return res[i].Timestamp.Before(res[j].Timestamp)
	})
	return res
}

func WriteJSONReport(result AnalysisResult, filename string) error {
	data, err := json.MarshalIndent(result, "", " ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func ParseCommandLineArgs() (Config, error) {
	inputDir := flag.String("input-dir", ".", "Directory containing .log files")
	outputFile := flag.String("output-file", "result.json", "JSON output file path")

	flag.Parse()

	_, err := os.Stat(*inputDir)
	if err != nil {
		return Config{}, err
	}

	return Config{InputDir: *inputDir, OutputFile: *outputFile}, err
}

func main() {
	line, _ := ParseLogLine("2023-12-25T14:30:15.123Z [INFO] user-service: User authenticated, request_id=req_abc123, user_id=12345")
	fmt.Println(fmt.Sprintf("%+v", line))
}
