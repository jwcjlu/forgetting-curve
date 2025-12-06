package biz

import (
	"context"
	"errors"

	"github.com/go-kratos/kratos/v2/log"
)

// StudentUsecase 学生业务逻辑接口
type StudentUsecase interface {
	CreateStudent(ctx context.Context, name, studentNo string) (*Student, error)
	GetStudent(ctx context.Context, id int64) (*Student, error)
	GetStudentByNo(ctx context.Context, studentNo string) (*Student, error)
	GetOrCreateStudentByOpenid(ctx context.Context, openid, name string) (*Student, bool, error)
}

type studentUsecase struct {
	repo StudentRepo
	log  *log.Helper
}

// NewStudentUsecase 创建学生业务逻辑
func NewStudentUsecase(repo StudentRepo, logger log.Logger) StudentUsecase {
	return &studentUsecase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// CreateStudent 创建学生
func (uc *studentUsecase) CreateStudent(ctx context.Context, name, studentNo string) (*Student, error) {
	// 验证输入
	if name == "" {
		return nil, errors.New("student name cannot be empty")
	}
	if studentNo == "" {
		return nil, errors.New("student number cannot be empty")
	}

	// 检查学号是否已存在
	_, err := uc.repo.GetByStudentNo(ctx, studentNo)
	if err == nil {
		return nil, errors.New("student number already exists")
	}

	// 创建学生
	student := &Student{
		Name:      name,
		StudentNo: studentNo,
	}

	return uc.repo.Create(ctx, student)
}

// GetStudent 获取学生信息
func (uc *studentUsecase) GetStudent(ctx context.Context, id int64) (*Student, error) {
	return uc.repo.GetByID(ctx, id)
}

// GetStudentByNo 根据学号获取学生
func (uc *studentUsecase) GetStudentByNo(ctx context.Context, studentNo string) (*Student, error) {
	return uc.repo.GetByStudentNo(ctx, studentNo)
}

// GetOrCreateStudentByOpenid 通过openid获取或创建学生
func (uc *studentUsecase) GetOrCreateStudentByOpenid(ctx context.Context, openid, name string) (*Student, bool, error) {
	if openid == "" {
		return nil, false, errors.New("openid cannot be empty")
	}

	// 尝试获取已存在的学生
	student, err := uc.repo.GetByOpenID(ctx, openid)
	if err == nil {
		// 学生已存在，更新名称（如果提供了新名称）
		if name != "" && name != student.Name {
			student.Name = name
			updated, err := uc.repo.Update(ctx, student)
			if err != nil {
				uc.log.Warnf("failed to update student name: %v", err)
				return student, false, nil
			}
			return updated, false, nil
		}
		return student, false, nil
	}

	// 学生不存在，创建新学生
	// 生成一个唯一的学号（使用openid的一部分，如果openid长度足够）
	studentNo := ""
	if len(openid) >= 8 {
		studentNo = "wx_" + openid[:8] // 使用openid前8位作为学号
	} else {
		studentNo = "wx_" + openid // 如果openid太短，直接使用
	}
	if name == "" {
		name = "微信用户"
	}

	student = &Student{
		Name:      name,
		StudentNo: studentNo,
		OpenID:    openid,
	}

	created, err := uc.repo.Create(ctx, student)
	if err != nil {
		return nil, false, err
	}

	return created, true, nil
}
