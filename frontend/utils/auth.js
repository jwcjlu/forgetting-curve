// utils/auth.js
// 微信登录和认证工具

const api = require('./api.js');

/**
 * 微信登录，获取code
 */
function wxLogin() {
  return new Promise((resolve, reject) => {
    wx.login({
      success: (res) => {
        if (res.code) {
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
    console.log('[AutoLogin] 调用登录接口，code 长度:', code ? code.length : 0);
    
    const result = await api.getOrCreateStudentByOpenid(null, code, name);
    console.log('[AutoLogin] 登录接口返回:', result ? JSON.stringify(result).substring(0, 200) : 'null');

    // 4. 保存学生ID
    if (result && result.student && result.student.id) {
      api.setStudentId(result.student.id);
      console.log('[AutoLogin] 登录成功，学生ID:', result.student.id, 'openid:', result.student.openid);
      return {
        student: result.student,
        isNew: result.is_new || result.isNew || false
      };
    }

    // 检查是否有错误信息
    if (result && result.ret && result.ret.code !== 0) {
      throw new Error(result.ret.message || '登录失败');
    }

    throw new Error('登录失败：未返回学生信息');
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
