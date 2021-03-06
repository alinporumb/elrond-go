package logger

import (
	"strings"
	"time"
)

const msgFixedLength = 40

func displayTime(timestamp int64) string {
	t := time.Unix(0, timestamp)

	return t.Format("2006-01-02 15:04:05.000")
}

func formatMessage(msg string) string {
	numWhiteSpaces := 0
	if len(msg) < msgFixedLength {
		numWhiteSpaces = msgFixedLength - len(msg)
	}

	return msg + strings.Repeat(" ", numWhiteSpaces)
}
