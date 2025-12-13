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
	GetByIDs(ctx context.Context, wordIDs []int64) ([]*Word, error)
}

// PlanRepo 计划数据仓库接口
type PlanRepo interface {
	Create(ctx context.Context, plan *Plan) (*Plan, error)
	GetByID(ctx context.Context, id int64) (*Plan, error)
	GetByStudentID(ctx context.Context, studentID int64) ([]*Plan, error)
	GetActivePlanByStudentID(ctx context.Context, studentID int64) (*Plan, error)
	Update(ctx context.Context, plan *Plan) (*Plan, error)
	Delete(ctx context.Context, id int64) error
}

// PlanWordRepo 计划单词关联数据仓库接口
type PlanWordRepo interface {
	BatchCreate(ctx context.Context, planWords []*PlanWord) error
	GetByPlanID(ctx context.Context, planID int64) ([]*PlanWord, error)
	GetWordIDsByPlanID(ctx context.Context, planID int64) ([]int64, error)
	DeleteByPlanIDAndWordID(ctx context.Context, planID int64, wordID int64) error
	DeleteByPlanID(ctx context.Context, planID int64) error
}

// OCRService OCR识别服务接口
type OCRService interface {
	RecognizeText(ctx context.Context, req *OCRRequest) (*OCRResult, error)
}

// LLMService 大模型服务接口（已在llm.go中定义，这里只是引用）
// type LLMService interface {
// 	GenerateReviewQuestions(ctx context.Context, word, meaning string) ([]*ReviewQuestion, error)
// }

// RecognizedWord OCR识别的单词
type RecognizedWord struct {
	Word       string
	Meaning    string
	Confidence float64
}
type OCRRequest struct {
	Base64  string                 `json:"base64"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// OCRResult 定义 OCR 识别结果结构
type OCRResult struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data"` // 可能是 []OCRTextData (dict格式) 或 string (text格式)
}

// OCRTextData OCR文本数据（当 data.format 为 dict 时使用）
type OCRTextData struct {
	Box   [][]int `json:"box"`   // 文本框顺时针四个角的xy坐标：[左上,右上,右下,左下]
	Score float64 `json:"score"` // 置信度 (0~1)
	Text  string  `json:"text"`  // 文本
	End   string  `json:"end"`   // 表示本行文字结尾的结束符，可能为空、空格或换行\n
}
