# 测试用例说明

本目录包含了业务逻辑层的单元测试用例。

## 测试文件

### 1. `ebbinghaus_test.go`
测试艾宾浩斯遗忘曲线算法：
- `GetDaysBetween`: 计算两个日期之间的天数差
- `ShouldReviewToday`: 判断单词是否需要在今天复习
- `FilterTodayWords`: 过滤出今天需要复习的单词
- `REVIEW_INTERVALS`: 验证复习间隔数组

### 2. `llm_test.go`
测试大模型服务：
- `NewLLMService`: 测试LLM服务的创建
- `GenerateReviewQuestions`: 测试生成复习题目（使用HTTP mock）
- `extractJSON`: 测试从文本中提取JSON
- `getGradeDifficultyHint`: 测试根据年级生成难度提示
- `contains`: 测试字符串包含检查

### 3. `word_test.go`
测试单词业务逻辑：
- `BatchAddWords`: 测试批量添加单词
- `ValidateStudentAccess`: 测试学生权限验证
- `MarkWordReviewed`: 测试标记单词为已复习
- `GenerateReviewQuestions`: 测试生成复习题目

### 4. `confused_word_test.go`
测试混淆词功能：
- `AddConfusedWord`: 测试添加混淆词
- `RemoveConfusedWord`: 测试删除混淆词
- `GetConfusedWords`: 测试获取混淆词列表
- `SearchWords`: 测试搜索单词

## 运行测试

### 运行所有测试
```bash
cd backend
go test ./internal/biz/...
```

### 运行特定测试文件
```bash
go test ./internal/biz/ebbinghaus_test.go ./internal/biz/ebbinghaus.go
```

### 运行特定测试函数
```bash
go test ./internal/biz -run TestShouldReviewToday
```

### 显示详细输出
```bash
go test -v ./internal/biz/...
```

### 显示覆盖率
```bash
go test -cover ./internal/biz/...
```

### 生成覆盖率报告
```bash
go test -coverprofile=coverage.out ./internal/biz/...
go tool cover -html=coverage.out
```

## Mock对象

测试中使用了以下Mock对象：
- `mockWordRepo`: 模拟单词数据仓库
- `mockStudentRepo`: 模拟学生数据仓库
- `mockLLMService`: 模拟LLM服务
- `mockConfusedWordRepo`: 模拟混淆词数据仓库

这些Mock对象实现了相应的接口，用于隔离测试，不依赖真实的数据库或外部服务。

## 注意事项

1. **LLM测试**: `llm_test.go` 使用HTTP测试服务器来模拟LLM API响应，不需要真实的API Key。

2. **日期测试**: `ebbinghaus_test.go` 中的日期测试使用固定日期，确保测试结果可预测。

3. **Mock实现**: Mock对象实现了接口的所有方法，但实现可能比真实实现简化。

4. **测试隔离**: 每个测试用例都是独立的，不会相互影响。

## 添加新测试

添加新测试时，请遵循以下规范：

1. 测试函数名以 `Test` 开头
2. 使用表驱动测试（table-driven tests）提高可读性
3. 为每个测试用例提供清晰的名称
4. 使用 `t.Helper()` 标记辅助函数
5. 测试错误情况，不仅测试正常情况

示例：
```go
func TestNewFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "正常情况",
            input:   "test",
            want:    "result",
            wantErr: false,
        },
        {
            name:    "错误情况",
            input:   "",
            want:    "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := NewFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("NewFunction() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("NewFunction() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

