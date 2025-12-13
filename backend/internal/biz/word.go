package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// WordUsecase 单词业务逻辑接口
type WordUsecase interface {
	BatchAddWords(ctx context.Context, planID int64, words []*WordItem) ([]*Word, error)
	GetPlanWords(ctx context.Context, planID int64, page, pageSize int32) ([]*Word, int64, error)
	GetTodayWords(ctx context.Context, planID int64, date string) ([]*Word, error)
	GetWord(ctx context.Context, planID int64, wordID int64) (*Word, error)
	MarkWordReviewed(ctx context.Context, planID int64, wordID int64) (*Word, error)
	MarkWordForgotten(ctx context.Context, planID int64, wordID int64) (*Word, error)
	UpdateWordReviewData(ctx context.Context, planID int64, wordID int64, thinkTime, difficulty int32, isRemembered bool) (*Word, error)
	ValidatePlanAccess(ctx context.Context, planID int64, wordID int64) error
	GenerateReviewQuestions(ctx context.Context, planID int64, wordID int64, grade string) ([]*ReviewQuestion, error)
}

// WordItem 单词项
type WordItem struct {
	Word      string
	Meaning   string
	StartDate string
}

type wordUsecase struct {
	wordRepo    WordRepo
	planRepo    PlanRepo
	studentRepo StudentRepo
	llmService  LLMService
	log         *log.Helper
}

// NewWordUsecase 创建单词业务逻辑
func NewWordUsecase(wordRepo WordRepo, planRepo PlanRepo, studentRepo StudentRepo, llmService LLMService, logger log.Logger) WordUsecase {
	return &wordUsecase{
		wordRepo:    wordRepo,
		planRepo:    planRepo,
		studentRepo: studentRepo,
		llmService:  llmService,
		log:         log.NewHelper(logger),
	}
}

// BatchAddWords 批量添加单词到计划
func (uc *wordUsecase) BatchAddWords(ctx context.Context, planID int64, words []*WordItem) ([]*Word, error) {
	// 验证计划是否存在
	_, err := uc.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, errors.New("plan not found")
	}

	// 验证输入
	if len(words) == 0 {
		return nil, errors.New("words list cannot be empty")
	}

	// 转换为数据模型
	dbWords := make([]*Word, 0, len(words))
	for _, item := range words {
		if item.Word == "" || item.Meaning == "" {
			continue // 跳过无效的单词
		}
		dbWords = append(dbWords, &Word{
			PlanID:         planID,
			Word:           item.Word,
			Meaning:        item.Meaning,
			StartDate:      item.StartDate,
			ReviewCount:    0,
			LastReviewDate: "",
		})
	}

	if len(dbWords) == 0 {
		return nil, errors.New("no valid words to add")
	}

	// 批量插入
	if err := uc.wordRepo.BatchCreate(ctx, dbWords); err != nil {
		return nil, err
	}

	return dbWords, nil
}

// GetPlanWords 获取计划的单词列表
func (uc *wordUsecase) GetPlanWords(ctx context.Context, planID int64, page, pageSize int32) ([]*Word, int64, error) {
	// 验证计划是否存在
	_, err := uc.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, 0, errors.New("plan not found")
	}

	// 设置默认值
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100 // 限制最大页大小
	}

	return uc.wordRepo.GetByPlanID(ctx, planID, page, pageSize)
}

// GetTodayWords 获取今日需要背诵的单词（根据艾宾浩斯曲线）
func (uc *wordUsecase) GetTodayWords(ctx context.Context, planID int64, date string) ([]*Word, error) {
	// 验证计划是否存在
	_, err := uc.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, errors.New("plan not found")
	}

	// 解析日期
	var today time.Time
	if date == "" {
		today = time.Now()
	} else {
		parsedDate, err := time.Parse("2006-01-02", date)
		if err != nil {
			return nil, errors.New("invalid date format, expected YYYY-MM-DD")
		}
		today = parsedDate
	}

	// 获取计划中的所有单词
	allWords, err := uc.wordRepo.GetAllByPlanID(ctx, planID)
	if err != nil {
		return nil, err
	}

	// 使用艾宾浩斯算法过滤出今天需要复习的单词
	var todayWords []*Word
	for _, word := range allWords {
		startDate, err := time.Parse("2006-01-02", word.StartDate)
		if err != nil {
			continue
		}

		if ShouldReviewToday(startDate, word.ReviewCount, today) {
			todayWords = append(todayWords, word)
		}
	}

	return todayWords, nil
}

// MarkWordReviewed 标记单词为已复习
func (uc *wordUsecase) MarkWordReviewed(ctx context.Context, planID int64, wordID int64) (*Word, error) {
	// 验证计划是否有权限
	if err := uc.ValidatePlanAccess(ctx, planID, wordID); err != nil {
		return nil, err
	}

	// 获取单词
	word, err := uc.wordRepo.GetByID(ctx, wordID)
	if err != nil {
		return nil, errors.New("word not found")
	}

	// 更新复习次数和最后复习日期
	lastReviewDate := word.LastReviewDate
	word.LastReviewDate = time.Now().Format("2006-01-02")
	if word.LastReviewDate != lastReviewDate {
		word.ReviewCount++
	}
	word.IsRemembered = true

	// 保存更新
	updatedWord, err := uc.wordRepo.Update(ctx, word)
	if err != nil {
		return nil, err
	}

	return updatedWord, nil
}

// MarkWordForgotten 标记单词为未记住
func (uc *wordUsecase) MarkWordForgotten(ctx context.Context, planID int64, wordID int64) (*Word, error) {
	// 验证计划是否有权限
	if err := uc.ValidatePlanAccess(ctx, planID, wordID); err != nil {
		return nil, err
	}

	// 获取单词
	word, err := uc.wordRepo.GetByID(ctx, wordID)
	if err != nil {
		return nil, errors.New("word not found")
	}

	// 更新忘记次数和记住状态
	word.ForgetCount++
	word.IsRemembered = false
	word.LastReviewDate = time.Now().Format("2006-01-02")

	// 保存更新
	updatedWord, err := uc.wordRepo.Update(ctx, word)
	if err != nil {
		return nil, err
	}

	return updatedWord, nil
}

// UpdateWordReviewData 更新单词复习数据（思考时间、难度等）
func (uc *wordUsecase) UpdateWordReviewData(ctx context.Context, planID int64, wordID int64, thinkTime, difficulty int32, isRemembered bool) (*Word, error) {
	// 验证计划是否有权限
	if err := uc.ValidatePlanAccess(ctx, planID, wordID); err != nil {
		return nil, err
	}

	// 获取单词
	word, err := uc.wordRepo.GetByID(ctx, wordID)
	if err != nil {
		return nil, errors.New("word not found")
	}

	// 更新数据
	if thinkTime > 0 {
		word.ThinkTime = thinkTime
	}
	if difficulty >= 0 && difficulty <= 10 {
		word.Difficulty = difficulty
	}
	word.IsRemembered = isRemembered
	word.LastReviewDate = time.Now().Format("2006-01-02")

	// 如果标记为记住，增加复习次数
	if isRemembered {
		word.ReviewCount++
	} else {
		word.ForgetCount++
	}

	// 保存更新
	updatedWord, err := uc.wordRepo.Update(ctx, word)
	if err != nil {
		return nil, err
	}

	return updatedWord, nil
}

// GetWord 获取单词（带权限验证）
func (uc *wordUsecase) GetWord(ctx context.Context, planID int64, wordID int64) (*Word, error) {
	// 验证权限（通过计划验证）
	if err := uc.ValidatePlanAccess(ctx, planID, wordID); err != nil {
		return nil, err
	}

	// 获取单词
	return uc.wordRepo.GetByID(ctx, wordID)
}

// GenerateReviewQuestions 生成复习题目
func (uc *wordUsecase) GenerateReviewQuestions(ctx context.Context, planID int64, wordID int64, grade string) ([]*ReviewQuestion, error) {
	// 验证权限
	if err := uc.ValidatePlanAccess(ctx, planID, wordID); err != nil {
		return nil, err
	}

	// 获取单词
	word, err := uc.wordRepo.GetByID(ctx, wordID)
	if err != nil {
		return nil, errors.New("word not found")
	}

	// 检查LLM服务是否可用
	if uc.llmService == nil {
		return nil, errors.New("llm service not configured")
	}

	// 调用LLM服务生成题目
	questions, err := uc.llmService.GenerateReviewQuestions(ctx, word.Word, word.Meaning, grade)
	if err != nil {
		uc.log.Errorf("failed to generate review questions: %v", err)
		return nil, fmt.Errorf("failed to generate review questions: %w", err)
	}

	return questions, nil
}

// ValidatePlanAccess 验证计划是否有权限访问该单词
func (uc *wordUsecase) ValidatePlanAccess(ctx context.Context, planID int64, wordID int64) error {
	word, err := uc.wordRepo.GetByID(ctx, wordID)
	if err != nil {
		return errors.New("word not found")
	}

	if word.PlanID != planID {
		return errors.New("access denied: word does not belong to this plan")
	}

	return nil
}
