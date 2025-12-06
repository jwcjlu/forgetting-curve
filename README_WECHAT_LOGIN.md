# 微信登录集成说明

## 功能说明

小程序已实现自动获取用户微信号（openid）作为学生ID的功能。用户首次打开小程序时会自动登录，无需手动输入学生ID。

## 实现方式

### 当前实现（简化版）

当前实现使用 `wx.login()` 获取的 `code` 作为临时标识。这种方式在开发环境可以使用，但在生产环境需要：

1. **配置微信小程序AppID和AppSecret**
2. **后端实现code到openid的转换**

### 生产环境配置

#### 1. 后端配置

在后端添加微信API调用功能，将code转换为openid：

```go
// 示例：在backend/internal/biz/wechat.go中实现
func GetOpenidByCode(code string) (string, error) {
    // 调用微信API: https://api.weixin.qq.com/sns/jscode2session
    // 需要AppID和AppSecret
    // 返回openid
}
```

#### 2. 修改小程序登录流程

修改 `utils/auth.js` 中的 `wxLogin` 函数：

```javascript
async function wxLogin() {
  const code = await new Promise((resolve, reject) => {
    wx.login({
      success: (res) => resolve(res.code),
      fail: reject
    });
  });
  
  // 调用后端接口，将code转换为openid
  const response = await api.request({
    url: '/api/v1/wechat/code2openid',
    method: 'POST',
    data: { code }
  });
  
  return response.openid;
}
```

#### 3. 后端添加接口

在 `backend/api/student/v1/student.proto` 中添加：

```protobuf
rpc Code2Openid (Code2OpenidRequest) returns (Code2OpenidReply) {
  option (google.api.http) = {
    post: "/api/v1/wechat/code2openid"
    body: "*"
  };
}
```

## 使用流程

### 用户首次使用

1. 打开小程序
2. 自动调用 `wx.login()` 获取code
3. 调用后端接口获取或创建学生
4. 自动保存学生ID
5. 可以正常使用所有功能

### 已登录用户

1. 打开小程序
2. 自动使用已保存的学生ID
3. 直接加载数据

## 数据模型

### Student表新增字段

- `open_id`: varchar(100), 唯一索引
- `student_no`: 改为可选（可以为空）

### 自动生成学号

如果学生通过微信登录创建，系统会自动生成学号：
- 格式：`wx_` + openid前8位
- 例如：`wx_12345678`

## 注意事项

1. **开发环境**：当前使用code作为临时标识，可以正常开发测试
2. **生产环境**：必须配置微信AppID和AppSecret，实现code到openid的转换
3. **用户隐私**：openid是用户的唯一标识，需要妥善保管
4. **兼容性**：保留手动设置学生ID的功能，兼容旧用户

## 配置步骤

### 1. 获取微信小程序AppID和AppSecret

1. 登录[微信公众平台](https://mp.weixin.qq.com/)
2. 进入"开发" -> "开发管理" -> "开发设置"
3. 获取AppID和AppSecret

### 2. 后端配置

在 `backend/configs/config.yaml` 中添加：

```yaml
wechat:
  app_id: "your_app_id"
  app_secret: "your_app_secret"
```

### 3. 实现code转换接口

参考上面的代码示例，实现code到openid的转换。

## 测试

1. 清除小程序缓存
2. 重新打开小程序
3. 应该自动登录并创建学生
4. 检查数据库中是否有新学生记录

