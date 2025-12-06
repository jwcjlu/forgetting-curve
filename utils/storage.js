// utils/storage.js
// 本地存储工具

const STORAGE_KEY = 'words';

/**
 * 获取所有单词
 * @returns {Array} 单词列表
 */
function getWords() {
  try {
    return wx.getStorageSync(STORAGE_KEY) || [];
  } catch (e) {
    console.error('获取单词列表失败:', e);
    return [];
  }
}

/**
 * 保存单词列表
 * @param {Array} words 单词列表
 */
function saveWords(words) {
  try {
    wx.setStorageSync(STORAGE_KEY, words);
    return true;
  } catch (e) {
    console.error('保存单词列表失败:', e);
    wx.showToast({
      title: '保存失败',
      icon: 'none'
    });
    return false;
  }
}

/**
 * 添加单词
 * @param {Object} word 单词对象 {word: string, meaning: string, startDate: string}
 * @returns {boolean} 是否成功
 */
function addWord(word) {
  const words = getWords();
  const newWord = {
    id: Date.now() + Math.random(), // 生成唯一ID
    word: word.word,
    meaning: word.meaning,
    startDate: word.startDate,
    reviewCount: 0,
    createdAt: new Date().toISOString()
  };
  
  words.push(newWord);
  return saveWords(words);
}

/**
 * 更新单词
 * @param {string} id 单词ID
 * @param {Object} updates 要更新的字段
 * @returns {boolean} 是否成功
 */
function updateWord(id, updates) {
  const words = getWords();
  const index = words.findIndex(w => w.id === id);
  
  if (index === -1) {
    return false;
  }
  
  words[index] = {
    ...words[index],
    ...updates
  };
  
  return saveWords(words);
}

/**
 * 删除单词
 * @param {string} id 单词ID
 * @returns {boolean} 是否成功
 */
function deleteWord(id) {
  const words = getWords();
  const filteredWords = words.filter(w => w.id !== id);
  return saveWords(filteredWords);
}

/**
 * 根据ID获取单词
 * @param {string} id 单词ID
 * @returns {Object|null} 单词对象
 */
function getWordById(id) {
  const words = getWords();
  return words.find(w => w.id === id) || null;
}

module.exports = {
  getWords,
  saveWords,
  addWord,
  updateWord,
  deleteWord,
  getWordById
};

