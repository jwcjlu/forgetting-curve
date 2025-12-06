package data

import (
	"context"
	"forgetting-curve/backend/internal/biz"
)

type wordRepo struct {
	data *Data
}

// NewWordRepo 创建单词仓库
func NewWordRepo(data *Data) biz.WordRepo {
	return &wordRepo{data: data}
}

// BatchCreate 批量创建单词
func (r *wordRepo) BatchCreate(ctx context.Context, words []*biz.Word) error {
	if len(words) == 0 {
		return nil
	}
	return r.data.db.WithContext(ctx).CreateInBatches(words, 100).Error
}

// GetByStudentID 根据学生ID获取单词列表（分页）
func (r *wordRepo) GetByStudentID(ctx context.Context, studentID int64, page, pageSize int32) ([]*biz.Word, int64, error) {
	var words []*biz.Word
	var total int64

	// 查询总数
	if err := r.data.db.WithContext(ctx).Model(&biz.Word{}).Where("student_id = ?", studentID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.data.db.WithContext(ctx).
		Where("student_id = ?", studentID).
		Order("created_at DESC").
		Offset(int(offset)).
		Limit(int(pageSize)).
		Find(&words).Error; err != nil {
		return nil, 0, err
	}

	return words, total, nil
}

// GetByID 根据ID获取单词
func (r *wordRepo) GetByID(ctx context.Context, id int64) (*biz.Word, error) {
	var word biz.Word
	if err := r.data.db.WithContext(ctx).First(&word, id).Error; err != nil {
		return nil, err
	}
	return &word, nil
}

// Update 更新单词
func (r *wordRepo) Update(ctx context.Context, word *biz.Word) (*biz.Word, error) {
	if err := r.data.db.WithContext(ctx).Save(word).Error; err != nil {
		return nil, err
	}
	return word, nil
}
