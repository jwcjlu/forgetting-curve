package biz

import (
	"time"
)

// REVIEW_INTERVALS 艾宾浩斯遗忘曲线复习间隔（天数）
var REVIEW_INTERVALS = []int{1, 2, 4, 7, 15, 30}

// GetDaysBetween 计算两个日期之间的天数差
func GetDaysBetween(date1, date2 time.Time) int {
	// 重置时间为 00:00:00
	d1 := time.Date(date1.Year(), date1.Month(), date1.Day(), 0, 0, 0, 0, date1.Location())
	d2 := time.Date(date2.Year(), date2.Month(), date2.Day(), 0, 0, 0, 0, date2.Location())
	diff := d2.Sub(d1)
	return int(diff.Hours() / 24)
}

// ShouldReviewToday 判断单词是否需要在今天复习
func ShouldReviewToday(startDate time.Time, reviewCount int32, today time.Time) bool {
	daysSinceStart := GetDaysBetween(startDate, today)

	// 如果还没到开始日期，不需要复习
	if daysSinceStart < 0 {
		return false
	}

	// 如果是开始日期（第0天），需要第一次学习
	if daysSinceStart == 0 {
		return true
	}

	// 计算下一次应该复习的日期（累计天数）
	cumulativeDays := 0
	for i := 0; i <= int(reviewCount); i++ {
		if i < len(REVIEW_INTERVALS) {
			cumulativeDays += REVIEW_INTERVALS[i]
		} else {
			// 如果超过预设的复习次数，使用最后一个间隔
			cumulativeDays += REVIEW_INTERVALS[len(REVIEW_INTERVALS)-1]
		}
	}

	// 如果今天正好是应该复习的日期，或者已经超过了应该复习的日期（需要补复习）
	if daysSinceStart >= cumulativeDays {
		return true
	}

	return false
}

// FilterTodayWords 过滤出今天需要复习的单词
func FilterTodayWords(words []*WordData, today time.Time) []*WordData {
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())

	var result []*WordData
	for _, word := range words {
		startDate, err := time.Parse("2006-01-02", word.StartDate)
		if err != nil {
			continue
		}

		if ShouldReviewToday(startDate, word.ReviewCount, today) {
			result = append(result, word)
		}
	}

	return result
}

// WordData 单词数据（用于业务逻辑）
type WordData struct {
	ID             int64
	StudentID      int64
	Word           string
	Meaning        string
	StartDate      string
	ReviewCount    int32
	LastReviewDate string
}
