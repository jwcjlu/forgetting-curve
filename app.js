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
      return;
    }

    try {
      wx.showLoading({
        title: '登录中...',
        mask: true
      });

      const result = await auth.autoLogin();
      
      wx.hideLoading();
      
      if (result.isNew) {
        console.log('新用户注册成功');
      } else {
        console.log('登录成功');
      }
    } catch (err) {
      wx.hideLoading();
      console.error('自动登录失败:', err);
      // 登录失败不影响使用，用户可以在设置页面手动设置
    }
  },

  globalData: {
    userInfo: null
  }
});

