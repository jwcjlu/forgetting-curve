# 微信小程序前端 - 使用说明

## 项目结构

```
.
├── app.js                 # 小程序入口文件
├── app.json              # 小程序配置文件
├── app.wxss              # 全局样式
├── pages/                 # 页面目录
│   ├── index/            # 今日背诵页面
│   │   ├── index.js
│   │   ├── index.wxml
│   │   └── index.wxss
│   ├── add-word/         # 添加单词页面
│   │   ├── add-word.js
│   │   ├── add-word.wxml
│   │   └── add-word.wxss
│   └── settings/         # 设置页面
│       ├── settings.js
│       ├── settings.wxml
│       └── settings.wxss
├── utils/                # 工具函数
│   ├── api.js           # API调用工具（基于proto协议）
│   ├── auth.js          # 微信登录工具
│   └── ebbinghaus.js    # 日期格式化工具
└── project.config.json  # 项目配置
```

## API 接口说明（基于 proto 协议）

### 1. GetOrCreateStudentByOpenid - 微信登录
- **路径**: `POST /api/v1/students/by-openid`
- **功能**: 通过 openid 获取或创建学生
- **请求**: `{ openid: string, name: string }`
- **响应**: `{ ret: BaseResponse, student: Student, is_new: bool }`

### 2. GetStudent - 获取学生信息
- **路径**: `GET /api/v1/students/{id}`
- **功能**: 根据学生ID获取学生信息
- **响应**: `{ ret: BaseResponse, student: Student }`

### 3. GetTodayWords - 获取今日单词
- **路径**: `GET /api/v1/students/{student_id}/words/today?date=YYYY-MM-DD`
- **功能**: 根据艾宾浩斯曲线获取今日需要背诵的单词
- **响应**: `{ ret: BaseResponse, date: string, count: int32, words: Word[] }`

### 4. MarkWordReviewed - 标记已复习
- **路径**: `POST /api/v1/students/{student_id}/words/{word_id}/review`
- **功能**: 标记单词为已复习
- **响应**: `{ ret: BaseResponse, word: Word }`

### 5. BatchAddWords - 批量添加单词
- **路径**: `POST /api/v1/students/{student_id}/words/batch`
- **功能**: 批量添加单词到学生名下
- **请求**: `{ words: WordItem[] }`
- **响应**: `{ ret: BaseResponse, count: int32, words: Word[] }`

### 6. GetStudentWords - 获取单词列表
- **路径**: `GET /api/v1/students/{student_id}/words?page=1&page_size=20`
- **功能**: 获取学生的单词列表（分页）
- **响应**: `{ ret: BaseResponse, total: int32, page: int32, page_size: int32, words: Word[] }`

## 数据模型（基于 proto）

### Student
```javascript
{
  id: number,           // 学生ID
  name: string,         // 姓名
  student_no: string,   // 学号
  openid: string,       // 微信openid
  created_at: number,   // 创建时间（Unix时间戳）
  updated_at: number    // 更新时间（Unix时间戳）
}
```

### Word
```javascript
{
  id: number,              // 单词ID
  student_id: number,      // 学生ID
  word: string,            // 单词
  meaning: string,         // 释义
  start_date: string,      // 开始日期 YYYY-MM-DD
  review_count: number,    // 复习次数
  last_review_date: string, // 上次复习日期
  created_at: number,      // 创建时间
  updated_at: number       // 更新时间
}
```

### BaseResponse
```javascript
{
  code: number,      // 错误码，0表示成功
  message: string,   // 错误消息
  reason: string     // 错误原因
}
```

## 配置说明

### 1. 修改 API 地址

在 `utils/api.js` 中修改：
```javascript
const API_BASE_URL = 'http://your-backend-url:8000';
```

或者在设置页面中配置（会保存到本地存储）。

### 2. 配置小程序 AppID

在 `project.config.json` 中修改 `appid` 为你的小程序 AppID。

### 3. 配置合法域名

在微信公众平台配置服务器域名：
- 开发环境：可以在开发者工具中勾选"不校验合法域名"
- 生产环境：需要在微信公众平台配置 request 合法域名

## 功能说明

### 1. 自动登录
- 小程序启动时自动调用微信登录
- 使用 openid 作为学生标识
- 自动创建学生记录（如果不存在）

### 2. 今日背诵
- 自动从后端获取今日需要背诵的单词
- 根据艾宾浩斯遗忘曲线算法计算
- 支持拼写练习模式
- 标记单词为已复习

### 3. 添加单词
- 支持单个添加单词
- 显示所有已添加的单词列表
- 支持分页加载

### 4. 设置
- 查看学生信息
- 手动设置学生ID（兼容旧用户）
- 配置API地址

## 错误处理

所有 API 响应都包含 `BaseResponse`，需要检查 `ret.code`：
- `code === 0`: 成功
- `code !== 0`: 失败，显示 `ret.message`

## 使用流程

1. **首次使用**：
   - 打开小程序
   - 自动调用微信登录
   - 自动创建学生记录
   - 可以开始使用

2. **添加单词**：
   - 进入"添加单词"页面
   - 输入单词、释义和开始日期
   - 点击"添加单词"

3. **背诵单词**：
   - 进入"今日背诵"页面
   - 查看今日需要复习的单词
   - 点击"开始拼写"进行练习
   - 完成后点击"已复习"

## 注意事项

1. **网络配置**：确保小程序可以访问后端服务
2. **数据同步**：所有数据保存在后端，卸载小程序不会丢失
3. **权限控制**：每个学生只能查看和操作自己的单词
4. **API地址**：开发环境可以使用 localhost，生产环境需要配置合法域名

