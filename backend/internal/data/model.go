package data

import (
	"time"

	"gorm.io/gorm"
)

// Student 学生模型
type Student struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"`
	StudentNo string         `gorm:"type:varchar(50);uniqueIndex;default:''" json:"student_no"`
	OpenID    string         `gorm:"type:varchar(100);uniqueIndex;column:open_id" json:"openid"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Words     []Word         `gorm:"foreignKey:StudentID" json:"words,omitempty"`
}

// Word 单词模型
type Word struct {
	ID             int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	StudentID      int64          `gorm:"type:bigint;not null;index" json:"student_id"`
	Word           string         `gorm:"type:varchar(100);not null" json:"word"`
	Meaning        string         `gorm:"type:varchar(500);not null" json:"meaning"`
	StartDate      string         `gorm:"type:varchar(20);not null" json:"start_date"`
	ReviewCount    int32          `gorm:"type:int;default:0" json:"review_count"`
	LastReviewDate string         `gorm:"type:varchar(20)" json:"last_review_date"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	Student        Student        `gorm:"foreignKey:StudentID" json:"student,omitempty"`
}

// TableName 指定表名
func (Student) TableName() string {
	return "students"
}

func (Word) TableName() string {
	return "words"
}
