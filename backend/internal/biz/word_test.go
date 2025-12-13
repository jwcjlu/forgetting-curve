package biz

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// mockWordRepo 模拟单词仓库
type mockWordRepo struct {
	words      []*Word
	createErr  error
	getErr     error
	updateErr  error
	updateWord *Word
}

func (m *mockWordRepo) BatchCreate(ctx context.Context, words []*Word) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.words = append(m.words, words...)
	return nil
}

func (m *mockWordRepo) GetByStudentID(ctx context.Context, studentID int64, page, pageSize int32) ([]*Word, int64, error) {
	if m.getErr != nil {
		return nil, 0, m.getErr
	}
	var result []*Word
	for _, w := range m.words {
		if w.StudentID == studentID {
			result = append(result, w)
		}
	}
	total := int64(len(result))
	return result, total, nil
}

func (m *mockWordRepo) GetByID(ctx context.Context, id int64) (*Word, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, w := range m.words {
		if w.ID == id {
			return w, nil
		}
	}
	return nil, errors.New("word not found")
}

func (m *mockWordRepo) Update(ctx context.Context, word *Word) (*Word, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	if m.updateWord != nil {
		return m.updateWord, nil
	}
	return word, nil
}

// mockStudentRepo 模拟学生仓库
type mockStudentRepo struct {
	students map[int64]*Student
	getErr   error
}

func (m *mockStudentRepo) Create(ctx context.Context, student *Student) (*Student, error) {
	if m.students == nil {
		m.students = make(map[int64]*Student)
	}
	m.students[student.ID] = student
	return student, nil
}

func (m *mockStudentRepo) GetByID(ctx context.Context, id int64) (*Student, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	student, ok := m.students[id]
	if !ok {
		return nil, errors.New("student not found")
	}
	return student, nil
}

func (m *mockStudentRepo) GetByStudentNo(ctx context.Context, studentNo string) (*Student, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, s := range m.students {
		if s.StudentNo == studentNo {
			return s, nil
		}
	}
	return nil, errors.New("student not found")
}

func (m *mockStudentRepo) GetByOpenID(ctx context.Context, openid string) (*Student, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, s := range m.students {
		if s.OpenID == openid {
			return s, nil
		}
	}
	return nil, errors.New("student not found")
}

func (m *mockStudentRepo) Update(ctx context.Context, student *Student) (*Student, error) {
	if m.students == nil {
		m.students = make(map[int64]*Student)
	}
	m.students[student.ID] = student
	return student, nil
}

// mockLLMService 模拟LLM服务
type mockLLMService struct {
	questions []*ReviewQuestion
	err       error
}

func (m *mockLLMService) GenerateReviewQuestions(ctx context.Context, word, meaning, grade string) ([]*ReviewQuestion, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.questions, nil
}

func TestWordUsecase_BatchAddWords(t *testing.T) {
	logger := log.NewStdLogger(nil)
	studentRepo := &mockStudentRepo{
		students: map[int64]*Student{
			1: {ID: 1, Name: "Test Student"},
		},
	}
	wordRepo := &mockWordRepo{}
	llmService := &mockLLMService{}

	uc := NewWordUsecase(wordRepo, studentRepo, llmService, logger)

	tests := []struct {
		name      string
		studentID int64
		words     []*WordItem
		wantErr   bool
	}{
		{
			name:      "正常添加",
			studentID: 1,
			words: []*WordItem{
				{Word: "hello", Meaning: "你好", StartDate: "2024-01-01"},
				{Word: "world", Meaning: "世界", StartDate: "2024-01-01"},
			},
			wantErr: false,
		},
		{
			name:      "学生不存在",
			studentID: 999,
			words: []*WordItem{
				{Word: "hello", Meaning: "你好", StartDate: "2024-01-01"},
			},
			wantErr: true,
		},
		{
			name:      "空单词列表",
			studentID: 1,
			words:     []*WordItem{},
			wantErr:   true,
		},
		{
			name:      "包含无效单词",
			studentID: 1,
			words: []*WordItem{
				{Word: "hello", Meaning: "你好", StartDate: "2024-01-01"},
				{Word: "", Meaning: "空单词", StartDate: "2024-01-01"},   // 无效
				{Word: "world", Meaning: "", StartDate: "2024-01-01"}, // 无效
			},
			wantErr: false, // 会跳过无效单词，但至少有一个有效单词
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wordRepo.words = []*Word{} // 重置
			ctx := context.Background()
			_, err := uc.BatchAddWords(ctx, tt.studentID, tt.words)

			if (err != nil) != tt.wantErr {
				t.Errorf("BatchAddWords() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(wordRepo.words) == 0 {
				t.Error("BatchAddWords() should create words")
			}
		})
	}
}

func TestWordUsecase_ValidateStudentAccess(t *testing.T) {
	logger := log.NewStdLogger(nil)
	studentRepo := &mockStudentRepo{}
	wordRepo := &mockWordRepo{
		words: []*Word{
			{ID: 1, StudentID: 1, Word: "hello", Meaning: "你好"},
			{ID: 2, StudentID: 2, Word: "world", Meaning: "世界"},
		},
	}
	llmService := &mockLLMService{}

	uc := NewWordUsecase(wordRepo, studentRepo, llmService, logger)

	tests := []struct {
		name      string
		studentID int64
		wordID    int64
		wantErr   bool
	}{
		{
			name:      "有权限访问",
			studentID: 1,
			wordID:    1,
			wantErr:   false,
		},
		{
			name:      "无权限访问",
			studentID: 1,
			wordID:    2,
			wantErr:   true,
		},
		{
			name:      "单词不存在",
			studentID: 1,
			wordID:    999,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := uc.ValidateStudentAccess(ctx, tt.studentID, tt.wordID)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateStudentAccess() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWordUsecase_MarkWordReviewed(t *testing.T) {
	logger := log.NewStdLogger(nil)
	studentRepo := &mockStudentRepo{}
	wordRepo := &mockWordRepo{
		words: []*Word{
			{
				ID:          1,
				StudentID:   1,
				Word:        "hello",
				Meaning:     "你好",
				ReviewCount: 0,
			},
		},
	}
	llmService := &mockLLMService{}

	uc := NewWordUsecase(wordRepo, studentRepo, llmService, logger)

	ctx := context.Background()
	word, err := uc.MarkWordReviewed(ctx, 1, 1)

	if err != nil {
		t.Fatalf("MarkWordReviewed() error = %v", err)
	}

	if word.ReviewCount != 1 {
		t.Errorf("MarkWordReviewed() ReviewCount = %d, want 1", word.ReviewCount)
	}

	if !word.IsRemembered {
		t.Error("MarkWordReviewed() IsRemembered = false, want true")
	}

	today := time.Now().Format("2006-01-02")
	if word.LastReviewDate != today {
		t.Errorf("MarkWordReviewed() LastReviewDate = %s, want %s", word.LastReviewDate, today)
	}
}

func TestWordUsecase_GenerateReviewQuestions(t *testing.T) {
	logger := log.NewStdLogger(nil)
	studentRepo := &mockStudentRepo{}
	wordRepo := &mockWordRepo{
		words: []*Word{
			{
				ID:        1,
				StudentID: 1,
				Word:      "hello",
				Meaning:   "你好",
			},
		},
	}
	llmService := &mockLLMService{
		questions: []*ReviewQuestion{
			{
				Type:          "multiple_choice",
				Question:      "哪个单词的意思是'你好'？",
				Options:       []string{"hello", "world", "test", "good"},
				CorrectAnswer: "hello",
			},
			{
				Type:          "fill_blank",
				Question:      "请填入单词：____ means hello in Chinese.",
				CorrectAnswer: "hello",
			},
		},
	}

	uc := NewWordUsecase(wordRepo, studentRepo, llmService, logger)

	ctx := context.Background()
	questions, err := uc.GenerateReviewQuestions(ctx, 1, 1, "初中一年级")

	if err != nil {
		t.Fatalf("GenerateReviewQuestions() error = %v", err)
	}

	if len(questions) != 2 {
		t.Fatalf("GenerateReviewQuestions() returned %d questions, want 2", len(questions))
	}

	if questions[0].Type != "multiple_choice" {
		t.Errorf("Question 0 type = %s, want multiple_choice", questions[0].Type)
	}

	if questions[1].Type != "fill_blank" {
		t.Errorf("Question 1 type = %s, want fill_blank", questions[1].Type)
	}
}

func TestWordUsecase_GenerateReviewQuestions_LLMNotConfigured(t *testing.T) {
	logger := log.NewStdLogger(nil)
	studentRepo := &mockStudentRepo{}
	wordRepo := &mockWordRepo{
		words: []*Word{
			{
				ID:        1,
				StudentID: 1,
				Word:      "hello",
				Meaning:   "你好",
			},
		},
	}
	var llmService LLMService = nil // LLM服务未配置

	uc := NewWordUsecase(wordRepo, studentRepo, llmService, logger)

	ctx := context.Background()
	_, err := uc.GenerateReviewQuestions(ctx, 1, 1, "")

	if err == nil {
		t.Error("GenerateReviewQuestions() error = nil, want error")
	}
}
