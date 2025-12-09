package service

import (
	"context"
	"fmt"
	"time"

	v1 "forgetting-curve/backend/api/student/v1"
	"forgetting-curve/backend/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
)

// StudentService 学生服务
type StudentService struct {
	v1.UnimplementedStudentServiceServer

	studentUc      biz.StudentUsecase
	wordUc         biz.WordUsecase
	confusedWordUc biz.ConfusedWordUsecase
	wechatSvc      biz.WechatService
	log            *log.Helper
}

// NewStudentService 创建学生服务
func NewStudentService(
	studentUc biz.StudentUsecase,
	wordUc biz.WordUsecase,
	confusedWordUc biz.ConfusedWordUsecase,
	wechatSvc biz.WechatService,
	logger log.Logger,
) *StudentService {
	return &StudentService{
		studentUc:      studentUc,
		wordUc:         wordUc,
		confusedWordUc: confusedWordUc,
		wechatSvc:      wechatSvc,
		log:            log.NewHelper(logger),
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
		v1Words = append(v1Words, convertWordToV1(word))
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
		Word: convertWordToV1(word),
	}, nil
}

// GetOrCreateStudentByOpenid 通过openid或code获取或创建学生
func (s *StudentService) GetOrCreateStudentByOpenid(ctx context.Context, req *v1.GetOrCreateStudentByOpenidRequest) (*v1.GetOrCreateStudentByOpenidReply, error) {
	var openid string
	var err error

	// 优先使用 code，如果没有 code 则使用 openid
	if req.Code != "" {
		// 调用微信接口换取 openid
		if s.wechatSvc == nil {
			return &v1.GetOrCreateStudentByOpenidReply{
				Ret: &v1.BaseResponse{
					Code:    500,
					Message: "wechat service not configured",
				},
			}, nil
		}

		openid, _, err = s.wechatSvc.Code2Session(ctx, req.Code)
		if err != nil {
			s.log.Errorf("failed to exchange code for openid: %v", err)
			return &v1.GetOrCreateStudentByOpenidReply{
				Ret: &v1.BaseResponse{
					Code:    500,
					Message: fmt.Sprintf("failed to exchange code: %v", err),
				},
			}, nil
		}
		s.log.Infof("successfully exchanged code for openid: %s", openid)
	} else if req.Openid != "" {
		openid = req.Openid
	} else {
		return &v1.GetOrCreateStudentByOpenidReply{
			Ret: &v1.BaseResponse{
				Code:    400,
				Message: "openid or code is required",
			},
		}, nil
	}

	student, isNew, err := s.studentUc.GetOrCreateStudentByOpenid(ctx, openid, req.Name)
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

// AddConfusedWord 添加混淆词
func (s *StudentService) AddConfusedWord(ctx context.Context, req *v1.AddConfusedWordRequest) (*v1.AddConfusedWordReply, error) {
	err := s.confusedWordUc.AddConfusedWord(ctx, req.StudentId, req.WordId, req.ConfusedWordId)
	if err != nil {
		return &v1.AddConfusedWordReply{
			Ret: &v1.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	// 获取更新后的单词
	word, err := s.wordUc.GetWord(ctx, req.StudentId, req.WordId)
	if err != nil {
		return &v1.AddConfusedWordReply{
			Ret: &v1.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	// 获取混淆词列表
	confusedWords, err := s.confusedWordUc.GetConfusedWords(ctx, req.StudentId, req.WordId)
	if err != nil {
		s.log.Warnf("failed to get confused words: %v", err)
	}

	v1Word := convertWordToV1(word)
	if len(confusedWords) > 0 {
		v1Word.ConfusedWords = make([]*v1.Word, 0, len(confusedWords))
		for _, cw := range confusedWords {
			v1Word.ConfusedWords = append(v1Word.ConfusedWords, convertWordToV1(cw))
		}
	}

	return &v1.AddConfusedWordReply{
		Ret: &v1.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Word: v1Word,
	}, nil
}

// GetConfusedWords 获取单词的混淆词列表
func (s *StudentService) GetConfusedWords(ctx context.Context, req *v1.GetConfusedWordsRequest) (*v1.GetConfusedWordsReply, error) {
	confusedWords, err := s.confusedWordUc.GetConfusedWords(ctx, req.StudentId, req.WordId)
	if err != nil {
		return &v1.GetConfusedWordsReply{
			Ret: &v1.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	v1Words := make([]*v1.Word, 0, len(confusedWords))
	for _, word := range confusedWords {
		v1Words = append(v1Words, convertWordToV1(word))
	}

	return &v1.GetConfusedWordsReply{
		Ret: &v1.BaseResponse{
			Code:    0,
			Message: "success",
		},
		ConfusedWords: v1Words,
	}, nil
}

// RemoveConfusedWord 删除混淆词
func (s *StudentService) RemoveConfusedWord(ctx context.Context, req *v1.RemoveConfusedWordRequest) (*v1.RemoveConfusedWordReply, error) {
	err := s.confusedWordUc.RemoveConfusedWord(ctx, req.StudentId, req.WordId, req.ConfusedWordId)
	if err != nil {
		return &v1.RemoveConfusedWordReply{
			Ret: &v1.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	return &v1.RemoveConfusedWordReply{
		Ret: &v1.BaseResponse{
			Code:    0,
			Message: "success",
		},
	}, nil
}

// SearchWords 搜索单词（支持正则表达式）
func (s *StudentService) SearchWords(ctx context.Context, req *v1.SearchWordsRequest) (*v1.SearchWordsReply, error) {
	words, err := s.confusedWordUc.SearchWords(ctx, req.StudentId, req.Keyword, req.Limit)
	if err != nil {
		return &v1.SearchWordsReply{
			Ret: &v1.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	v1Words := make([]*v1.Word, 0, len(words))
	for _, word := range words {
		v1Words = append(v1Words, convertWordToV1(word))
	}

	return &v1.SearchWordsReply{
		Ret: &v1.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Words: v1Words,
	}, nil
}

// MarkWordForgotten 标记单词为未记住
func (s *StudentService) MarkWordForgotten(ctx context.Context, req *v1.MarkWordForgottenRequest) (*v1.MarkWordForgottenReply, error) {
	word, err := s.wordUc.MarkWordForgotten(ctx, req.StudentId, req.WordId)
	if err != nil {
		return &v1.MarkWordForgottenReply{
			Ret: &v1.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	return &v1.MarkWordForgottenReply{
		Ret: &v1.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Word: convertWordToV1(word),
	}, nil
}

// UpdateWordReviewData 更新单词复习数据
func (s *StudentService) UpdateWordReviewData(ctx context.Context, req *v1.UpdateWordReviewDataRequest) (*v1.UpdateWordReviewDataReply, error) {
	word, err := s.wordUc.UpdateWordReviewData(ctx, req.StudentId, req.WordId, req.ThinkTime, req.Difficulty, req.IsRemembered)
	if err != nil {
		return &v1.UpdateWordReviewDataReply{
			Ret: &v1.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	return &v1.UpdateWordReviewDataReply{
		Ret: &v1.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Word: convertWordToV1(word),
	}, nil
}

// convertWordToV1 转换 biz.Word 到 v1.Word
func convertWordToV1(word *biz.Word) *v1.Word {
	return &v1.Word{
		Id:             word.ID,
		StudentId:      word.StudentID,
		Word:           word.Word,
		Meaning:        word.Meaning,
		StartDate:      word.StartDate,
		ReviewCount:    word.ReviewCount,
		LastReviewDate: word.LastReviewDate,
		Difficulty:     word.Difficulty,
		ThinkTime:      word.ThinkTime,
		IsRemembered:   word.IsRemembered,
		ForgetCount:    word.ForgetCount,
		CreatedAt:      word.CreatedAt.Unix(),
		UpdatedAt:      word.UpdatedAt.Unix(),
	}
}
