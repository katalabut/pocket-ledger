package email

import (
	"context"
	"log"
)

type LoggingSender struct {
	logger *log.Logger
	from   string
}

func NewLoggingSender(logger *log.Logger, from string) *LoggingSender {
	if logger == nil {
		logger = log.Default()
	}
	return &LoggingSender{logger: logger, from: from}
}

func (s *LoggingSender) Send(_ context.Context, to, subject, body string) error {
	s.logger.Printf("email test mode: from=%q to=%q subject=%q body=%q", s.from, to, subject, body)
	return nil
}
