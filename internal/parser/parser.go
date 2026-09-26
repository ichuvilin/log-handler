package parser

import (
	"errors"
	"regexp"
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
