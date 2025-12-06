// utils/api.js
// API 配置和调用工具

const API_BASE_URL = 'http://localhost:8000'; // 后端API地址，需要根据实际情况修改

/**
 * 获取学生ID（从本地存储）
 */
function getStudentId() {
  return wx.getStorageSync('studentId') || null;
}

/**
 * 设置学生ID
 */
function setStudentId(studentId) {
  wx.setStorageSync('studentId', studentId);
}

/**
 * 通用请求方法
 */
function request(options) {
  return new Promise((resolve, reject) => {
    wx.request({
      url: API_BASE_URL + options.url,
      method: options.method || 'GET',
      data: options.data || {},
      header: {
        'Content-Type': 'application/json',
        ...options.header
      },
      success: (res) => {
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve(res.data);
        } else {
          reject(new Error(res.data.message || '请求失败'));
        }
      },
      fail: (err) => {
        reject(err);
      }
    });
  });
}

/**
 * 获取今日需要背诵的单词
 */
function getTodayWords(date) {
  const studentId = getStudentId();
  if (!studentId) {
    return Promise.reject(new Error('请先设置学生ID'));
  }

  let url = `/api/v1/students/${studentId}/words/today`;
  if (date) {
    url += `?date=${date}`;
  }

  return request({
    url: url,
    method: 'GET'
  });
}

/**
 * 标记单词为已复习
 */
function markWordReviewed(wordId) {
  const studentId = getStudentId();
  if (!studentId) {
    return Promise.reject(new Error('请先设置学生ID'));
  }

  return request({
    url: `/api/v1/students/${studentId}/words/${wordId}/review`,
    method: 'POST'
  });
}

/**
 * 批量添加单词
 */
function batchAddWords(words) {
  const studentId = getStudentId();
  if (!studentId) {
    return Promise.reject(new Error('请先设置学生ID'));
  }

  return request({
    url: `/api/v1/students/${studentId}/words/batch`,
    method: 'POST',
    data: {
      words: words
    }
  });
}

/**
 * 获取学生信息
 */
function getStudent(studentId) {
  return request({
    url: `/api/v1/students/${studentId}`,
    method: 'GET'
  });
}

/**
 * 创建学生
 */
function createStudent(name, studentNo) {
  return request({
    url: '/api/v1/students',
    method: 'POST',
    data: {
      name: name,
      student_no: studentNo
    }
  });
}

/**
 * 通过openid获取或创建学生
 */
function getOrCreateStudentByOpenid(openid, name) {
  return request({
    url: '/api/v1/students/by-openid',
    method: 'POST',
    data: {
      openid: openid,
      name: name || ''
    }
  });
}

module.exports = {
  getStudentId,
  setStudentId,
  getTodayWords,
  markWordReviewed,
  batchAddWords,
  getStudent,
  createStudent,
  getOrCreateStudentByOpenid,
  API_BASE_URL
};

