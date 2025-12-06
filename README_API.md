# API 使用说明

## 配置

在小程序中使用后端API前，需要：

1. **修改API地址**：在 `utils/api.js` 中修改 `API_BASE_URL` 为你的后端服务地址
   ```javascript
   const API_BASE_URL = 'http://your-backend-url:8000';
   ```

2. **设置学生ID**：
   - 打开小程序，进入"设置"页面
   - 输入学生ID并保存
   - 学生ID需要先在后端创建学生后获得

## API 接口

### 1. 获取今日单词

**接口**: `GET /api/v1/students/{student_id}/words/today?date=YYYY-MM-DD`

**说明**: 根据艾宾浩斯遗忘曲线算法，返回指定学生今天需要背诵的单词列表

**参数**:
- `student_id`: 学生ID（路径参数）
- `date`: 日期，格式 YYYY-MM-DD（可选，默认为今天）

**响应**:
```json
{
  "words": [
    {
      "id": 1,
      "student_id": 1,
      "word": "hello",
      "meaning": "你好",
      "start_date": "2024-01-01",
      "review_count": 0,
      "last_review_date": "",
      "created_at": 1699000000,
      "updated_at": 1699000000
    }
  ],
  "date": "2024-01-15",
  "count": 1
}
```

### 2. 标记单词为已复习

**接口**: `POST /api/v1/students/{student_id}/words/{word_id}/review`

**说明**: 标记单词为已复习，会自动增加复习次数并更新最后复习日期

**参数**:
- `student_id`: 学生ID（路径参数）
- `word_id`: 单词ID（路径参数）

**响应**:
```json
{
  "word": {
    "id": 1,
    "student_id": 1,
    "word": "hello",
    "meaning": "你好",
    "start_date": "2024-01-01",
    "review_count": 1,
    "last_review_date": "2024-01-15",
    "created_at": 1699000000,
    "updated_at": 1699000000
  }
}
```

### 3. 批量添加单词

**接口**: `POST /api/v1/students/{student_id}/words/batch`

**说明**: 批量添加单词到指定学生名下

**请求体**:
```json
{
  "words": [
    {
      "word": "hello",
      "meaning": "你好",
      "start_date": "2024-01-01"
    }
  ]
}
```

### 4. 获取学生信息

**接口**: `GET /api/v1/students/{id}`

**说明**: 根据学生ID获取学生信息

## 小程序使用流程

1. **首次使用**：
   - 打开小程序
   - 如果提示需要设置学生ID，点击"设置"
   - 输入学生ID（需要先在后端创建学生）
   - 保存后返回主页面

2. **查看今日单词**：
   - 主页面会自动从后端加载今日需要背诵的单词
   - 单词列表根据艾宾浩斯遗忘曲线算法自动计算

3. **复习单词**：
   - 点击"开始拼写"进行拼写练习
   - 完成后点击"已复习"标记为已复习
   - 系统会自动更新复习次数

## 注意事项

1. **网络配置**：确保小程序可以访问后端服务地址
2. **学生ID**：必须先在后端创建学生，获得学生ID后才能使用
3. **数据同步**：所有数据都保存在后端，卸载小程序不会丢失数据
4. **权限控制**：每个学生只能查看和操作自己的单词

