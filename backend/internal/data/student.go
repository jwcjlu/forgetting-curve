package data

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// StudentRepo 学生数据仓库接口
type StudentRepo interface {
	Create(ctx context.Context, student *Student) (*Student, error)
	GetByID(ctx context.Context, id int64) (*Student, error)
	GetByStudentNo(ctx context.Context, studentNo string) (*Student, error)
	GetByOpenID(ctx context.Context, openid string) (*Student, error)
	Update(ctx context.Context, student *Student) (*Student, error)
}

type studentRepo struct {
	data *Data
}

// NewStudentRepo 创建学生仓库
func NewStudentRepo(data *Data) StudentRepo {
	return &studentRepo{data: data}
}

// Create 创建学生
func (r *studentRepo) Create(ctx context.Context, student *Student) (*Student, error) {
	if err := r.data.db.WithContext(ctx).Create(student).Error; err != nil {
		return nil, err
	}
	return student, nil
}

// GetByID 根据ID获取学生
func (r *studentRepo) GetByID(ctx context.Context, id int64) (*Student, error) {
	var student Student
	if err := r.data.db.WithContext(ctx).First(&student, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("student not found")
		}
		return nil, err
	}
	return &student, nil
}

// GetByStudentNo 根据学号获取学生
func (r *studentRepo) GetByStudentNo(ctx context.Context, studentNo string) (*Student, error) {
	var student Student
	if err := r.data.db.WithContext(ctx).Where("student_no = ?", studentNo).First(&student).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("student not found")
		}
		return nil, err
	}
	return &student, nil
}

// GetByOpenID 根据openid获取学生
func (r *studentRepo) GetByOpenID(ctx context.Context, openid string) (*Student, error) {
	var student Student
	if err := r.data.db.WithContext(ctx).Where("open_id = ?", openid).First(&student).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("student not found")
		}
		return nil, err
	}
	return &student, nil
}

// Update 更新学生信息
func (r *studentRepo) Update(ctx context.Context, student *Student) (*Student, error) {
	if err := r.data.db.WithContext(ctx).Save(student).Error; err != nil {
		return nil, err
	}
	return student, nil
}
