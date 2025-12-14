# 验证微信配置

## 检查步骤

1. **确认 AppID 和 AppSecret**
   - 登录微信公众平台：https://mp.weixin.qq.com
   - 进入"开发" -> "开发管理" -> "开发设置"
   - 查看"AppID(小程序ID)" 和 "AppSecret(小程序密钥)"
   - 确保与 `config.docker.yaml` 中的配置完全一致

2. **检查配置格式**
   ```yaml
   wechat:
     app_id: wx2c2b1fd1b7d586c0      # 必须与微信公众平台中的 AppID 完全一致
     app_secret: f3e9987492fcef649c13cc3a1cdecbb3  # 必须与微信公众平台中的 AppSecret 完全一致
   ```

3. **测试微信 API 连接**
   ```bash
   # 在服务器上测试是否能访问微信 API
   curl "https://api.weixin.qq.com/sns/jscode2session?appid=YOUR_APPID&secret=YOUR_SECRET&js_code=TEST_CODE&grant_type=authorization_code"
   ```

4. **常见错误码说明**
   - `40029`: code 无效或已过期
     - 可能原因：AppID/AppSecret 错误、code 过期、code 被重复使用
   - `40013`: AppID 无效
     - 可能原因：AppID 配置错误
   - `40125`: AppSecret 无效
     - 可能原因：AppSecret 配置错误

5. **调试建议**
   - 查看后端日志中的 `appid` 和 `code_length`，确认配置正确
   - 检查网络连接，确保服务器能访问 `https://api.weixin.qq.com`
   - 确认小程序是否已发布或使用测试号




