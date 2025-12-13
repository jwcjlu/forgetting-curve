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

// GetByPlanID 根据计划ID获取单词列表（分页）
func (r *wordRepo) GetByPlanID(ctx context.Context, planID int64, page, pageSize int32) ([]*biz.Word, int64, error) {
	var words []*biz.Word
	var total int64

	// 查询总数
	if err := r.data.db.WithContext(ctx).Model(&biz.Word{}).Where("plan_id = ?", planID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := r.data.db.WithContext(ctx).
		Where("plan_id = ?", planID).
		Order("created_at DESC").
		Offset(int(offset)).
		Limit(int(pageSize)).
		Find(&words).Error; err != nil {
		return nil, 0, err
	}

	return words, total, nil
}

// GetAllByPlanID 根据计划ID获取所有单词（不分页，用于今日单词计算）
func (r *wordRepo) GetAllByPlanID(ctx context.Context, planID int64) ([]*biz.Word, error) {
	var words []*biz.Word
	if err := r.data.db.WithContext(ctx).
		Where("plan_id = ?", planID).
		Order("created_at DESC").
		Find(&words).Error; err != nil {
		return nil, err
	}
	return words, nil
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

// GetByIDs 根据ID列表获取单词
func (r *wordRepo) GetByIDs(ctx context.Context, wordIDs []int64) ([]*biz.Word, error) {
	if len(wordIDs) == 0 {
		return []*biz.Word{}, nil
	}
	var words []*biz.Word
	if err := r.data.db.WithContext(ctx).
		Where("id IN ?", wordIDs).
		Find(&words).Error; err != nil {
		return nil, err
	}
	return words, nil
}
