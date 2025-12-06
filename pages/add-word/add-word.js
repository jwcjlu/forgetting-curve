// pages/add-word/add-word.js
const storage = require('../../utils/storage.js');
const ebbinghaus = require('../../utils/ebbinghaus.js');

Page({
  data: {
    word: '',
    meaning: '',
    startDate: '',
    minDate: '',
    allWords: []
  },

  onLoad() {
    // 设置最小日期为今天
    const today = new Date();
    const minDate = ebbinghaus.formatDate(today);
    const defaultStartDate = minDate;
    
    this.setData({
      minDate: minDate,
      startDate: defaultStartDate
    });
    
    this.loadAllWords();
  },

  onShow() {
    // 每次显示页面时重新加载单词列表
    this.loadAllWords();
  },

  /**
   * 加载所有单词
   */
  loadAllWords() {
    const words = storage.getWords();
    // 按创建时间倒序排列
    words.sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt));
    this.setData({
      allWords: words
    });
  },

  /**
   * 单词输入
   */
  onWordInput(e) {
    this.setData({
      word: e.detail.value
    });
  },

  /**
   * 释义输入
   */
  onMeaningInput(e) {
    this.setData({
      meaning: e.detail.value
    });
  },

  /**
   * 日期选择
   */
  onDateChange(e) {
    this.setData({
      startDate: e.detail.value
    });
  },

  /**
   * 提交单词
   */
  submitWord() {
    let { word, meaning, startDate } = this.data;

    // 去除首尾空格
    word = word ? word.trim() : '';
    meaning = meaning ? meaning.trim() : '';

    // 验证输入
    if (!word) {
      wx.showToast({
        title: '请输入单词',
        icon: 'none'
      });
      return;
    }

    if (!meaning) {
      wx.showToast({
        title: '请输入释义',
        icon: 'none'
      });
      return;
    }

    if (!startDate) {
      wx.showToast({
        title: '请选择开始日期',
        icon: 'none'
      });
      return;
    }

    // 检查是否已存在相同单词
    const words = storage.getWords();
    const exists = words.some(w => w.word.toLowerCase() === word.toLowerCase());
    
    if (exists) {
      wx.showModal({
        title: '提示',
        content: '该单词已存在，是否继续添加？',
        success: (res) => {
          if (res.confirm) {
            this.addWord(word, meaning, startDate);
          }
        }
      });
    } else {
      this.addWord(word, meaning, startDate);
    }
  },

  /**
   * 添加单词
   */
  addWord(word, meaning, startDate) {
    const success = storage.addWord({
      word: word,
      meaning: meaning,
      startDate: startDate
    });

    if (success) {
      wx.showToast({
        title: '添加成功',
        icon: 'success'
      });

      // 清空表单
      this.setData({
        word: '',
        meaning: ''
      });

      // 重新加载列表
      this.loadAllWords();
    }
  },

  /**
   * 删除单词
   */
  deleteWord(e) {
    const id = e.currentTarget.dataset.id;
    
    wx.showModal({
      title: '确认删除',
      content: '确定要删除这个单词吗？',
      success: (res) => {
        if (res.confirm) {
          storage.deleteWord(id);
          wx.showToast({
            title: '删除成功',
            icon: 'success'
          });
          this.loadAllWords();
        }
      }
    });
  }
});

