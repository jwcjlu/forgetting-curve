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

	studentUc        biz.StudentUsecase
	wordUc           biz.WordUsecase
	confusedWordUc   biz.ConfusedWordUsecase
	planUc           biz.PlanUsecase
	wechatSvc        biz.WechatService
	ocrSvc           biz.OCRService
	pronunciationSvc biz.PronunciationService
	log              *log.Helper
}

// NewStudentService 创建学生服务
func NewStudentService(
	studentUc biz.StudentUsecase,
	wordUc biz.WordUsecase,
	confusedWordUc biz.ConfusedWordUsecase,
	planUc biz.PlanUsecase,
	wechatSvc biz.WechatService,
	ocrSvc biz.OCRService,
	pronunciationSvc biz.PronunciationService,
	logger log.Logger,
) *StudentService {
	return &StudentService{
		studentUc:        studentUc,
		wordUc:           wordUc,
		confusedWordUc:   confusedWordUc,
		planUc:           planUc,
		wechatSvc:        wechatSvc,
		ocrSvc:           ocrSvc,
		pronunciationSvc: pronunciationSvc,
		log:              log.NewHelper(logger),
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

// BatchAddWords 批量添加单词到计划
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

	// 调用业务逻辑（现在单词直接添加到计划）
	words, err := s.wordUc.BatchAddWords(ctx, req.PlanId, wordItems)
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
		v1Words = append(v1Words, s.convertWordToV1(word))
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

// GetPlanWords 获取计划的单词列表
func (s *StudentService) GetPlanWords(ctx context.Context, req *v1.GetPlanWordsRequest) (*v1.GetPlanWordsReply, error) {
	// 设置默认分页参数
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	// 调用业务逻辑（通过 PlanUsecase 获取计划下的单词）
	words, total, err := s.planUc.GetPlanWords(ctx, req.StudentId, req.PlanId, page, pageSize)
	if err != nil {
		return &v1.GetPlanWordsReply{
			Ret: &v1.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	// 转换响应数据
	v1Words := make([]*v1.Word, 0, len(words))
	for _, word := range words {
		v1Words = append(v1Words, s.convertWordToV1(word))
	}

	return &v1.GetPlanWordsReply{
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
	// 现在单词直接属于计划，必须提供计划ID
	words, err := s.wordUc.GetTodayWords(ctx, req.PlanId, req.Date)
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
		v1Words = append(v1Words, s.convertWordToV1(word))
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
	word, err := s.wordUc.MarkWordReviewed(ctx, req.PlanId, req.WordId)
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
		Word: s.convertWordToV1(word),
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

// AddConfusedWord 添加混淆词AddConfusedWordRequest
func (s *StudentService) AddConfusedWord(ctx context.Context, req *v1.AddConfusedWordRequest) (*v1.AddConfusedWordReply, error) {
	err := s.confusedWordUc.AddConfusedWord(ctx, req.StudentId, req.PlanId, req.WordId, req.ConfusedWordId)
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
	confusedWords, err := s.confusedWordUc.GetConfusedWords(ctx, req.StudentId, req.PlanId, req.WordId)
	if err != nil {
		s.log.Warnf("failed to get confused words: %v", err)
	}

	v1Word := s.convertWordToV1(word)
	if len(confusedWords) > 0 {
		v1Word.ConfusedWords = make([]*v1.Word, 0, len(confusedWords))
		for _, cw := range confusedWords {
			v1Word.ConfusedWords = append(v1Word.ConfusedWords, s.convertWordToV1(cw))
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
	confusedWords, err := s.confusedWordUc.GetConfusedWords(ctx, req.StudentId, req.PlanId, req.WordId)
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
		v1Words = append(v1Words, s.convertWordToV1(word))
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
	err := s.confusedWordUc.RemoveConfusedWord(ctx, req.StudentId, req.PlanId, req.WordId, req.ConfusedWordId)
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
	words, err := s.confusedWordUc.SearchWords(ctx, req.StudentId, req.PlanId, req.Keyword, req.Limit)
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
		v1Words = append(v1Words, s.convertWordToV1(word))
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
	word, err := s.wordUc.MarkWordForgotten(ctx, req.PlanId, req.WordId)
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
		Word: s.convertWordToV1(word),
	}, nil
}

// UpdateWordReviewData 更新单词复习数据
func (s *StudentService) UpdateWordReviewData(ctx context.Context, req *v1.UpdateWordReviewDataRequest) (*v1.UpdateWordReviewDataReply, error) {
	// 确保使用 plan_id 验证权限
	word, err := s.wordUc.UpdateWordReviewData(ctx, req.PlanId, req.WordId, req.ThinkTime, req.Difficulty, req.IsRemembered)
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
		Word: s.convertWordToV1(word),
	}, nil
}

// GenerateReviewQuestions 生成复习题目
func (s *StudentService) GenerateReviewQuestions(ctx context.Context, req *v1.GenerateReviewQuestionsRequest) (*v1.GenerateReviewQuestionsReply, error) {
	grade := ""
	if req.Grade != "" {
		grade = req.Grade
	}

	questions, err := s.wordUc.GenerateReviewQuestions(ctx, req.PlanId, req.WordId, grade)
	if err != nil {
		return &v1.GenerateReviewQuestionsReply{
			Ret: &v1.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	// 转换题目
	v1Questions := make([]*v1.ReviewQuestion, 0, len(questions))
	for _, q := range questions {
		v1Questions = append(v1Questions, &v1.ReviewQuestion{
			Type:          q.Type,
			Question:      q.Question,
			Options:       q.Options,
			CorrectAnswer: q.CorrectAnswer,
		})
	}

	return &v1.GenerateReviewQuestionsReply{
		Ret: &v1.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Questions: v1Questions,
	}, nil
}

// convertWordToV1 转换 biz.Word 到 v1.Word
func (s *StudentService) convertWordToV1(word *biz.Word) *v1.Word {
	v1Word := &v1.Word{
		Id:             word.ID,
		PlanId:         word.PlanID,
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

	// 添加音频 URL
	if s.pronunciationSvc != nil {
		audioURLs := s.pronunciationSvc.GetAudioURLs(word.Word)
		if len(audioURLs) > 0 {
			v1Word.AudioUrls = audioURLs
		}
	}

	return v1Word
}

// CreatePlan 创建复习计划
func (s *StudentService) CreatePlan(ctx context.Context, req *v1.CreatePlanRequest) (*v1.CreatePlanReply, error) {
	plan, err := s.planUc.CreatePlan(ctx, req.StudentId, req.Name, req.Grade)
	if err != nil {
		return &v1.CreatePlanReply{
			Ret: &v1.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	return &v1.CreatePlanReply{
		Ret: &v1.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Plan: &v1.Plan{
			Id:        plan.ID,
			StudentId: plan.StudentID,
			Name:      plan.Name,
			Grade:     plan.Grade,
			IsActive:  plan.IsActive,
			CreatedAt: plan.CreatedAt.Unix(),
			UpdatedAt: plan.UpdatedAt.Unix(),
		},
	}, nil
}

// GetPlans 获取学生的所有计划
func (s *StudentService) GetPlans(ctx context.Context, req *v1.GetPlansRequest) (*v1.GetPlansReply, error) {
	plans, err := s.planUc.GetPlans(ctx, req.StudentId)
	if err != nil {
		return &v1.GetPlansReply{
			Ret: &v1.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	v1Plans := make([]*v1.Plan, 0, len(plans))
	for _, plan := range plans {
		v1Plans = append(v1Plans, &v1.Plan{
			Id:        plan.ID,
			StudentId: plan.StudentID,
			Name:      plan.Name,
			Grade:     plan.Grade,
			IsActive:  plan.IsActive,
			CreatedAt: plan.CreatedAt.Unix(),
			UpdatedAt: plan.UpdatedAt.Unix(),
		})
	}

	return &v1.GetPlansReply{
		Ret: &v1.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Plans: v1Plans,
	}, nil
}

// SelectPlan 选择计划
func (s *StudentService) SelectPlan(ctx context.Context, req *v1.SelectPlanRequest) (*v1.SelectPlanReply, error) {
	plan, err := s.planUc.SelectPlan(ctx, req.StudentId, req.PlanId)
	if err != nil {
		return &v1.SelectPlanReply{
			Ret: &v1.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	return &v1.SelectPlanReply{
		Ret: &v1.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Plan: &v1.Plan{
			Id:        plan.ID,
			StudentId: plan.StudentID,
			Name:      plan.Name,
			Grade:     plan.Grade,
			IsActive:  plan.IsActive,
			CreatedAt: plan.CreatedAt.Unix(),
			UpdatedAt: plan.UpdatedAt.Unix(),
		},
	}, nil
}

// GetPlanWords 获取计划中的单词列表（通过 PlanUsecase）
func (s *StudentService) GetPlanWordsViaPlan(ctx context.Context, req *v1.GetPlanWordsRequest) (*v1.GetPlanWordsReply, error) {
	// 设置默认分页参数
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	words, total, err := s.planUc.GetPlanWords(ctx, req.StudentId, req.PlanId, page, pageSize)
	if err != nil {
		return &v1.GetPlanWordsReply{
			Ret: &v1.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	v1Words := make([]*v1.Word, 0, len(words))
	for _, word := range words {
		v1Words = append(v1Words, s.convertWordToV1(word))
	}

	return &v1.GetPlanWordsReply{
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

// UpdatePlan 更新计划
func (s *StudentService) UpdatePlan(ctx context.Context, req *v1.UpdatePlanRequest) (*v1.UpdatePlanReply, error) {
	plan, err := s.planUc.UpdatePlan(ctx, req.StudentId, req.PlanId, req.Name, req.Grade)
	if err != nil {
		return &v1.UpdatePlanReply{
			Ret: &v1.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	return &v1.UpdatePlanReply{
		Ret: &v1.BaseResponse{
			Code:    0,
			Message: "success",
		},
		Plan: &v1.Plan{
			Id:        plan.ID,
			StudentId: plan.StudentID,
			Name:      plan.Name,
			Grade:     plan.Grade,
			IsActive:  plan.IsActive,
			CreatedAt: plan.CreatedAt.Unix(),
			UpdatedAt: plan.UpdatedAt.Unix(),
		},
	}, nil
}

// DeletePlan 删除计划
func (s *StudentService) DeletePlan(ctx context.Context, req *v1.DeletePlanRequest) (*v1.DeletePlanReply, error) {
	err := s.planUc.DeletePlan(ctx, req.StudentId, req.PlanId)
	if err != nil {
		return &v1.DeletePlanReply{
			Ret: &v1.BaseResponse{
				Code:    500,
				Message: err.Error(),
			},
		}, nil
	}

	return &v1.DeletePlanReply{
		Ret: &v1.BaseResponse{
			Code:    0,
			Message: "success",
		},
	}, nil
}
