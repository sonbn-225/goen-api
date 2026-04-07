package utils

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrEmptyTime   = errors.New("empty time string")
	ErrInvalidTime = errors.New("invalid time format")
)

// timeFormats được khai báo ngoài hàm để tránh tốn RAM cấp phát lại mỗi lần parse.
var timeFormats = []string{
	time.RFC3339,
	time.RFC3339Nano,
	"2006-01-02T15:04:05Z07:00",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
}

// Now trả về thời gian hiện tại chuẩn UTC.
// Cắt bỏ phần Monotonic Clock (Microsecond) để đồng nhất dữ liệu giữa RAM và Database.
func Now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}

// NowString trả về chuỗi thời gian hiện tại theo chuẩn RFC3339.
func NowString() string {
	return Now().Format(time.RFC3339)
}

// NowDateString trả về chuỗi ngày hiện tại (Ví dụ: "2026-04-07").
func NowDateString() string {
	return Now().Format(time.DateOnly)
}

// ParseTimeOrDate sử dụng Heuristic Parsing để tăng tốc độ phân tích chuỗi thời gian.
func ParseTimeOrDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	l := len(s)
	if l == 0 {
		return time.Time{}, ErrEmptyTime
	}

	// Tối ưu 1: Bắt thẳng định dạng "YYYY-MM-DD" để bỏ qua vòng lặp tốn kém.
	if l == 10 {
		t, err := time.ParseInLocation(time.DateOnly, s, time.UTC)
		if err == nil {
			return t, nil
		}
	}

	// Tối ưu 2: Thử các format phức tạp hơn nếu cần.
	for _, f := range timeFormats {
		if t, err := time.Parse(f, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, ErrInvalidTime
}

// NormalizeOccurredAt gộp ngày và giờ một cách thông minh mà không cần phép nối chuỗi (Zero-Allocation).
func NormalizeOccurredAt(occurredAtStr, dateStr, timeStr *string) (time.Time, string, error) {
	// 1. Nếu có sẵn chuỗi occurredAt, ưu tiên xử lý trước
	if occurredAtStr != nil {
		if s := strings.TrimSpace(*occurredAtStr); s != "" {
			t, err := ParseTimeOrDate(s)
			if err != nil {
				return time.Time{}, "", err
			}
			return t, t.Format(time.DateOnly), nil
		}
	}

	now := Now()

	// 2. Xử lý ghép ngày (dateStr) và giờ (timeStr)
	if dateStr != nil {
		if d := strings.TrimSpace(*dateStr); d != "" {

			// Parse phần ngày (luôn ép về UTC)
			datePart, err := time.ParseInLocation(time.DateOnly, d, time.UTC)
			if err != nil {
				return time.Time{}, "", err
			}

			if timeStr != nil {
				if ts := strings.TrimSpace(*timeStr); ts != "" {

					// Parse phần giờ
					timeOnly, err := time.Parse("15:04:05", ts)
					if err == nil {
						// Tối ưu 3: Dùng phép gán toán học time.Date thay vì nối chuỗi string
						finalTime := time.Date(
							datePart.Year(), datePart.Month(), datePart.Day(),
							timeOnly.Hour(), timeOnly.Minute(), timeOnly.Second(), 0, time.UTC,
						)
						return finalTime, d, nil
					}
				}
			}

			// Nếu không có phần giờ hợp lệ, lấy giờ hiện tại (now) đắp vào
			finalTime := time.Date(
				datePart.Year(), datePart.Month(), datePart.Day(),
				now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), time.UTC,
			)
			return finalTime, d, nil
		}
	}

	// 3. Fallback: Nếu không có bất kỳ data nào, trả về thời điểm hiện hành
	return now, now.Format(time.DateOnly), nil
}
