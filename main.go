package main

import (
	"fmt"
	"log-handler/internal/cli"
	"log-handler/internal/processor"
	"log-handler/internal/reporter"
	"log-handler/internal/scanner"
	"os"
)

func main() {
	cfg, err := cli.ParseCommandLineArgs()
	if err != nil {
		fmt.Printf("error during read log file: %+v", err)
		os.Exit(1)
	}

	files, err := scanner.ScanLogDirectory(cfg.OutputFile)
	if err != nil {
		fmt.Printf("error during read log file: %+v", err)
		os.Exit(1)
	}

	for _, file := range files {
		logs, err := processor.ReadLogFile(file)
		if err != nil {
			fmt.Printf("error during read log file: %+v", err)
		} else {
			requests := processor.CorrelateRequests(logs)
			_ = processor.DetectFailedRequests(requests)
		}
	}

	err = reporter.WriteJSONReport(reporter.AnalysisResult{}, cfg.OutputFile)
	if err != nil {
		fmt.Printf("error during read log file: %+v", err)
		os.Exit(1)
	}
}
