// utils/auth.js
// 微信登录和认证工具

const api = require('./api.js');

// 登录状态管理
let isLoggingIn = false; // 是否正在登录
let loginPromise = null;  // 登录 Promise，用于防止并发登录

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
 * 使用锁机制防止并发登录，避免 code 被重复使用
 */
async function autoLogin() {
  // 如果已有学生ID，直接返回
  if (api.getStudentId()) {
    console.log('[AutoLogin] 已有学生ID，跳过登录');
    return {
      student: { id: api.getStudentId() },
      isNew: false
    };
  }

  // 如果正在登录，等待正在进行的登录完成
  if (isLoggingIn && loginPromise) {
    console.log('[AutoLogin] 登录正在进行中，等待完成...');
    return loginPromise;
  }

  // 开始新的登录流程
  isLoggingIn = true;
  loginPromise = (async () => {
    try {
      // 1. 微信登录获取code（每次调用都获取新的 code）
      console.log('[AutoLogin] 开始获取微信 code...');
      const codeStartTime = Date.now();
      const code = await wxLogin();
      const codeGetTime = Date.now() - codeStartTime;
      console.log('[AutoLogin] 获取到 code，长度:', code ? code.length : 0, '耗时:', codeGetTime + 'ms', 'code 前10位:', code ? code.substring(0, 10) : 'null');
      
      if (!code || code.trim() === '') {
        throw new Error('获取微信 code 失败');
      }
      
      // 2. 立即调用后端接口，不要延迟（code 有时效性，且只能使用一次）
      // 注意：code 只能使用一次，且有时效性（约5分钟），必须立即使用
      console.log('[AutoLogin] 立即调用后端接口，传递 code（不传递 openid）...');
      const apiStartTime = Date.now();
      
      // 先调用后端接口，获取用户信息可以并行或稍后进行
      // 3. 立即调用后端接口获取或创建学生（传递 code，后端会转换为 openid）
      // 必须在获取 code 后立即调用，避免过期
      // 注意：第一个参数传 null，确保使用 code 而不是 openid
      let result;
      let retryCount = 0;
      const maxRetries = 2; // 最多重试2次
      
      while (retryCount <= maxRetries) {
        try {
          // 如果重试，需要重新获取 code（因为 code 只能使用一次）
          let codeToUse = code;
          if (retryCount > 0) {
            console.log('[AutoLogin] 重试登录，重新获取 code... (第 ' + retryCount + ' 次)');
            codeToUse = await wxLogin();
            console.log('[AutoLogin] 重新获取到 code，长度:', codeToUse ? codeToUse.length : 0);
            if (!codeToUse || codeToUse.trim() === '') {
              throw new Error('重新获取 code 失败');
            }
          }
          
          result = await api.getOrCreateStudentByOpenid(null, codeToUse, '');
          const apiTime = Date.now() - apiStartTime;
          console.log('[AutoLogin] 后端接口调用完成，耗时:', apiTime + 'ms');
          break; // 成功，退出循环
        } catch (err) {
          const apiTime = Date.now() - apiStartTime;
          console.error('[AutoLogin] 后端接口调用失败，耗时:', apiTime + 'ms', '重试次数:', retryCount, err);
          
          // 检查是否是 code 相关错误
          const errorMsg = err.message || '';
          const isCodeError = errorMsg.indexOf('code 无效') !== -1 || 
                             errorMsg.indexOf('invalid code') !== -1 ||
                             errorMsg.indexOf('code 已被使用') !== -1;
          
          // 如果是 code 错误且还有重试次数，则重试
          if (isCodeError && retryCount < maxRetries) {
            retryCount++;
            // 等待一小段时间再重试
            await new Promise(resolve => setTimeout(resolve, 500));
            continue;
          }
          
          // 其他错误或重试次数用完，抛出错误
          throw err;
        }
      }
      
      // 如果登录成功，尝试更新用户信息（可选）
      if (result && result.student) {
        try {
          const userInfo = await getUserProfile();
          if (userInfo && userInfo.nickName && userInfo.nickName !== result.student.name) {
            console.log('[AutoLogin] 获取到用户信息，更新名称:', userInfo.nickName);
            // 可以在这里调用更新接口，但为了简化，暂时不更新
          }
        } catch (err) {
          console.log('[AutoLogin] 获取用户信息失败，不影响登录:', err);
        }
      }

      // 4. 保存学生ID
      if (result && result.student) {
        api.setStudentId(result.student.id);
        console.log('[AutoLogin] 登录成功，学生ID:', result.student.id);
        return {
          student: result.student,
          isNew: result.is_new || false
        };
      }

      throw new Error('登录失败：未返回学生信息');
    } catch (err) {
      console.error('[AutoLogin] 自动登录失败:', err);
      
      // 提供更友好的错误信息
      let errorMessage = err.message || '登录失败';
      // 使用 indexOf 替代 includes，兼容小程序环境
      if (errorMessage.indexOf('invalid code') !== -1 || errorMessage.indexOf('code 无效') !== -1) {
        errorMessage = '登录码已过期或无效，请重新打开小程序重试';
      } else if (errorMessage.indexOf('code 已被使用') !== -1) {
        errorMessage = '登录码已被使用，请重新打开小程序';
      } else if (errorMessage.indexOf('AppID') !== -1 || errorMessage.indexOf('AppSecret') !== -1) {
        errorMessage = '服务器配置错误，请联系管理员';
      }
      
      throw new Error(errorMessage);
    } finally {
      // 重置登录状态
      isLoggingIn = false;
      loginPromise = null;
    }
  })();

  return loginPromise;
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
