// utils/auth.js
// 微信登录和认证工具

const api = require('./api.js');

/**
 * 微信登录，获取openid
 */
function wxLogin() {
  return new Promise((resolve, reject) => {
    wx.login({
      success: (res) => {
        if (res.code) {
          // 这里需要将code发送到后端，后端通过code换取openid
          // 为了简化，我们直接使用code作为openid（实际项目中应该调用后端接口）
          // 实际应该调用：api.getOpenidByCode(res.code)
          resolve(res.code);
        } else {
          reject(new Error('登录失败：' + res.errMsg));
        }
      },
      fail: (err) => {
        reject(err);
      }
    });
  });
}

/**
 * 获取用户信息（需要用户授权）
 */
function getUserProfile() {
  return new Promise((resolve, reject) => {
    wx.getUserProfile({
      desc: '用于完善用户资料',
      success: (res) => {
        resolve(res.userInfo);
      },
      fail: (err) => {
        reject(err);
      }
    });
  });
}

/**
 * 自动登录并获取学生信息
 */
async function autoLogin() {
  try {
    // 1. 微信登录获取code
    const code = await wxLogin();
    
    // 2. 获取用户信息（可选）
    let userInfo = null;
    try {
      userInfo = await getUserProfile();
    } catch (err) {
      console.log('获取用户信息失败，使用默认名称');
    }

    // 3. 调用后端接口获取或创建学生
    const name = userInfo ? userInfo.nickName : '';
    const result = await api.getOrCreateStudentByOpenid(code, name);

    // 4. 保存学生ID
    if (result && result.student) {
      api.setStudentId(result.student.id);
      return {
        student: result.student,
        isNew: result.is_new
      };
    }

    throw new Error('登录失败');
  } catch (err) {
    console.error('自动登录失败:', err);
    throw err;
  }
}

/**
 * 检查是否已登录
 */
function isLoggedIn() {
  return api.getStudentId() !== null;
}

module.exports = {
  wxLogin,
  getUserProfile,
  autoLogin,
  isLoggedIn
};

