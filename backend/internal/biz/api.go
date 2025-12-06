package biz

import "context"

// StudentRepo 学生数据仓库接口
type StudentRepo interface {
	Create(ctx context.Context, student *Student) (*Student, error)
	GetByID(ctx context.Context, id int64) (*Student, error)
	GetByStudentNo(ctx context.Context, studentNo string) (*Student, error)
	GetByOpenID(ctx context.Context, openid string) (*Student, error)
	Update(ctx context.Context, student *Student) (*Student, error)
}

// WordRepo 单词数据仓库接口
type WordRepo interface {
	BatchCreate(ctx context.Context, words []*Word) error
	GetByStudentID(ctx context.Context, studentID int64, page, pageSize int32) ([]*Word, int64, error)
	GetByID(ctx context.Context, id int64) (*Word, error)
	Update(ctx context.Context, word *Word) (*Word, error)
}
