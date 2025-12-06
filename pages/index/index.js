// pages/index/index.js
const ebbinghaus = require('../../utils/ebbinghaus.js');
const api = require('../../utils/api.js');

Page({
  data: {
    todayWords: [],
    todayDate: ''
  },

  onLoad() {
    this.loadTodayWords();
  },

  /**
   * 跳转到设置页面
   */
  goToSettings() {
    wx.navigateTo({
      url: '/pages/settings/settings'
    });
  },

  onShow() {
    // 每次显示页面时重新加载，确保数据是最新的
    this.loadTodayWords();
  },

  /**
   * 加载今天需要背诵的单词
   */
  loadTodayWords() {
    const studentId = api.getStudentId();
    if (!studentId) {
      // 尝试自动登录
      const auth = require('../../utils/auth.js');
      auth.autoLogin()
        .then(() => {
          // 登录成功后重新加载
          this.loadTodayWords();
        })
        .catch(() => {
          wx.showModal({
            title: '提示',
            content: '请先登录',
            showCancel: false,
            success: () => {
              wx.navigateTo({
                url: '/pages/settings/settings'
              });
            }
          });
        });
      return;
    }

    // 显示加载中
    wx.showLoading({
      title: '加载中...',
      mask: true
    });

    const today = new Date();
    const todayStr = ebbinghaus.formatDate(today);

    // 从后端获取今日单词
    api.getTodayWords(todayStr)
      .then(res => {
        // 为每个单词添加拼写模式相关状态
        const wordsWithState = res.words.map(word => ({
          id: word.id,
          word: word.word,
          meaning: word.meaning,
          startDate: word.start_date,
          reviewCount: word.review_count || 0,
          lastReviewDate: word.last_review_date || '',
          studentId: word.student_id,
          spellingMode: false,
          userAnswer: '',
          showAnswer: false,
          isCorrect: false
        }));

        this.setData({
          todayWords: wordsWithState,
          todayDate: res.date || todayStr
        });
      })
      .catch(err => {
        console.error('加载单词失败:', err);
        wx.showToast({
          title: '加载失败',
          icon: 'none'
        });
        this.setData({
          todayWords: [],
          todayDate: todayStr
        });
      })
      .finally(() => {
        wx.hideLoading();
      });
  },

  /**
   * 开始拼写模式
   */
  startSpelling(e) {
    const index = e.currentTarget.dataset.index;
    const words = this.data.todayWords;
    words[index].spellingMode = true;
    words[index].userAnswer = '';
    words[index].showAnswer = false;
    words[index].isCorrect = false;
    
    this.setData({
      todayWords: words
    });
  },

  /**
   * 拼写输入
   */
  onSpellingInput(e) {
    const index = e.currentTarget.dataset.index;
    const value = e.detail.value;
    const words = this.data.todayWords;
    words[index].userAnswer = value;
    
    this.setData({
      todayWords: words
    });
  },

  /**
   * 检查答案
   */
  checkAnswer(e) {
    const index = e.currentTarget.dataset.index;
    const words = this.data.todayWords;
    const word = words[index];
    const userAnswer = (word.userAnswer || '').trim().toLowerCase();
    const correctAnswer = word.word.trim().toLowerCase();
    
    const isCorrect = userAnswer === correctAnswer;
    words[index].showAnswer = true;
    words[index].isCorrect = isCorrect;
    
    this.setData({
      todayWords: words
    });

    // 给出反馈
    if (isCorrect) {
      wx.showToast({
        title: '拼写正确！',
        icon: 'success',
        duration: 1500
      });
    } else {
      wx.showToast({
        title: '拼写错误',
        icon: 'none',
        duration: 1500
      });
    }
  },

  /**
   * 下一个单词
   */
  nextWord(e) {
    const index = e.currentTarget.dataset.index;
    const words = this.data.todayWords;
    
    // 重置当前单词状态
    words[index].spellingMode = false;
    words[index].userAnswer = '';
    words[index].showAnswer = false;
    words[index].isCorrect = false;
    
    // 如果有下一个单词，自动进入拼写模式
    if (index + 1 < words.length) {
      words[index + 1].spellingMode = true;
      words[index + 1].userAnswer = '';
      words[index + 1].showAnswer = false;
      words[index + 1].isCorrect = false;
    }
    
    this.setData({
      todayWords: words
    });
  },

  /**
   * 返回查看模式
   */
  backToView(e) {
    const index = e.currentTarget.dataset.index;
    const words = this.data.todayWords;
    words[index].spellingMode = false;
    words[index].userAnswer = '';
    words[index].showAnswer = false;
    words[index].isCorrect = false;
    
    this.setData({
      todayWords: words
    });
  },

  /**
   * 标记单词为已复习
   */
  markAsReviewed(e) {
    const wordId = e.currentTarget.dataset.id;
    
    wx.showLoading({
      title: '处理中...',
      mask: true
    });

    // 调用后端API标记为已复习
    api.markWordReviewed(wordId)
      .then(res => {
        wx.showToast({
          title: '已标记为复习',
          icon: 'success',
          duration: 1500
        });

        // 重新加载列表
        setTimeout(() => {
          this.loadTodayWords();
        }, 500);
      })
      .catch(err => {
        console.error('标记失败:', err);
        wx.showToast({
          title: '标记失败',
          icon: 'none'
        });
      })
      .finally(() => {
        wx.hideLoading();
      });
  },

  /**
   * 删除单词（暂时移除，如需删除功能可后续添加API）
   */
  deleteWord(e) {
    wx.showToast({
      title: '删除功能暂未实现',
      icon: 'none'
    });
  }
});

