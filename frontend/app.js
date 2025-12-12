// app.js
const auth = require('./utils/auth.js');
const api = require('./utils/api.js');

App({
  onLaunch() {
    // 自动登录
    this.autoLogin();
  },

  /**
   * 自动登录
   */
  async autoLogin() {
    // 检查是否已有学生ID
    if (api.getStudentId()) {
      console.log('[AutoLogin] 已有学生ID，跳过登录');
      return;
    }

    try {
      wx.showLoading({
        title: '登录中...',
        mask: true
      });

      console.log('[AutoLogin] 开始自动登录...');
      const result = await auth.autoLogin();
      
      wx.hideLoading();
      
      if (result.isNew) {
        console.log('[AutoLogin] 新用户注册成功', result.student);
        wx.showToast({
          title: '注册成功',
          icon: 'success',
          duration: 1500
        });
      } else {
        console.log('[AutoLogin] 登录成功', result.student);
      }
    } catch (err) {
      wx.hideLoading();
      console.error('[AutoLogin] 自动登录失败:', err);
      
      // 显示错误提示
      wx.showModal({
        title: '登录失败',
        content: err.message || '登录失败，请检查网络连接和API配置',
        showCancel: true,
        cancelText: '稍后重试',
        confirmText: '去设置',
        success: (res) => {
          if (res.confirm) {
            wx.switchTab({
              url: '/pages/settings/settings'
            });
          }
        }
      });
    }
  },

  globalData: {
    userInfo: null
  }
});
