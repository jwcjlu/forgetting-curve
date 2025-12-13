package biz

import (
	"context"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
)

// mockConfusedWordRepo 模拟混淆词仓库
type mockConfusedWordRepo struct {
	confusedWords []*ConfusedWord
	words         []*Word
	err           error
}

func (m *mockConfusedWordRepo) Create(ctx context.Context, confusedWord *ConfusedWord) error {
	if m.err != nil {
		return m.err
	}
	m.confusedWords = append(m.confusedWords, confusedWord)
	return nil
}

func (m *mockConfusedWordRepo) Delete(ctx context.Context, wordID, confusedWordID int64) error {
	if m.err != nil {
		return m.err
	}
	var filtered []*ConfusedWord
	for _, cw := range m.confusedWords {
		if !(cw.WordID == wordID && cw.ConfusedWordID == confusedWordID) {
			filtered = append(filtered, cw)
		}
	}
	m.confusedWords = filtered
	return nil
}

func (m *mockConfusedWordRepo) GetByWordID(ctx context.Context, wordID int64) ([]*ConfusedWord, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []*ConfusedWord
	for _, cw := range m.confusedWords {
		if cw.WordID == wordID {
			result = append(result, cw)
		}
	}
	return result, nil
}

func (m *mockConfusedWordRepo) CheckExists(ctx context.Context, wordID, confusedWordID int64) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	for _, cw := range m.confusedWords {
		if cw.WordID == wordID && cw.ConfusedWordID == confusedWordID {
			return true, nil
		}
	}
	return false, nil
}

// mockWordRepoForSearch 用于搜索的模拟单词仓库
type mockWordRepoForSearch struct {
	words []*Word
	err   error
}

func (m *mockWordRepoForSearch) BatchCreate(ctx context.Context, words []*Word) error {
	return nil
}

func (m *mockWordRepoForSearch) GetByStudentID(ctx context.Context, studentID int64, page, pageSize int32) ([]*Word, int64, error) {
	if m.err != nil {
		return nil, 0, m.err
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

func (m *mockWordRepoForSearch) GetByID(ctx context.Context, id int64) (*Word, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, w := range m.words {
		if w.ID == id {
			return w, nil
		}
	}
	return nil, nil
}

func (m *mockWordRepoForSearch) Update(ctx context.Context, word *Word) (*Word, error) {
	return word, nil
}

func TestConfusedWordUsecase_AddConfusedWord(t *testing.T) {
	logger := log.NewStdLogger(nil)
	confusedWordRepo := &mockConfusedWordRepo{
		words: []*Word{
			{ID: 1, StudentID: 1, Word: "hello", Meaning: "你好"},
			{ID: 2, StudentID: 1, Word: "world", Meaning: "世界"},
		},
	}
	wordRepo := &mockWordRepo{
		words: []*Word{
			{ID: 1, StudentID: 1, Word: "hello", Meaning: "你好"},
			{ID: 2, StudentID: 1, Word: "world", Meaning: "世界"},
		},
	}
	studentRepo := &mockStudentRepo{
		students: map[int64]*Student{
			1: {ID: 1, Name: "Test Student"},
		},
	}

	uc := NewConfusedWordUsecase(confusedWordRepo, wordRepo, studentRepo, logger)

	tests := []struct {
		name           string
		studentID      int64
		wordID         int64
		confusedWordID int64
		wantErr        bool
	}{
		{
			name:           "正常添加",
			studentID:      1,
			wordID:         1,
			confusedWordID: 2,
			wantErr:        false,
		},
		{
			name:           "单词不存在",
			studentID:      1,
			wordID:         999,
			confusedWordID: 2,
			wantErr:        true,
		},
		{
			name:           "混淆词不存在",
			studentID:      1,
			wordID:         1,
			confusedWordID: 999,
			wantErr:        true,
		},
		{
			name:           "不能添加自己",
			studentID:      1,
			wordID:         1,
			confusedWordID: 1,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confusedWordRepo.confusedWords = []*ConfusedWord{} // 重置
			ctx := context.Background()
			err := uc.AddConfusedWord(ctx, tt.studentID, tt.wordID, tt.confusedWordID)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddConfusedWord() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(confusedWordRepo.confusedWords) == 0 {
				t.Error("AddConfusedWord() should create confused word")
			}
		})
	}
}

func TestConfusedWordUsecase_RemoveConfusedWord(t *testing.T) {
	logger := log.NewStdLogger(nil)
	confusedWordRepo := &mockConfusedWordRepo{
		confusedWords: []*ConfusedWord{
			{ID: 1, WordID: 1, ConfusedWordID: 2},
		},
		words: []*Word{
			{ID: 1, StudentID: 1, Word: "hello", Meaning: "你好"},
			{ID: 2, StudentID: 1, Word: "world", Meaning: "世界"},
		},
	}
	wordRepo := &mockWordRepo{
		words: []*Word{
			{ID: 1, StudentID: 1, Word: "hello", Meaning: "你好"},
		},
	}
	studentRepo := &mockStudentRepo{
		students: map[int64]*Student{
			1: {ID: 1, Name: "Test Student"},
		},
	}

	uc := NewConfusedWordUsecase(confusedWordRepo, wordRepo, studentRepo, logger)

	ctx := context.Background()
	err := uc.RemoveConfusedWord(ctx, 1, 1, 2)

	if err != nil {
		t.Fatalf("RemoveConfusedWord() error = %v", err)
	}

	if len(confusedWordRepo.confusedWords) != 0 {
		t.Error("RemoveConfusedWord() should remove confused word")
	}
}

func TestConfusedWordUsecase_GetConfusedWords(t *testing.T) {
	logger := log.NewStdLogger(nil)
	confusedWordRepo := &mockConfusedWordRepo{
		confusedWords: []*ConfusedWord{
			{ID: 1, WordID: 1, ConfusedWordID: 2},
			{ID: 2, WordID: 1, ConfusedWordID: 3},
		},
	}
	wordRepo := &mockWordRepo{
		words: []*Word{
			{ID: 1, StudentID: 1, Word: "hello", Meaning: "你好"},
			{ID: 2, StudentID: 1, Word: "world", Meaning: "世界"},
			{ID: 3, StudentID: 1, Word: "test", Meaning: "测试"},
		},
	}
	studentRepo := &mockStudentRepo{
		students: map[int64]*Student{
			1: {ID: 1, Name: "Test Student"},
		},
	}

	uc := NewConfusedWordUsecase(confusedWordRepo, wordRepo, studentRepo, logger)

	ctx := context.Background()
	words, err := uc.GetConfusedWords(ctx, 1, 1)

	if err != nil {
		t.Fatalf("GetConfusedWords() error = %v", err)
	}

	if len(words) != 2 {
		t.Errorf("GetConfusedWords() returned %d words, want 2", len(words))
	}
}

func TestConfusedWordUsecase_SearchWords(t *testing.T) {
	logger := log.NewStdLogger(nil)
	confusedWordRepo := &mockConfusedWordRepo{}
	wordRepo := &mockWordRepo{
		words: []*Word{
			{ID: 1, StudentID: 1, Word: "hello", Meaning: "你好"},
			{ID: 2, StudentID: 1, Word: "world", Meaning: "世界"},
			{ID: 3, StudentID: 1, Word: "test", Meaning: "测试"},
			{ID: 4, StudentID: 2, Word: "other", Meaning: "其他"},
		},
	}
	studentRepo := &mockStudentRepo{
		students: map[int64]*Student{
			1: {ID: 1, Name: "Test Student"},
			2: {ID: 2, Name: "Other Student"},
		},
	}

	uc := NewConfusedWordUsecase(confusedWordRepo, wordRepo, studentRepo, logger)

	tests := []struct {
		name      string
		studentID int64
		keyword   string
		limit     int32
		wantCount int
	}{
		{
			name:      "搜索单词",
			studentID: 1,
			keyword:   "hello",
			limit:     10,
			wantCount: 1,
		},
		{
			name:      "搜索释义",
			studentID: 1,
			keyword:   "世界",
			limit:     10,
			wantCount: 1,
		},
		{
			name:      "限制数量",
			studentID: 1,
			keyword:   "l", // 匹配hello和world
			limit:     1,
			wantCount: 1,
		},
		{
			name:      "不同学生",
			studentID: 2,
			keyword:   "hello",
			limit:     10,
			wantCount: 0, // 学生2没有hello
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			words, err := uc.SearchWords(ctx, tt.studentID, tt.keyword, tt.limit)

			if err != nil {
				t.Fatalf("SearchWords() error = %v", err)
			}

			if len(words) != tt.wantCount {
				t.Errorf("SearchWords() returned %d words, want %d", len(words), tt.wantCount)
			}
		})
	}
}
