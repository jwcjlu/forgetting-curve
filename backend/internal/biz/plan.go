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
	AddWordsToPlan(ctx context.Context, studentID int64, planID int64, wordIDs []int64) (int32, error)
	RemoveWordsFromPlan(ctx context.Context, studentID int64, planID int64, wordID int64) error
	GetPlanWords(ctx context.Context, studentID int64, planID int64) ([]*Word, error)
	DeletePlan(ctx context.Context, studentID int64, planID int64) error
	GetActivePlan(ctx context.Context, studentID int64) (*Plan, error)
}

type planUsecase struct {
	planRepo     PlanRepo
	planWordRepo PlanWordRepo
	wordRepo     WordRepo
	studentRepo  StudentRepo
	log          *log.Helper
}

// NewPlanUsecase 创建计划业务逻辑
func NewPlanUsecase(
	planRepo PlanRepo,
	planWordRepo PlanWordRepo,
	wordRepo WordRepo,
	studentRepo StudentRepo,
	logger log.Logger,
) PlanUsecase {
	return &planUsecase{
		planRepo:     planRepo,
		planWordRepo: planWordRepo,
		wordRepo:     wordRepo,
		studentRepo:  studentRepo,
		log:          log.NewHelper(logger),
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

// AddWordsToPlan 添加单词到计划
func (uc *planUsecase) AddWordsToPlan(ctx context.Context, studentID int64, planID int64, wordIDs []int64) (int32, error) {
	// 验证学生是否存在
	_, err := uc.studentRepo.GetByID(ctx, studentID)
	if err != nil {
		return 0, errors.New("学生不存在")
	}

	// 验证计划是否存在且属于该学生
	plan, err := uc.planRepo.GetByID(ctx, planID)
	if err != nil {
		return 0, errors.New("计划不存在")
	}
	if plan.StudentID != studentID {
		return 0, errors.New("无权访问该计划")
	}

	// 验证单词是否存在且属于该学生
	words, err := uc.wordRepo.GetByIDs(ctx, wordIDs)
	if err != nil {
		return 0, err
	}

	// 过滤出属于该学生的单词
	validWordIDs := make([]int64, 0)
	for _, word := range words {
		if word.StudentID == studentID {
			validWordIDs = append(validWordIDs, word.ID)
		}
	}

	if len(validWordIDs) == 0 {
		return 0, nil
	}

	// 检查单词是否已在计划中
	existingPlanWords, err := uc.planWordRepo.GetByPlanID(ctx, planID)
	if err != nil {
		return 0, err
	}

	existingWordIDs := make(map[int64]bool)
	for _, pw := range existingPlanWords {
		existingWordIDs[pw.WordID] = true
	}

	// 只添加不存在的单词
	newPlanWords := make([]*PlanWord, 0)
	for _, wordID := range validWordIDs {
		if !existingWordIDs[wordID] {
			newPlanWords = append(newPlanWords, &PlanWord{
				PlanID: planID,
				WordID: wordID,
			})
		}
	}

	if len(newPlanWords) == 0 {
		return 0, nil
	}

	err = uc.planWordRepo.BatchCreate(ctx, newPlanWords)
	if err != nil {
		return 0, err
	}

	return int32(len(newPlanWords)), nil
}

// RemoveWordsFromPlan 从计划中移除单词
func (uc *planUsecase) RemoveWordsFromPlan(ctx context.Context, studentID int64, planID int64, wordID int64) error {
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

	return uc.planWordRepo.DeleteByPlanIDAndWordID(ctx, planID, wordID)
}

// GetPlanWords 获取计划中的单词列表
func (uc *planUsecase) GetPlanWords(ctx context.Context, studentID int64, planID int64) ([]*Word, error) {
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

	// 获取计划中的单词ID列表
	wordIDs, err := uc.planWordRepo.GetWordIDsByPlanID(ctx, planID)
	if err != nil {
		return nil, err
	}

	if len(wordIDs) == 0 {
		return []*Word{}, nil
	}

	// 获取单词详情
	return uc.wordRepo.GetByIDs(ctx, wordIDs)
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

	// 删除计划中的单词关联
	err = uc.planWordRepo.DeleteByPlanID(ctx, planID)
	if err != nil {
		return err
	}

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
