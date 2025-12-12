# SmartReview 功能实现状态

参照 [SmartReview 项目](https://github.com/EchoShoot/SmartReview) 的需求，已实现以下功能：

## ✅ 已完成的后端功能

### 1. 数据模型扩展
- ✅ 添加多维度数据字段：
  - `difficulty` - 学习难度 (0-10)
  - `think_time` - 思考时间（秒）
  - `is_remembered` - 是否记住
  - `forget_count` - 忘记次数
- ✅ 添加混淆词关联表 `ConfusedWord`

### 2. API 接口（proto 已定义）
- ✅ `AddConfusedWord` - 添加混淆词
- ✅ `GetConfusedWords` - 获取单词的混淆词列表
- ✅ `RemoveConfusedWord` - 删除混淆词
- ✅ `SearchWords` - 搜索单词（支持正则表达式）
- ✅ `MarkWordForgotten` - 标记单词为未记住
- ✅ `UpdateWordReviewData` - 更新单词复习数据（思考时间、难度等）

### 3. 业务逻辑
- ✅ 混淆词管理逻辑（`internal/biz/confused_word.go`）
- ✅ 单词复习数据更新逻辑
- ✅ 单词搜索逻辑（支持正则表达式）

## ⚠️ 需要执行的步骤

### 1. 生成 Proto 代码
```bash
cd backend
make api
```

### 2. 生成 Wire 代码
```bash
cd backend/cmd/server
wire
```

### 3. 更新数据库
重启后端服务，GORM 会自动创建新表：
- `confused_words` - 混淆词关联表
- 更新 `words` 表，添加新字段

## 📋 待实现的前端功能

### 1. 改进复习流程（参照 SmartReview）
- [ ] 按住键显示单词，松开显示含义
- [ ] 标记未记住功能（空格键）
- [ ] 显示混淆词列表
- [ ] 添加混淆词功能

### 2. 多维度数据采集
- [ ] 记录思考时间
- [ ] 记录学习难度
- [ ] 显示忘记次数

### 3. 词库导入功能
- [ ] 支持 .txt 格式导入
- [ ] 支持 .json 格式导入
- [ ] 批量导入单词

## 🎯 SmartReview 核心功能对比

| 功能 | SmartReview | 本项目状态 |
|------|-------------|-----------|
| 艾宾浩斯遗忘曲线 | ✅ | ✅ 已实现 |
| 混淆词关联 | ✅ | ✅ 后端已实现 |
| 按住显示单词 | ✅ | ⏳ 待实现 |
| 标记未记住 | ✅ | ✅ 后端已实现 |
| 思考时间采集 | ✅ | ✅ 后端已实现 |
| 学习难度采集 | ✅ | ✅ 后端已实现 |
| 词库导入 | ✅ | ⏳ 待实现 |

## 📝 下一步工作

1. **生成 Proto 代码**：运行 `make api` 生成新的 API 代码
2. **更新前端**：实现按住显示单词、添加混淆词等功能
3. **测试**：测试所有新功能



