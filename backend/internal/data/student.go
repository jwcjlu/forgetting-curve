package data

import (
	"context"
	"errors"
	"forgetting-curve/backend/internal/biz"

	"gorm.io/gorm"
)

type studentRepo struct {
	data *Data
}

// NewStudentRepo 创建学生仓库
func NewStudentRepo(data *Data) biz.StudentRepo {
	return &studentRepo{data: data}
}

// Create 创建学生
func (r *studentRepo) Create(ctx context.Context, student *biz.Student) (*biz.Student, error) {
	if err := r.data.db.WithContext(ctx).Create(student).Error; err != nil {
		return nil, err
	}
	return student, nil
}

// GetByID 根据ID获取学生
func (r *studentRepo) GetByID(ctx context.Context, id int64) (*biz.Student, error) {
	var student biz.Student
	if err := r.data.db.WithContext(ctx).First(&student, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("student not found")
		}
		return nil, err
	}
	return &student, nil
}

// GetByStudentNo 根据学号获取学生
func (r *studentRepo) GetByStudentNo(ctx context.Context, studentNo string) (*biz.Student, error) {
	var student biz.Student
	if err := r.data.db.WithContext(ctx).Where("student_no = ?", studentNo).First(&student).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("student not found")
		}
		return nil, err
	}
	return &student, nil
}

// GetByOpenID 根据openid获取学生
func (r *studentRepo) GetByOpenID(ctx context.Context, openid string) (*biz.Student, error) {
	var student biz.Student
	if err := r.data.db.WithContext(ctx).Where("open_id = ?", openid).First(&student).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("student not found")
		}
		return nil, err
	}
	return &student, nil
}

// Update 更新学生信息
func (r *studentRepo) Update(ctx context.Context, student *biz.Student) (*biz.Student, error) {
	if err := r.data.db.WithContext(ctx).Save(student).Error; err != nil {
		return nil, err
	}
	return student, nil
}
