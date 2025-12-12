# ⚠️ 重要配置说明

## AppID 和 AppSecret 配置

### 问题根源
`invalid code` 错误的**最常见原因**是 AppID/AppSecret 配置错误或不匹配。

### 配置要求

1. **AppID 必须一致**
   - 小程序 `project.config.json` 中的 `appid`
   - 后端 `config.docker.yaml` 中的 `wechat.app_id`
   - **这两个必须完全一致！**

2. **AppSecret 必须匹配**
   - 后端配置的 `app_secret` 必须与微信公众平台中对应 AppID 的 AppSecret 完全一致
   - AppID 和 AppSecret 必须来自同一个小程序

### 当前配置状态

**小程序 AppID** (project.config.json):
```
appid: "wxe66675d3364a6e2e"
```

**后端 AppID** (config.docker.yaml):
```
app_id: wxe66675d3364a6e2e  # 已更新为与小程序一致
```

### 配置步骤

1. **获取 AppSecret**
   - 登录微信公众平台：https://mp.weixin.qq.com
   - 进入"开发" → "开发管理" → "开发设置"
   - 查看"AppSecret(小程序密钥)"
   - 如果看不到，点击"重置"获取新的 AppSecret

2. **更新后端配置**
   ```yaml
   wechat:
     app_id: wxe66675d3364a6e2e  # 与小程序 AppID 一致
     app_secret: YOUR_APP_SECRET_HERE  # 替换为实际的 AppSecret
   ```

3. **验证配置**
   - 确保 AppID 与小程序完全一致
   - 确保 AppSecret 与微信公众平台完全一致
   - 确保没有多余的空格或引号

4. **重启服务**
   - 更新配置后，必须重启后端服务
   - 清除小程序缓存，重新测试

### 验证清单

- [ ] 小程序 `project.config.json` 中的 `appid` 已确认
- [ ] 后端 `config.docker.yaml` 中的 `app_id` 与小程序 AppID 一致
- [ ] 后端 `config.docker.yaml` 中的 `app_secret` 已更新为正确的值
- [ ] AppSecret 来自微信公众平台，与 AppID 匹配
- [ ] 已重启后端服务
- [ ] 已清除小程序缓存

### 常见错误

❌ **错误配置示例**：
```yaml
wechat:
  app_id: wx2c2b1fd1b7d586c0  # 与小程序 AppID 不一致！
  app_secret: f3e9987492fcef649c13cc3a1cdecbb3
```

✅ **正确配置示例**：
```yaml
wechat:
  app_id: wxe66675d3364a6e2e  # 与小程序 AppID 一致
  app_secret: abc123def456...  # 与 AppID 匹配的 AppSecret
```

### 调试信息

查看后端日志，应该看到：
```
Calling WeChat API: appid=wxe66675d3364a6e2e, secret_length=32, secret_prefix=...
```

- `appid` 应该与小程序 AppID 完全一致
- `secret_length=32` 表示 AppSecret 长度正确
- `secret_prefix` 应该与配置的前 10 位匹配



