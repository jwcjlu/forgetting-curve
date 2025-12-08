# 故障排查指南 - 微信登录 invalid code 错误

## 错误现象
```
wechat api error [40029]: code 无效或已过期
```

## 最可能的原因
**AppID/AppSecret 配置错误或不匹配**（99% 的情况）

## 解决步骤

### 1. 验证微信公众平台配置

1. 登录微信公众平台：https://mp.weixin.qq.com
2. 进入"开发" → "开发管理" → "开发设置"
3. 查看并记录：
   - **AppID(小程序ID)**: `wx...`
   - **AppSecret(小程序密钥)**: `...` (32位字符串)

### 2. 检查配置文件

编辑 `backend/configs/config.docker.yaml`:

```yaml
wechat:
  app_id: wx2c2b1fd1b7d586c0      # 必须与微信公众平台中的 AppID 完全一致
  app_secret: f3e9987492fcef649c13cc3a1cdecbb3  # 必须与微信公众平台中的 AppSecret 完全一致
```

**重要检查项**：
- ✅ 没有多余的空格
- ✅ 没有引号（YAML 中不需要引号）
- ✅ 大小写完全一致
- ✅ AppID 和 AppSecret 来自同一个小程序

### 3. 如果看不到 AppSecret

如果看不到 AppSecret，需要重置：

1. 在微信公众平台点击"重置"
2. 按照提示完成验证
3. 获取新的 AppSecret
4. 更新配置文件
5. **重启后端服务**

### 4. 验证配置是否正确加载

查看后端日志，应该看到：
```
Calling WeChat API: appid=wx2c2b1fd1b7d586c0, secret_length=32, secret_prefix=...
```

- `secret_length=32` 表示 AppSecret 长度正确（通常是 32 位）
- `secret_prefix` 应该与配置的前 10 位匹配

### 5. 测试微信 API（可选）

可以使用测试脚本验证配置：

```bash
cd backend/scripts
chmod +x test-wechat-api.sh
./test-wechat-api.sh <appid> <secret> <code>
```

## 其他可能的原因（较少见）

### 原因 2: code 被重复使用
- **现象**: 同一个 code 被多次使用
- **解决**: 代码已添加锁机制，确保 code 只使用一次

### 原因 3: code 过期
- **现象**: code 获取后超过 5 分钟才使用
- **解决**: 代码已优化，确保 code 获取后立即使用

### 原因 4: 网络问题
- **现象**: 服务器无法访问微信 API
- **解决**: 检查服务器网络，确保能访问 `https://api.weixin.qq.com`

## 常见错误码

| 错误码 | 含义 | 解决方法 |
|--------|------|----------|
| 40029 | code 无效或已过期 | 检查 AppID/AppSecret 配置 |
| 40013 | AppID 无效 | 检查 AppID 是否正确 |
| 40125 | AppSecret 无效 | 检查 AppSecret 是否正确 |
| 40163 | code 已被使用 | 重新获取 code |

## 验证清单

- [ ] AppID 与微信公众平台完全一致
- [ ] AppSecret 与微信公众平台完全一致
- [ ] AppID 和 AppSecret 来自同一个小程序
- [ ] 配置文件格式正确（YAML，无多余空格）
- [ ] 已重启后端服务
- [ ] 后端日志显示 `secret_length=32`
- [ ] 网络可以访问微信 API

## 如果问题仍然存在

1. 在微信公众平台重置 AppSecret
2. 更新配置文件
3. 重启后端服务
4. 清除小程序缓存，重新测试

如果以上步骤都完成但问题仍然存在，请提供：
- 后端日志中的 `appid` 和 `secret_prefix`
- 微信公众平台中的 AppID（前 10 位）
- 配置文件中的 AppSecret（前 10 位）

