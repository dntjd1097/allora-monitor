package utils

import (
	"fmt"
	"time"
)

// FormatTime 시간을 지정된 형식으로 포맷팅합니다
func FormatTime(t time.Time, format string) string {
	return t.Format(format)
}

// CurrentTimeFormatted 현재 시간을 지정된 형식으로 포맷팅합니다
func CurrentTimeFormatted(format string) string {
	return FormatTime(time.Now(), format)
}

// LogMessage 로그 메시지를 포맷팅합니다
func LogMessage(level, message string) string {
	timestamp := CurrentTimeFormatted(time.RFC3339)
	return fmt.Sprintf("[%s] %s: %s", timestamp, level, message)
}
