package data

import (
	"context"
	"errors"
	"forgetting-curve/backend/internal/biz"
)

type confusedWordRepo struct {
	data *Data
}

// NewConfusedWordRepo 创建混淆词仓库
func NewConfusedWordRepo(data *Data) biz.ConfusedWordRepo {
	return &confusedWordRepo{data: data}
}

// Create 创建混淆词关联
func (r *confusedWordRepo) Create(ctx context.Context, confusedWord *biz.ConfusedWord) error {
	return r.data.db.WithContext(ctx).Create(confusedWord).Error
}

// GetByWordID 根据单词ID获取混淆词列表
func (r *confusedWordRepo) GetByWordID(ctx context.Context, wordID int64) ([]*biz.ConfusedWord, error) {
	var confusedWords []*biz.ConfusedWord
	err := r.data.db.WithContext(ctx).
		Where("word_id = ?", wordID).
		Find(&confusedWords).Error
	return confusedWords, err
}

// Delete 删除混淆词关联
func (r *confusedWordRepo) Delete(ctx context.Context, wordID, confusedWordID int64) error {
	result := r.data.db.WithContext(ctx).
		Where("word_id = ? AND confused_word_id = ?", wordID, confusedWordID).
		Delete(&biz.ConfusedWord{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("confused word not found")
	}
	return nil
}

// CheckExists 检查混淆词关联是否存在
func (r *confusedWordRepo) CheckExists(ctx context.Context, wordID, confusedWordID int64) (bool, error) {
	var count int64
	err := r.data.db.WithContext(ctx).
		Model(&biz.ConfusedWord{}).
		Where("word_id = ? AND confused_word_id = ?", wordID, confusedWordID).
		Count(&count).Error
	return count > 0, err
}
