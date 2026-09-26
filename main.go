package main

import (
	"context"
	"fmt"
	"log-handler/internal/cli"
	"log-handler/internal/processor"
	"log-handler/internal/reporter"
	"log-handler/internal/scanner"
	"os"
	"os/signal"
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

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)

	go func() {
		<-sigChan
		fmt.Println("Получен сигнал прерывания...")
		cancel()
	}()

	logs, _ := processor.ProcessFilesConcurrently(ctx, files, 8)
	requests := processor.CorrelateRequests(logs)
	failed := processor.DetectFailedRequests(requests)

	report := reporter.AnalysisResult{
		TotalEntriesProcessed: len(logs),
		FailedRequestsFound:   len(failed),
	}

	err = reporter.WriteJSONReport(report, cfg.OutputFile)
	if err != nil {
		fmt.Printf("error during read log file: %+v", err)
		os.Exit(1)
	}
}
