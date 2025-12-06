// utils/ebbinghaus.js
// 艾宾浩斯遗忘曲线复习间隔（天数）
const REVIEW_INTERVALS = [1, 2, 4, 7, 15, 30];

/**
 * 计算两个日期之间的天数差
 * @param {Date} date1 
 * @param {Date} date2 
 * @returns {number} 天数差
 */
function getDaysBetween(date1, date2) {
  const oneDay = 24 * 60 * 60 * 1000;
  // 确保日期时间都设置为 00:00:00
  const d1 = new Date(date1);
  d1.setHours(0, 0, 0, 0);
  const d2 = new Date(date2);
  d2.setHours(0, 0, 0, 0);
  const diffTime = d2.getTime() - d1.getTime();
  return Math.floor(diffTime / oneDay);
}

/**
 * 格式化日期为 YYYY-MM-DD
 * @param {Date} date 
 * @returns {string}
 */
function formatDate(date) {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

/**
 * 判断单词是否需要在今天复习
 * @param {Object} word 单词对象
 * @param {Date} today 今天的日期
 * @returns {boolean}
 */
function shouldReviewToday(word, today) {
  const startDate = new Date(word.startDate);
  startDate.setHours(0, 0, 0, 0);
  const daysSinceStart = getDaysBetween(startDate, today);
  
  // 如果还没到开始日期，不需要复习
  if (daysSinceStart < 0) {
    return false;
  }
  
  // 如果是开始日期（第0天），需要第一次学习
  if (daysSinceStart === 0) {
    return true;
  }
  
  // 计算下一次应该复习的日期（累计天数）
  // reviewCount = 0 时，应该在开始日期 + REVIEW_INTERVALS[0] 天复习
  // reviewCount = 1 时，应该在开始日期 + REVIEW_INTERVALS[0] + REVIEW_INTERVALS[1] 天复习
  let cumulativeDays = 0;
  const reviewCount = word.reviewCount || 0;
  
  // 计算到当前复习次数为止的累计天数
  for (let i = 0; i <= reviewCount; i++) {
    if (i < REVIEW_INTERVALS.length) {
      cumulativeDays += REVIEW_INTERVALS[i];
    } else {
      // 如果超过预设的复习次数，使用最后一个间隔
      cumulativeDays += REVIEW_INTERVALS[REVIEW_INTERVALS.length - 1];
    }
  }
  
  // 如果今天正好是应该复习的日期，或者已经超过了应该复习的日期（需要补复习）
  if (daysSinceStart >= cumulativeDays) {
    return true;
  }
  
  return false;
}

/**
 * 获取今天需要复习的单词列表
 * @param {Array} words 所有单词列表
 * @param {Date} today 今天的日期（可选，默认为今天）
 * @returns {Array} 今天需要复习的单词列表
 */
function getTodayWords(words, today = new Date()) {
  // 重置今天的时间为 00:00:00，便于日期比较
  today.setHours(0, 0, 0, 0);
  
  return words.filter(word => {
    return shouldReviewToday(word, today);
  }).sort((a, b) => {
    // 按开始日期排序，早的在前
    return new Date(a.startDate) - new Date(b.startDate);
  });
}

/**
 * 标记单词为已复习
 * @param {Object} word 单词对象
 * @returns {Object} 更新后的单词对象
 */
function markAsReviewed(word) {
  return {
    ...word,
    reviewCount: (word.reviewCount || 0) + 1,
    lastReviewDate: formatDate(new Date())
  };
}

/**
 * 获取单词的下一次复习日期
 * @param {Object} word 单词对象
 * @returns {Date|null} 下一次复习日期
 */
function getNextReviewDate(word) {
  const startDate = new Date(word.startDate);
  let cumulativeDays = 0;
  
  for (let i = 0; i <= word.reviewCount; i++) {
    if (i < REVIEW_INTERVALS.length) {
      cumulativeDays += REVIEW_INTERVALS[i];
    } else {
      cumulativeDays += REVIEW_INTERVALS[REVIEW_INTERVALS.length - 1];
    }
  }
  
  const nextDate = new Date(startDate);
  nextDate.setDate(nextDate.getDate() + cumulativeDays);
  return nextDate;
}

module.exports = {
  REVIEW_INTERVALS,
  getDaysBetween,
  formatDate,
  shouldReviewToday,
  getTodayWords,
  markAsReviewed,
  getNextReviewDate
};

