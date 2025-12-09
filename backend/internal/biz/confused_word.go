package biz

import (
	"context"
	"errors"
	"regexp"

	"github.com/go-kratos/kratos/v2/log"
)

// ConfusedWordRepo 混淆词数据仓库接口
type ConfusedWordRepo interface {
	Create(ctx context.Context, confusedWord *ConfusedWord) error
	GetByWordID(ctx context.Context, wordID int64) ([]*ConfusedWord, error)
	Delete(ctx context.Context, wordID, confusedWordID int64) error
	CheckExists(ctx context.Context, wordID, confusedWordID int64) (bool, error)
}

// ConfusedWordUsecase 混淆词业务逻辑接口
type ConfusedWordUsecase interface {
	AddConfusedWord(ctx context.Context, studentID, wordID, confusedWordID int64) error
	GetConfusedWords(ctx context.Context, studentID, wordID int64) ([]*Word, error)
	RemoveConfusedWord(ctx context.Context, studentID, wordID, confusedWordID int64) error
	SearchWords(ctx context.Context, studentID int64, keyword string, limit int32) ([]*Word, error)
}

type confusedWordUsecase struct {
	confusedWordRepo ConfusedWordRepo
	wordRepo         WordRepo
	studentRepo      StudentRepo
	log              *log.Helper
}

// NewConfusedWordUsecase 创建混淆词业务逻辑
func NewConfusedWordUsecase(
	confusedWordRepo ConfusedWordRepo,
	wordRepo WordRepo,
	studentRepo StudentRepo,
	logger log.Logger,
) ConfusedWordUsecase {
	return &confusedWordUsecase{
		confusedWordRepo: confusedWordRepo,
		wordRepo:         wordRepo,
		studentRepo:      studentRepo,
		log:              log.NewHelper(logger),
	}
}

// AddConfusedWord 添加混淆词
func (uc *confusedWordUsecase) AddConfusedWord(ctx context.Context, studentID, wordID, confusedWordID int64) error {
	// 验证学生权限
	if err := uc.validateStudentAccess(ctx, studentID, wordID); err != nil {
		return err
	}
	if err := uc.validateStudentAccess(ctx, studentID, confusedWordID); err != nil {
		return errors.New("confused word does not belong to this student")
	}

	// 检查是否已存在
	exists, err := uc.confusedWordRepo.CheckExists(ctx, wordID, confusedWordID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("confused word already exists")
	}

	// 不能添加自己作为混淆词
	if wordID == confusedWordID {
		return errors.New("cannot add word as its own confused word")
	}

	// 创建混淆词关联
	confusedWord := &ConfusedWord{
		WordID:         wordID,
		ConfusedWordID: confusedWordID,
	}

	return uc.confusedWordRepo.Create(ctx, confusedWord)
}

// GetConfusedWords 获取单词的混淆词列表
func (uc *confusedWordUsecase) GetConfusedWords(ctx context.Context, studentID, wordID int64) ([]*Word, error) {
	// 验证学生权限
	if err := uc.validateStudentAccess(ctx, studentID, wordID); err != nil {
		return nil, err
	}

	// 获取混淆词关联
	confusedWords, err := uc.confusedWordRepo.GetByWordID(ctx, wordID)
	if err != nil {
		return nil, err
	}

	// 获取混淆词详情
	result := make([]*Word, 0, len(confusedWords))
	for _, cw := range confusedWords {
		word, err := uc.wordRepo.GetByID(ctx, cw.ConfusedWordID)
		if err != nil {
			uc.log.Warnf("failed to get confused word %d: %v", cw.ConfusedWordID, err)
			continue
		}
		result = append(result, word)
	}

	return result, nil
}

// RemoveConfusedWord 删除混淆词
func (uc *confusedWordUsecase) RemoveConfusedWord(ctx context.Context, studentID, wordID, confusedWordID int64) error {
	// 验证学生权限
	if err := uc.validateStudentAccess(ctx, studentID, wordID); err != nil {
		return err
	}

	return uc.confusedWordRepo.Delete(ctx, wordID, confusedWordID)
}

// SearchWords 搜索单词（支持正则表达式）
func (uc *confusedWordUsecase) SearchWords(ctx context.Context, studentID int64, keyword string, limit int32) ([]*Word, error) {
	// 验证学生是否存在
	_, err := uc.studentRepo.GetByID(ctx, studentID)
	if err != nil {
		return nil, errors.New("student not found")
	}

	if keyword == "" {
		return []*Word{}, nil
	}

	// 设置默认限制
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// 获取学生的所有单词
	allWords, _, err := uc.wordRepo.GetByStudentID(ctx, studentID, 1, 10000)
	if err != nil {
		return nil, err
	}

	// 尝试作为正则表达式匹配
	var matchedWords []*Word
	regex, err := regexp.Compile("(?i)" + keyword) // 不区分大小写
	if err != nil {
		// 如果不是有效的正则表达式，使用简单的字符串匹配
		for _, word := range allWords {
			if len(matchedWords) >= int(limit) {
				break
			}
			// 简单的包含匹配
			if containsIgnoreCase(word.Word, keyword) || containsIgnoreCase(word.Meaning, keyword) {
				matchedWords = append(matchedWords, word)
			}
		}
	} else {
		// 使用正则表达式匹配
		for _, word := range allWords {
			if len(matchedWords) >= int(limit) {
				break
			}
			if regex.MatchString(word.Word) || regex.MatchString(word.Meaning) {
				matchedWords = append(matchedWords, word)
			}
		}
	}

	return matchedWords, nil
}

// validateStudentAccess 验证学生是否有权限访问该单词
func (uc *confusedWordUsecase) validateStudentAccess(ctx context.Context, studentID, wordID int64) error {
	word, err := uc.wordRepo.GetByID(ctx, wordID)
	if err != nil {
		return errors.New("word not found")
	}

	if word.StudentID != studentID {
		return errors.New("access denied: word does not belong to this student")
	}

	return nil
}

// containsIgnoreCase 不区分大小写的字符串包含检查
func containsIgnoreCase(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if toLower(s[i+j]) != toLower(substr[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 32
	}
	return c
}
