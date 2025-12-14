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
    // 即使本地有 studentId，也要验证，确保不同设备使用同一个账号
    const localStudentId = api.getStudentId();
    if (localStudentId) {
      console.log('[AutoLogin] 本地已有学生ID:', localStudentId, '，验证登录以确保账号一致性');
    } else {
      console.log('[AutoLogin] 本地无学生ID，开始登录...');
    }

    try {
      wx.showLoading({
        title: '登录中...',
        mask: true
      });

      console.log('[AutoLogin] 开始自动登录...');
      const result = await auth.autoLogin();
      
      wx.hideLoading();
      
      // 检查返回的学生ID是否与本地存储的不同
      if (result && result.student && result.student.id) {
        const returnedStudentId = result.student.id;
        if (localStudentId && localStudentId !== returnedStudentId) {
          console.warn('[AutoLogin] 检测到学生ID不一致！本地:', localStudentId, '服务器:', returnedStudentId);
          console.log('[AutoLogin] 已更新为学生ID:', returnedStudentId);
          // 本地存储的 studentId 会被 auth.autoLogin 中的 setStudentId 更新
        }
      }
      
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
      
      // 如果登录失败且本地有 studentId，清除本地存储，下次重新登录
      if (localStudentId) {
        console.warn('[AutoLogin] 登录失败，清除本地 studentId，下次重新登录');
        api.clearStudentId();
      }
      
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
