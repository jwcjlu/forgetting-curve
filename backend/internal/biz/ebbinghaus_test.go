package biz

import (
	"testing"
	"time"
)

func TestGetDaysBetween(t *testing.T) {
	tests := []struct {
		name     string
		date1    time.Time
		date2    time.Time
		expected int
	}{
		{
			name:     "同一天",
			date1:    time.Date(2024, 1, 1, 10, 30, 0, 0, time.UTC),
			date2:    time.Date(2024, 1, 1, 15, 45, 0, 0, time.UTC),
			expected: 0,
		},
		{
			name:     "相差1天",
			date1:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			date2:    time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			expected: 1,
		},
		{
			name:     "相差7天",
			date1:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			date2:    time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC),
			expected: 7,
		},
		{
			name:     "date2早于date1",
			date1:    time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
			date2:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: -9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetDaysBetween(tt.date1, tt.date2)
			if result != tt.expected {
				t.Errorf("GetDaysBetween() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestShouldReviewToday(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name        string
		startDate   time.Time
		reviewCount int32
		today       time.Time
		expected    bool
	}{
		{
			name:        "开始日期，第0次复习",
			startDate:   baseDate,
			reviewCount: 0,
			today:       baseDate,
			expected:    true,
		},
		{
			name:        "开始日期前，不需要复习",
			startDate:   baseDate,
			reviewCount: 0,
			today:       baseDate.AddDate(0, 0, -1),
			expected:    false,
		},
		{
			name:        "第1天，第0次复习后",
			startDate:   baseDate,
			reviewCount: 0,
			today:       baseDate.AddDate(0, 0, 1),
			expected:    true, // 第1天应该复习
		},
		{
			name:        "第2天，第1次复习后（还没到第2次复习时间）",
			startDate:   baseDate,
			reviewCount: 1,
			today:       baseDate.AddDate(0, 0, 2),
			expected:    false, // 第2天还没到第2次复习时间（应该是第1+2=3天）
		},
		{
			name:        "第3天，第1次复习后",
			startDate:   baseDate,
			reviewCount: 1,
			today:       baseDate.AddDate(0, 0, 3),
			expected:    true, // 第3天应该复习（第1+2=3天）
		},
		{
			name:        "第6天，第2次复习后（还没到第3次复习时间）",
			startDate:   baseDate,
			reviewCount: 2,
			today:       baseDate.AddDate(0, 0, 6),
			expected:    false, // 第6天还没到第3次复习时间（应该是第1+2+4=7天）
		},
		{
			name:        "第7天，第2次复习后",
			startDate:   baseDate,
			reviewCount: 2,
			today:       baseDate.AddDate(0, 0, 7),
			expected:    true, // 第7天应该复习（第1+2+4=7天）
		},
		{
			name:        "第14天，第3次复习后",
			startDate:   baseDate,
			reviewCount: 3,
			today:       baseDate.AddDate(0, 0, 14),
			expected:    true, // 第14天应该复习（第1+2+4+7=14天）
		},
		{
			name:        "超过复习次数，使用最后一个间隔",
			startDate:   baseDate,
			reviewCount: 10,
			// 累计天数 = 1+2+4+7+15+30+30+30+30+30+30 = 209天（前6次用预设间隔，后面都用30）
			today:    baseDate.AddDate(0, 0, 300),
			expected: true, // 超过预设次数，使用最后一个间隔30天，累计天数会很大
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldReviewToday(tt.startDate, tt.reviewCount, tt.today)
			if result != tt.expected {
				t.Errorf("ShouldReviewToday() = %v, want %v (startDate: %v, reviewCount: %d, today: %v)",
					result, tt.expected, tt.startDate, tt.reviewCount, tt.today)
			}
		})
	}
}

func TestFilterTodayWords(t *testing.T) {
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	today := baseDate.AddDate(0, 0, 1) // 第2天

	words := []*WordData{
		{
			ID:          1,
			Word:        "hello",
			Meaning:     "你好",
			StartDate:   "2024-01-01",
			ReviewCount: 0, // 第1天应该复习
		},
		{
			ID:          2,
			Word:        "world",
			Meaning:     "世界",
			StartDate:   "2024-01-01",
			ReviewCount: 1, // 第3天应该复习（1+2=3）
		},
		{
			ID:          3,
			Word:        "test",
			Meaning:     "测试",
			StartDate:   "2024-01-01",
			ReviewCount: 2, // 第7天应该复习（1+2+4=7），第2天还没到
		},
		{
			ID:          4,
			Word:        "future",
			Meaning:     "未来",
			StartDate:   "2024-01-10", // 还没开始
			ReviewCount: 0,
		},
	}

	result := FilterTodayWords(words, today)

	// 第2天，只有hello（第0次复习，第1天）应该复习
	// world需要第3天，test需要第7天
	if len(result) != 1 {
		t.Errorf("FilterTodayWords() returned %d words, want 1", len(result))
	}

	// 检查返回的单词
	foundWords := make(map[string]bool)
	for _, w := range result {
		foundWords[w.Word] = true
	}

	if !foundWords["hello"] {
		t.Error("FilterTodayWords() should include 'hello'")
	}
	if foundWords["world"] {
		t.Error("FilterTodayWords() should not include 'world' (needs day 3)")
	}
	if foundWords["test"] {
		t.Error("FilterTodayWords() should not include 'test' (needs day 7)")
	}
	if foundWords["future"] {
		t.Error("FilterTodayWords() should not include 'future'")
	}
}

func TestReviewIntervals(t *testing.T) {
	// 验证复习间隔数组
	expectedIntervals := []int{1, 2, 4, 7, 15, 30}
	if len(REVIEW_INTERVALS) != len(expectedIntervals) {
		t.Errorf("REVIEW_INTERVALS length = %d, want %d", len(REVIEW_INTERVALS), len(expectedIntervals))
	}

	for i, interval := range REVIEW_INTERVALS {
		if interval != expectedIntervals[i] {
			t.Errorf("REVIEW_INTERVALS[%d] = %d, want %d", i, interval, expectedIntervals[i])
		}
	}
}
