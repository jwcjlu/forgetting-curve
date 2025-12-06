package service

import (
	"context"
	"time"

	v1 "forgetting-curve/backend/api/student/v1"
	"forgetting-curve/backend/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// StudentService 学生服务
type StudentService struct {
	v1.UnimplementedStudentServiceServer

	studentUc biz.StudentUsecase
	wordUc    biz.WordUsecase
	log       *log.Helper
}

// NewStudentService 创建学生服务
func NewStudentService(studentUc biz.StudentUsecase, wordUc biz.WordUsecase, logger log.Logger) *StudentService {
	return &StudentService{
		studentUc: studentUc,
		wordUc:    wordUc,
		log:       log.NewHelper(logger),
	}
}

// CreateStudent 添加学生
func (s *StudentService) CreateStudent(ctx context.Context, req *v1.CreateStudentRequest) (*v1.CreateStudentReply, error) {
	student, err := s.studentUc.CreateStudent(ctx, req.Name, req.StudentNo)
	if err != nil {
		return &v1.CreateStudentReply{
			Ret: &v1.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	return &v1.CreateStudentReply{
		Ret: &v1.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Student: &v1.Student{
			Id:        student.ID,
			Name:      student.Name,
			StudentNo: student.StudentNo,
			Openid:    student.OpenID,
			CreatedAt: student.CreatedAt.Unix(),
			UpdatedAt: student.UpdatedAt.Unix(),
		},
	}, nil
}

// GetStudent 获取学生信息
func (s *StudentService) GetStudent(ctx context.Context, req *v1.GetStudentRequest) (*v1.GetStudentReply, error) {
	student, err := s.studentUc.GetStudent(ctx, req.Id)
	if err != nil {
		return &v1.GetStudentReply{
			Ret: &v1.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	return &v1.GetStudentReply{
		Ret: &v1.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Student: &v1.Student{
			Id:        student.ID,
			Name:      student.Name,
			StudentNo: student.StudentNo,
			Openid:    student.OpenID,
			CreatedAt: student.CreatedAt.Unix(),
			UpdatedAt: student.UpdatedAt.Unix(),
		},
	}, nil
}

// BatchAddWords 批量添加单词到学生名下
func (s *StudentService) BatchAddWords(ctx context.Context, req *v1.BatchAddWordsRequest) (*v1.BatchAddWordsReply, error) {
	// 转换请求数据
	wordItems := make([]*biz.WordItem, 0, len(req.Words))
	for _, item := range req.Words {
		wordItems = append(wordItems, &biz.WordItem{
			Word:      item.Word,
			Meaning:   item.Meaning,
			StartDate: item.StartDate,
		})
	}

	// 调用业务逻辑
	words, err := s.wordUc.BatchAddWords(ctx, req.StudentId, wordItems)
	if err != nil {
		return &v1.BatchAddWordsReply{
			Ret: &v1.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	// 转换响应数据
	v1Words := make([]*v1.Word, 0, len(words))
	for _, word := range words {
		v1Words = append(v1Words, &v1.Word{
			Id:             word.ID,
			StudentId:      word.StudentID,
			Word:           word.Word,
			Meaning:        word.Meaning,
			StartDate:      word.StartDate,
			ReviewCount:    word.ReviewCount,
			LastReviewDate: word.LastReviewDate,
			CreatedAt:      word.CreatedAt.Unix(),
			UpdatedAt:      word.UpdatedAt.Unix(),
		})
	}

	return &v1.BatchAddWordsReply{
		Ret: &v1.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Count: int32(len(v1Words)),
		Words: v1Words,
	}, nil
}

// GetStudentWords 获取学生的单词列表（学生只能看自己的单词）
func (s *StudentService) GetStudentWords(ctx context.Context, req *v1.GetStudentWordsRequest) (*v1.GetStudentWordsReply, error) {
	// 设置默认分页参数
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	// 调用业务逻辑（已经验证了学生ID，确保只能看自己的单词）
	words, total, err := s.wordUc.GetStudentWords(ctx, req.StudentId, page, pageSize)
	if err != nil {
		return &v1.GetStudentWordsReply{
			Ret: &v1.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	// 转换响应数据
	v1Words := make([]*v1.Word, 0, len(words))
	for _, word := range words {
		v1Words = append(v1Words, &v1.Word{
			Id:             word.ID,
			StudentId:      word.StudentID,
			Word:           word.Word,
			Meaning:        word.Meaning,
			StartDate:      word.StartDate,
			ReviewCount:    word.ReviewCount,
			LastReviewDate: word.LastReviewDate,
			CreatedAt:      word.CreatedAt.Unix(),
			UpdatedAt:      word.UpdatedAt.Unix(),
		})
	}

	return &v1.GetStudentWordsReply{
		Ret: &v1.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Words:    v1Words,
		Total:    int32(total),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetTodayWords 获取今日需要背诵的单词（根据艾宾浩斯曲线）
func (s *StudentService) GetTodayWords(ctx context.Context, req *v1.GetTodayWordsRequest) (*v1.GetTodayWordsReply, error) {
	words, err := s.wordUc.GetTodayWords(ctx, req.StudentId, req.Date)
	if err != nil {
		return &v1.GetTodayWordsReply{
			Ret: &v1.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	// 转换响应数据
	v1Words := make([]*v1.Word, 0, len(words))
	for _, word := range words {
		v1Words = append(v1Words, &v1.Word{
			Id:             word.ID,
			StudentId:      word.StudentID,
			Word:           word.Word,
			Meaning:        word.Meaning,
			StartDate:      word.StartDate,
			ReviewCount:    word.ReviewCount,
			LastReviewDate: word.LastReviewDate,
			CreatedAt:      word.CreatedAt.Unix(),
			UpdatedAt:      word.UpdatedAt.Unix(),
		})
	}

	// 确定日期
	date := req.Date
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	return &v1.GetTodayWordsReply{
		Ret: &v1.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Words: v1Words,
		Date:  date,
		Count: int32(len(v1Words)),
	}, nil
}

// MarkWordReviewed 标记单词为已复习
func (s *StudentService) MarkWordReviewed(ctx context.Context, req *v1.MarkWordReviewedRequest) (*v1.MarkWordReviewedReply, error) {
	word, err := s.wordUc.MarkWordReviewed(ctx, req.StudentId, req.WordId)
	if err != nil {
		return &v1.MarkWordReviewedReply{
			Ret: &v1.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	return &v1.MarkWordReviewedReply{
		Ret: &v1.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Word: &v1.Word{
			Id:             word.ID,
			StudentId:      word.StudentID,
			Word:           word.Word,
			Meaning:        word.Meaning,
			StartDate:      word.StartDate,
			ReviewCount:    word.ReviewCount,
			LastReviewDate: word.LastReviewDate,
			CreatedAt:      word.CreatedAt.Unix(),
			UpdatedAt:      word.UpdatedAt.Unix(),
		},
	}, nil
}

// GetOrCreateStudentByOpenid 通过openid获取或创建学生
func (s *StudentService) GetOrCreateStudentByOpenid(ctx context.Context, req *v1.GetOrCreateStudentByOpenidRequest) (*v1.GetOrCreateStudentByOpenidReply, error) {
	student, isNew, err := s.studentUc.GetOrCreateStudentByOpenid(ctx, req.Openid, req.Name)
	if err != nil {
		return &v1.GetOrCreateStudentByOpenidReply{
			Ret: &v1.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	return &v1.GetOrCreateStudentByOpenidReply{
		Ret: &v1.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Student: &v1.Student{
			Id:        student.ID,
			Name:      student.Name,
			StudentNo: student.StudentNo,
			Openid:    student.OpenID,
			CreatedAt: student.CreatedAt.Unix(),
			UpdatedAt: student.UpdatedAt.Unix(),
		},
		IsNew: isNew,
	}, nil
}
