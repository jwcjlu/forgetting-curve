package data

import (
	"context"
	"errors"
	"forgetting-curve/backend/internal/biz"

	"gorm.io/gorm"
)

type planRepo struct {
	data *Data
}

// NewPlanRepo 创建计划仓库
func NewPlanRepo(data *Data) biz.PlanRepo {
	return &planRepo{data: data}
}

// Create 创建计划
func (r *planRepo) Create(ctx context.Context, plan *biz.Plan) (*biz.Plan, error) {
	if err := r.data.db.WithContext(ctx).Create(plan).Error; err != nil {
		return nil, err
	}
	return plan, nil
}

// GetByID 根据ID获取计划
func (r *planRepo) GetByID(ctx context.Context, id int64) (*biz.Plan, error) {
	var plan biz.Plan
	if err := r.data.db.WithContext(ctx).First(&plan, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("plan not found")
		}
		return nil, err
	}
	return &plan, nil
}

// GetByStudentID 根据学生ID获取所有计划
func (r *planRepo) GetByStudentID(ctx context.Context, studentID int64) ([]*biz.Plan, error) {
	var plans []*biz.Plan
	if err := r.data.db.WithContext(ctx).
		Where("student_id = ?", studentID).
		Order("created_at DESC").
		Find(&plans).Error; err != nil {
		return nil, err
	}
	return plans, nil
}

// GetActivePlanByStudentID 根据学生ID获取当前激活的计划
func (r *planRepo) GetActivePlanByStudentID(ctx context.Context, studentID int64) (*biz.Plan, error) {
	var plan biz.Plan
	if err := r.data.db.WithContext(ctx).
		Where("student_id = ? AND is_active = ?", studentID, true).
		First(&plan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 没有激活的计划，返回nil而不是错误
		}
		return nil, err
	}
	return &plan, nil
}

// Update 更新计划
func (r *planRepo) Update(ctx context.Context, plan *biz.Plan) (*biz.Plan, error) {
	if err := r.data.db.WithContext(ctx).Save(plan).Error; err != nil {
		return nil, err
	}
	return plan, nil
}

// Delete 删除计划
func (r *planRepo) Delete(ctx context.Context, id int64) error {
	return r.data.db.WithContext(ctx).Delete(&biz.Plan{}, id).Error
}
