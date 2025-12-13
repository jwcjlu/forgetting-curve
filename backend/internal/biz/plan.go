package biz

import (
	"context"
	"errors"

	"github.com/go-kratos/kratos/v2/log"
)

// PlanUsecase 计划业务逻辑接口
type PlanUsecase interface {
	CreatePlan(ctx context.Context, studentID int64, name, grade string) (*Plan, error)
	GetPlans(ctx context.Context, studentID int64) ([]*Plan, error)
	SelectPlan(ctx context.Context, studentID int64, planID int64) (*Plan, error)
	UpdatePlan(ctx context.Context, studentID int64, planID int64, name, grade string) (*Plan, error)
	GetPlanWords(ctx context.Context, studentID int64, planID int64, page, pageSize int32) ([]*Word, int64, error)
	DeletePlan(ctx context.Context, studentID int64, planID int64) error
	GetActivePlan(ctx context.Context, studentID int64) (*Plan, error)
}

type planUsecase struct {
	planRepo    PlanRepo
	wordRepo    WordRepo
	studentRepo StudentRepo
	log         *log.Helper
}

// NewPlanUsecase 创建计划业务逻辑
func NewPlanUsecase(
	planRepo PlanRepo,
	wordRepo WordRepo,
	studentRepo StudentRepo,
	logger log.Logger,
) PlanUsecase {
	return &planUsecase{
		planRepo:    planRepo,
		wordRepo:    wordRepo,
		studentRepo: studentRepo,
		log:         log.NewHelper(logger),
	}
}

// CreatePlan 创建复习计划
func (uc *planUsecase) CreatePlan(ctx context.Context, studentID int64, name, grade string) (*Plan, error) {
	// 验证学生是否存在
	_, err := uc.studentRepo.GetByID(ctx, studentID)
	if err != nil {
		return nil, errors.New("学生不存在")
	}

	plan := &Plan{
		StudentID: studentID,
		Name:      name,
		Grade:     grade,
		IsActive:  false,
	}

	return uc.planRepo.Create(ctx, plan)
}

// GetPlans 获取学生的所有计划
func (uc *planUsecase) GetPlans(ctx context.Context, studentID int64) ([]*Plan, error) {
	// 验证学生是否存在
	_, err := uc.studentRepo.GetByID(ctx, studentID)
	if err != nil {
		return nil, errors.New("学生不存在")
	}

	return uc.planRepo.GetByStudentID(ctx, studentID)
}

// SelectPlan 选择计划（设置为当前激活的计划）
func (uc *planUsecase) SelectPlan(ctx context.Context, studentID int64, planID int64) (*Plan, error) {
	// 验证学生是否存在
	_, err := uc.studentRepo.GetByID(ctx, studentID)
	if err != nil {
		return nil, errors.New("学生不存在")
	}

	// 获取计划
	plan, err := uc.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, errors.New("计划不存在")
	}

	// 验证计划是否属于该学生
	if plan.StudentID != studentID {
		return nil, errors.New("无权访问该计划")
	}

	// 将该学生的所有计划设置为非激活
	plans, err := uc.planRepo.GetByStudentID(ctx, studentID)
	if err != nil {
		return nil, err
	}

	for _, p := range plans {
		if p.IsActive {
			p.IsActive = false
			_, err = uc.planRepo.Update(ctx, p)
			if err != nil {
				return nil, err
			}
		}
	}

	// 设置当前计划为激活
	plan.IsActive = true
	return uc.planRepo.Update(ctx, plan)
}

// UpdatePlan 更新计划
func (uc *planUsecase) UpdatePlan(ctx context.Context, studentID int64, planID int64, name, grade string) (*Plan, error) {
	// 验证学生是否存在
	_, err := uc.studentRepo.GetByID(ctx, studentID)
	if err != nil {
		return nil, errors.New("学生不存在")
	}

	// 验证计划是否存在且属于该学生
	plan, err := uc.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, errors.New("计划不存在")
	}
	if plan.StudentID != studentID {
		return nil, errors.New("无权访问该计划")
	}

	// 更新计划信息
	plan.Name = name
	plan.Grade = grade

	return uc.planRepo.Update(ctx, plan)
}

// GetPlanWords 获取计划中的单词列表
func (uc *planUsecase) GetPlanWords(ctx context.Context, studentID int64, planID int64, page, pageSize int32) ([]*Word, int64, error) {
	// 验证学生是否存在
	_, err := uc.studentRepo.GetByID(ctx, studentID)
	if err != nil {
		return nil, 0, errors.New("学生不存在")
	}

	// 验证计划是否存在且属于该学生
	plan, err := uc.planRepo.GetByID(ctx, planID)
	if err != nil {
		return nil, 0, errors.New("计划不存在")
	}
	if plan.StudentID != studentID {
		return nil, 0, errors.New("无权访问该计划")
	}

	// 直接从计划获取单词列表
	return uc.wordRepo.GetByPlanID(ctx, planID, page, pageSize)
}

// DeletePlan 删除计划
func (uc *planUsecase) DeletePlan(ctx context.Context, studentID int64, planID int64) error {
	// 验证学生是否存在
	_, err := uc.studentRepo.GetByID(ctx, studentID)
	if err != nil {
		return errors.New("学生不存在")
	}

	// 验证计划是否存在且属于该学生
	plan, err := uc.planRepo.GetByID(ctx, planID)
	if err != nil {
		return errors.New("计划不存在")
	}
	if plan.StudentID != studentID {
		return errors.New("无权访问该计划")
	}

	// 注意：删除计划时，可以选择是否同时删除计划下的单词
	// 这里我们选择只删除计划，单词保留（可以通过软删除或迁移到其他计划）
	// 如果需要删除单词，可以添加逻辑：
	// err = uc.wordRepo.DeleteByPlanID(ctx, planID)

	// 删除计划
	return uc.planRepo.Delete(ctx, planID)
}

// GetActivePlan 获取当前激活的计划
func (uc *planUsecase) GetActivePlan(ctx context.Context, studentID int64) (*Plan, error) {
	// 验证学生是否存在
	_, err := uc.studentRepo.GetByID(ctx, studentID)
	if err != nil {
		return nil, errors.New("学生不存在")
	}

	return uc.planRepo.GetActivePlanByStudentID(ctx, studentID)
}
