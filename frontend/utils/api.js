// utils/api.js
// API 配置和调用工具（基于 proto 协议）

// 后端API地址，需要根据实际情况修改
const API_BASE_URL = 'https://forgetting-curve.cpxdmz.top';

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
    const startTime = Date.now();
    console.log('[API Request]', options.method || 'GET', options.url, options.data);
    
    wx.request({
      url: API_BASE_URL + options.url,
      method: options.method || 'GET',
      data: options.data || {},
      header: {
        'Content-Type': 'application/json',
        ...options.header
      },
      timeout: 10000, // 10秒超时
      success: (res) => {
        const duration = Date.now() - startTime;
        console.log('[API Response]', res.statusCode, '耗时:', duration + 'ms', res.data);
        
        if (res.statusCode >= 200 && res.statusCode < 300) {
          // 检查业务错误码
          if (res.data && res.data.ret) {
            if (res.data.ret.code === 0) {
              resolve(res.data);
            } else {
              const errorMsg = res.data.ret.message || '请求失败';
              console.error('[API Error]', errorMsg, res.data.ret);
              reject(new Error(errorMsg));
            }
          } else {
            resolve(res.data);
          }
        } else {
          const errorMsg = res.data?.ret?.message || `HTTP ${res.statusCode} 错误`;
          console.error('[API HTTP Error]', res.statusCode, errorMsg);
          reject(new Error(errorMsg));
        }
      },
      fail: (err) => {
        const duration = Date.now() - startTime;
        console.error('[API Request Failed]', err, '耗时:', duration + 'ms');
        
        let errorMsg = '网络请求失败';
        if (err.errMsg) {
          // 使用 indexOf 替代 includes，兼容小程序环境
          if (err.errMsg.indexOf('timeout') !== -1) {
            errorMsg = '请求超时，请检查网络连接';
          } else if (err.errMsg.indexOf('fail') !== -1) {
            errorMsg = '无法连接到服务器，请检查API地址配置';
          } else {
            errorMsg = err.errMsg;
          }
        }
        reject(new Error(errorMsg));
      }
    });
  });
}

/**
 * 通过openid或code获取或创建学生（微信登录）
 * 对应: GetOrCreateStudentByOpenid
 * @param {string|null} openid - openid（如果直接提供）
 * @param {string|null} code - 微信登录code（优先使用，后端会转换为openid）
 * @param {string} name - 用户昵称（可选）
 */
function getOrCreateStudentByOpenid(openid, code, name) {
  const data = {
    name: name || ''
  };
  
  // 优先使用 code，如果没有 code 则使用 openid
  if (code && code.trim() !== '') {
    data.code = code;
    console.log('[API] 使用 code 登录，code 长度:', code.length);
  } else if (openid && openid.trim() !== '') {
    data.openid = openid;
    console.log('[API] 使用 openid 登录');
  } else {
    console.error('[API] 错误：openid 和 code 都为空');
    return Promise.reject(new Error('openid 或 code 必须提供一个'));
  }
  
  console.log('[API] 请求数据:', JSON.stringify(data).replace(/code":"[^"]+/, 'code":"***'));
  
  return request({
    url: '/api/v1/students/by-openid',
    method: 'POST',
    data: data
  });
}

/**
 * 获取学生信息
 * 对应: GetStudent
 */
function getStudent(studentId) {
  return request({
    url: `/api/v1/students/${studentId}`,
    method: 'GET'
  });
}

/**
 * 创建学生
 * 对应: CreateStudent
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
 * 获取今日需要背诵的单词（根据艾宾浩斯曲线）
 * 对应: GetTodayWords
 * @param {string} date - 日期（可选，格式：YYYY-MM-DD）
 * @deprecated 已废弃，请使用 getPlanTodayWords(planId, date)
 */
function getTodayWords(date) {
  // 尝试从本地存储获取激活的计划ID
  const activePlanId = wx.getStorageSync('activePlanId');
  if (!activePlanId) {
    return Promise.reject(new Error('请先选择复习计划'));
  }
  return getPlanTodayWords(activePlanId, date);
}

/**
 * 标记单词为已复习
 * 对应: MarkWordReviewed
 * @param {number} planId - 计划ID
 * @param {number} wordId - 单词ID
 */
function markWordReviewed(planId, wordId) {
  const studentId = getStudentId();
  if (!studentId) {
    return Promise.reject(new Error('请先登录'));
  }

  if (!planId) {
    return Promise.reject(new Error('请先选择复习计划'));
  }

  return request({
    url: `/api/v1/students/${studentId}/plans/${planId}/words/${wordId}/review`,
    method: 'POST'
  });
}

/**
 * 批量添加单词到计划
 * 对应: BatchAddWords
 * @param {number} planId - 计划ID
 * @param {Array} words - 单词列表
 */
function batchAddWords(planId, words) {
  const studentId = getStudentId();
  if (!studentId) {
    return Promise.reject(new Error('请先登录'));
  }

  if (!planId) {
    return Promise.reject(new Error('请先选择复习计划'));
  }

  return request({
    url: `/api/v1/students/${studentId}/plans/${planId}/words/batch`,
    method: 'POST',
    data: {
      words: words
    }
  });
}

/**
 * 获取计划的单词列表（分页）
 * 对应: GetPlanWords
 * @param {number} planId - 计划ID
 * @param {number} page - 页码，默认1
 * @param {number} pageSize - 每页数量，默认20
 */
function getPlanWords(planId, page = 1, pageSize = 20) {
  const studentId = getStudentId();
  if (!studentId) {
    return Promise.reject(new Error('请先登录'));
  }

  if (!planId) {
    return Promise.reject(new Error('请先选择复习计划'));
  }

  return request({
    url: `/api/v1/students/${studentId}/plans/${planId}/words?page=${page}&page_size=${pageSize}`,
    method: 'GET'
  });
}

/**
 * @deprecated 已废弃，请使用 getPlanWords(planId, page, pageSize)
 * 获取学生的单词列表（分页）- 保留用于兼容，实际会调用 getPlanWords
 */
function getStudentWords(page = 1, pageSize = 20) {
  // 尝试从本地存储获取激活的计划ID
  const activePlanId = wx.getStorageSync('activePlanId');
  if (!activePlanId) {
    return Promise.reject(new Error('请先选择复习计划'));
  }
  return getPlanWords(activePlanId, page, pageSize);
}

/**
 * 添加混淆词
 * 对应: AddConfusedWord
 * @param {number} wordId - 主单词ID
 * @param {number} confusedWordId - 混淆词ID
 */
function addConfusedWord(wordId, confusedWordId) {
  const studentId = getStudentId();
  if (!studentId) {
    return Promise.reject(new Error('请先登录'));
  }

  return request({
    url: `/api/v1/students/${studentId}/words/${wordId}/confused`,
    method: 'POST',
    data: {
      confused_word_id: confusedWordId
    }
  });
}

/**
 * 获取单词的混淆词列表
 * 对应: GetConfusedWords
 * @param {number} planId - 计划ID
 * @param {number} wordId - 单词ID
 */
function getConfusedWords(planId, wordId) {
  const studentId = getStudentId();
  if (!studentId) {
    return Promise.reject(new Error('请先登录'));
  }

  if (!planId) {
    return Promise.reject(new Error('请先选择复习计划'));
  }

  return request({
    url: `/api/v1/students/${studentId}/plans/${planId}/words/${wordId}/confused`,
    method: 'GET'
  });
}

/**
 * 删除混淆词
 * 对应: RemoveConfusedWord
 * @param {number} planId - 计划ID
 * @param {number} wordId - 主单词ID
 * @param {number} confusedWordId - 混淆词ID
 */
function removeConfusedWord(planId, wordId, confusedWordId) {
  const studentId = getStudentId();
  if (!studentId) {
    return Promise.reject(new Error('请先登录'));
  }

  if (!planId) {
    return Promise.reject(new Error('请先选择复习计划'));
  }

  return request({
    url: `/api/v1/students/${studentId}/plans/${planId}/words/${wordId}/confused/${confusedWordId}`,
    method: 'DELETE'
  });
}

/**
 * 搜索单词（用于添加混淆词）
 * 对应: SearchWords
 * @param {number} planId - 计划ID
 * @param {string} keyword - 搜索关键词（支持正则表达式）
 * @param {number} limit - 返回数量限制，默认20
 */
function searchWords(planId, keyword, limit = 20) {
  const studentId = getStudentId();
  if (!studentId) {
    return Promise.reject(new Error('请先登录'));
  }

  if (!planId) {
    return Promise.reject(new Error('请先选择复习计划'));
  }

  let url = `/api/v1/students/${studentId}/plans/${planId}/words/search?keyword=${encodeURIComponent(keyword)}`;
  if (limit) {
    url += `&limit=${limit}`;
  }

  return request({
    url: url,
    method: 'GET'
  });
}

/**
 * 标记单词为未记住
 * 对应: MarkWordForgotten
 * @param {number} planId - 计划ID
 * @param {number} wordId - 单词ID
 */
function markWordForgotten(planId, wordId) {
  const studentId = getStudentId();
  if (!studentId) {
    return Promise.reject(new Error('请先登录'));
  }

  if (!planId) {
    return Promise.reject(new Error('请先选择复习计划'));
  }

  return request({
    url: `/api/v1/students/${studentId}/plans/${planId}/words/${wordId}/forget`,
    method: 'POST'
  });
}

/**
 * 更新单词复习数据（思考时间、难度等）
 * 对应: UpdateWordReviewData
 * @param {number} planId - 计划ID
 * @param {number} wordId - 单词ID
 * @param {number} thinkTime - 思考时间（秒）
 * @param {number} difficulty - 学习难度 (0-10)
 * @param {boolean} isRemembered - 是否记住
 */
function updateWordReviewData(planId, wordId, thinkTime, difficulty, isRemembered) {
  const studentId = getStudentId();
  if (!studentId) {
    return Promise.reject(new Error('请先登录'));
  }

  if (!planId) {
    return Promise.reject(new Error('请先选择复习计划'));
  }

  return request({
    url: `/api/v1/students/${studentId}/plans/${planId}/words/${wordId}/review-data`,
    method: 'POST',
    data: {
      think_time: thinkTime,
      difficulty: difficulty,
      is_remembered: isRemembered
    }
  });
}

/**
 * OCR识别图片中的单词
 * 对应: RecognizeWordsFromImage
 * @param {string} imageBase64 - base64编码的图片数据
 * @param {string} startDate - 开始日期，格式：YYYY-MM-DD
 */
function recognizeWordsFromImage(imageBase64, startDate) {
  const studentId = getStudentId();
  if (!studentId) {
    return Promise.reject(new Error('请先登录'));
  }

  return request({
    url: `/api/v1/students/${studentId}/words/recognize`,
    method: 'POST',
    data: {
      image_base64: imageBase64,
      start_date: startDate || ''
    }
  });
}

/**
 * 生成复习题目
 * 对应: GenerateReviewQuestions
 * @param {number} planId - 计划ID
 * @param {number} wordId - 单词ID
 * @param {string} grade - 年级（可选，可以从计划中获取）
 */
function generateReviewQuestions(planId, wordId, grade) {
  const studentId = getStudentId();
  if (!studentId) {
    return Promise.reject(new Error('请先登录'));
  }

  if (!planId) {
    return Promise.reject(new Error('请先选择复习计划'));
  }

  const data = {};
  if (grade && grade.trim() !== '') {
    data.grade = grade.trim();
  }

  return request({
    url: `/api/v1/students/${studentId}/plans/${planId}/words/${wordId}/review-questions`,
    method: 'POST',
    data: data
  });
}

/**
 * 创建复习计划
 * @param {string} name - 计划名称
 * @param {string} grade - 年级
 */
function createPlan(name, grade) {
  const studentId = getStudentId();
  if (!studentId) {
    return Promise.reject(new Error('请先登录'));
  }

  return request({
    url: `/api/v1/students/${studentId}/plans`,
    method: 'POST',
    data: {
      name: name,
      grade: grade || ''
    }
  });
}

/**
 * 获取学生的所有计划
 */
function getPlans() {
  const studentId = getStudentId();
  if (!studentId) {
    return Promise.reject(new Error('请先登录'));
  }

  return request({
    url: `/api/v1/students/${studentId}/plans`,
    method: 'GET'
  });
}

/**
 * 选择计划（设置为当前激活的计划）
 * @param {number} planId - 计划ID
 */
function selectPlan(planId) {
  const studentId = getStudentId();
  if (!studentId) {
    return Promise.reject(new Error('请先登录'));
  }

  return request({
    url: `/api/v1/students/${studentId}/plans/${planId}/select`,
    method: 'POST'
  });
}

/**
 * 添加单词到计划
 * @param {number} planId - 计划ID
 * @param {Array<number>} wordIds - 单词ID列表
 */
function addWordsToPlan(planId, wordIds) {
  const studentId = getStudentId();
  if (!studentId) {
    return Promise.reject(new Error('请先登录'));
  }

  return request({
    url: `/api/v1/students/${studentId}/plans/${planId}/words`,
    method: 'POST',
    data: {
      word_ids: wordIds
    }
  });
}

/**
 * 从计划中移除单词
 * @param {number} planId - 计划ID
 * @param {number} wordId - 单词ID
 */
function removeWordsFromPlan(planId, wordId) {
  const studentId = getStudentId();
  if (!studentId) {
    return Promise.reject(new Error('请先登录'));
  }

  return request({
    url: `/api/v1/students/${studentId}/plans/${planId}/words/${wordId}`,
    method: 'DELETE'
  });
}


/**
 * 更新计划
 * @param {number} planId - 计划ID
 * @param {string} name - 计划名称
 * @param {string} grade - 年级
 */
function updatePlan(planId, name, grade) {
  const studentId = getStudentId();
  if (!studentId) {
    return Promise.reject(new Error('请先登录'));
  }

  return request({
    url: `/api/v1/students/${studentId}/plans/${planId}`,
    method: 'PUT',
    data: {
      name: name,
      grade: grade || ''
    }
  });
}

/**
 * 删除计划
 * @param {number} planId - 计划ID
 */
function deletePlan(planId) {
  const studentId = getStudentId();
  if (!studentId) {
    return Promise.reject(new Error('请先登录'));
  }

  return request({
    url: `/api/v1/students/${studentId}/plans/${planId}`,
    method: 'DELETE'
  });
}

/**
 * 获取计划的今日单词（根据艾宾浩斯曲线）
 * @param {number} planId - 计划ID
 * @param {string} date - 日期（可选，格式：YYYY-MM-DD）
 */
function getPlanTodayWords(planId, date) {
  const studentId = getStudentId();
  if (!studentId) {
    return Promise.reject(new Error('请先登录'));
  }

  const url = `/api/v1/students/${studentId}/plans/${planId}/words/today${date ? '?date=' + date : ''}`;
  return request({
    url: url,
    method: 'GET'
  });
}

module.exports = {
  getStudentId,
  setStudentId,
  getOrCreateStudentByOpenid,
  getStudent,
  createStudent,
  getTodayWords,
  markWordReviewed,
  batchAddWords,
  getStudentWords,
  addConfusedWord,
  getConfusedWords,
  removeConfusedWord,
  searchWords,
  markWordForgotten,
  updateWordReviewData,
  recognizeWordsFromImage,
  generateReviewQuestions,
  createPlan,
  getPlans,
  selectPlan,
  updatePlan,
  addWordsToPlan,
  removeWordsFromPlan,
  getPlanWords,
  deletePlan,
  getPlanTodayWords,
  API_BASE_URL
};
