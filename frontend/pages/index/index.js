// pages/index/index.js
const api = require('../../utils/api.js');
const auth = require('../../utils/auth.js');
const ebbinghaus = require('../../utils/ebbinghaus.js');

Page({
  data: {
    todayWords: [],
    todayDate: '',
    loading: false,
    activePlanId: null,
    activePlanName: '',
    hasActivePlan: false,
    // 混淆词相关
    showConfusedWordModal: false,
    currentWordIndex: -1,
    currentWordId: null,
    confusedWordKeyword: '',
    confusedWordSearchResults: [],
    // 音频播放相关
    audioContext: null,
    playingWordId: null
  },

  onLoad() {
    // 先检查计划，再加载单词
    this.checkActivePlan().then(() => {
      this.loadTodayWords();
    });
    // 创建音频上下文
    try {
      this.data.audioContext = wx.createInnerAudioContext();
      if (!this.data.audioContext) {
        console.error('创建音频上下文失败: 返回 null');
        return;
      }
      this.data.audioContext.onEnded(() => {
        this.setData({
          playingWordId: null
        });
      });
      this.data.audioContext.onError((err) => {
        console.error('音频播放失败:', err);
        wx.showToast({
          title: '播放失败',
          icon: 'none',
          duration: 1500
        });
        this.setData({
          playingWordId: null
        });
      });
    } catch (err) {
      console.error('初始化音频上下文失败:', err);
      this.data.audioContext = null;
    }
  },

  onUnload() {
    // 页面卸载时销毁音频上下文
    if (this.data.audioContext) {
      try {
        this.data.audioContext.stop();
        this.data.audioContext.destroy();
      } catch (err) {
        console.error('销毁音频上下文失败:', err);
      }
      this.data.audioContext = null;
    }
  },

  onShow() {
    // 每次显示页面时重新加载，确保数据是最新的
    // 先检查计划，再加载单词
    this.checkActivePlan().then(() => {
      this.loadTodayWords();
    });
  },

  /**
   * 跳转到设置页面
   */
  goToSettings() {
    wx.navigateTo({
      url: '/pages/settings/settings'
    });
  },

  /**
   * 跳转到计划管理页面
   */
  goToPlans() {
    wx.navigateTo({
      url: '/pages/plan/plan'
    });
  },

  /**
   * 检查并加载激活的计划信息
   */
  checkActivePlan() {
    const activePlanId = wx.getStorageSync('activePlanId');
    
    if (!activePlanId) {
      this.setData({
        hasActivePlan: false,
        activePlanId: null,
        activePlanName: ''
      });
      return Promise.resolve(false);
    }

    // 获取计划列表，找到激活的计划
    return api.getPlans()
      .then(res => {
        const activePlan = (res.plans || []).find(p => p.id === activePlanId || p.is_active);
        if (activePlan) {
          this.setData({
            hasActivePlan: true,
            activePlanId: activePlan.id,
            activePlanName: activePlan.name
          });
          return true;
        } else {
          // 计划不存在，清除本地存储
          wx.removeStorageSync('activePlanId');
          this.setData({
            hasActivePlan: false,
            activePlanId: null,
            activePlanName: ''
          });
          return false;
        }
      })
      .catch(err => {
        console.error('获取计划列表失败:', err);
        this.setData({
          hasActivePlan: false,
          activePlanId: null,
          activePlanName: ''
        });
        return false;
      });
  },

  /**
   * 加载今天需要背诵的单词
   */
  loadTodayWords() {
    // 先定义 todayStr，确保在所有地方都能访问
    const today = new Date();
    const todayStr = ebbinghaus.formatDate(today);
    
    const studentId = api.getStudentId();
    if (!studentId) {
      // 尝试自动登录
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

    // 先检查是否有激活的计划
    this.checkActivePlan()
      .then(hasPlan => {
        if (!hasPlan) {
          // 没有激活的计划，显示提示
          this.setData({ 
            loading: false,
            todayWords: []
          });
          return;
        }

        // 有激活的计划，加载单词
        this.setData({ loading: true });

        const activePlanId = this.data.activePlanId;
        
        // 使用计划的今日单词API
        return api.getPlanTodayWords(activePlanId, todayStr);
      })
      .then(res => {
        if (!res) {
          return; // 没有激活的计划，已处理
        }
        // 为每个单词添加拼写模式相关状态
        const wordsWithState = (res.words || []).map(word => {
          // 检查当天是否已复习
          const isReviewedToday = word.last_review_date === todayStr;
          // 兼容两种字段名：audio_urls（下划线）和 audioUrls（驼峰）
          const audioUrls = word.audio_urls || word.audioUrls || [];
          
          // 调试日志：检查音频字段
          if (audioUrls.length > 0) {
            console.log(`单词 "${word.word}" 有 ${audioUrls.length} 个音频:`, audioUrls);
          }
          
          // 生成遮盖的单词和释义
          // 单词：保留第一个和最后一个字母，中间用下划线遮盖
          const wordText = word.word || '';
          let maskedWord = '';
          if (wordText.length <= 2) {
            // 如果单词长度小于等于2，全部显示
            maskedWord = wordText;
          } else {
            // 保留第一个和最后一个字母，中间用下划线
            const firstChar = wordText[0];
            const lastChar = wordText[wordText.length - 1];
            const middleChars = '_'.repeat(wordText.length - 2);
            maskedWord = firstChar + middleChars + lastChar;
          }
          // 释义：全部遮盖
          const meaningText = word.meaning || '';
          const maskedMeaning = meaningText.replace(/./g, '_');
          
          return {
            id: word.id,
            word: word.word,
            meaning: word.meaning,
            start_date: word.start_date,
            review_count: word.review_count || 0,
            last_review_date: word.last_review_date || '',
            student_id: word.student_id,
            confused_words: word.confused_words || [], // 混淆词列表（如果后端已返回）
            audio_urls: audioUrls, // 音频 URL 列表（统一使用下划线命名）
            spellingMode: false,
            reviewMode: false,
            questions: [],
            currentQuestionIndex: 0,
            userAnswers: [],
            showAnswers: [],
            canSubmit: false,
            userAnswer: '',
            showAnswer: false,
            isCorrect: false,
            // 遮盖相关
            showWord: false, // 是否显示单词（默认遮盖）
            maskedWord: maskedWord, // 遮盖的单词
            maskedMeaning: maskedMeaning, // 遮盖的释义
            // 答题状态
            hasAnswered: false, // 是否已提交答案
            allCorrect: false, // 是否全部答对
            // 复习状态
            isReviewedToday: isReviewedToday // 当天是否已复习
          };
        });

        // 如果后端没有返回混淆词列表，则加载每个单词的混淆词列表
        // 否则直接使用后端返回的数据
        const needLoadConfusedWords = wordsWithState.some(word => !word.confused_words || word.confused_words.length === 0);
        if (needLoadConfusedWords) {
          this.loadConfusedWordsForAll(wordsWithState);
        } else {
          this.setData({
            todayWords: wordsWithState,
            todayDate: res.date || todayStr,
            loading: false
          });
        }
      })
      .catch(err => {
        console.error('加载单词失败:', err);
        wx.showToast({
          title: err.message || '加载失败',
          icon: 'none',
          duration: 2000
        });
        this.setData({
          todayWords: [],
          todayDate: todayStr,
          loading: false
        });
      });
  },

  /**
   * 开始复习模式
   */
  startReview(e) {
    const index = e.currentTarget.dataset.index;
    const word = this.data.todayWords[index];
    
    // 获取年级信息
    const grade = wx.getStorageSync('grade') || '';
    
    wx.showLoading({
      title: '生成题目中...',
      mask: true
    });

    // 调用后端API生成题目
    api.generateReviewQuestions(this.data.activePlanId, word.id, grade)
      .then(res => {
        wx.hideLoading();
        
        console.log('生成题目API响应:', res);
        console.log('题目数据:', res.questions);
        
        const words = this.data.todayWords;
        words[index].reviewMode = true;
        
        // 掩盖单词和释义
        const wordText = words[index].word || '';
        const meaningText = words[index].meaning || '';
        
        // 生成掩盖的单词（用下划线代替字母）
        // 生成掩盖的单词（保留第一个和最后一个字母）
        let maskedWord = '';
        if (wordText.length <= 2) {
          // 如果单词长度小于等于2，全部显示
          maskedWord = wordText;
        } else {
          // 保留第一个和最后一个字母，中间用下划线
          const firstChar = wordText[0];
          const lastChar = wordText[wordText.length - 1];
          const middleChars = '_'.repeat(wordText.length - 2);
          maskedWord = firstChar + middleChars + lastChar;
        }
        words[index].maskedWord = maskedWord;
        // 生成掩盖的释义（用下划线代替字符）
        words[index].maskedMeaning = meaningText.replace(/./g, '_');
        
        // 处理题目数据，确保字段名正确
        const questions = (res.questions || []).map((q, idx) => {
          const question = {
            type: q.type || q.Type || '',
            question: q.question || q.Question || '',
            options: q.options || q.Options || [],
            correct_answer: q.correct_answer || q.correctAnswer || q.CorrectAnswer || ''
          };
          console.log(`题目 ${idx + 1}:`, question);
          return question;
        });
        
        if (questions.length === 0) {
          console.error('没有获取到题目数据，原始响应:', res);
          wx.showToast({
            title: '未获取到题目',
            icon: 'none',
            duration: 2000
          });
          return;
        }
        
        words[index].questions = questions;
        words[index].currentQuestionIndex = 0;
        words[index].userAnswers = [];
        words[index].showAnswers = [];
        words[index].canSubmit = true; // 可以提交答案
        words[index].spellingMode = false; // 确保spellingMode为false
        words[index].hasAnswered = false; // 重置答题状态
        words[index].allCorrect = false; // 重置正确状态
        
        console.log('处理后的题目:', questions);
        console.log('设置reviewMode为true, spellingMode为false');
        console.log('wordItem状态:', {
          reviewMode: words[index].reviewMode,
          spellingMode: words[index].spellingMode,
          questionsCount: words[index].questions.length
        });
        
        this.setData({
          todayWords: words
        }, () => {
          console.log('setData完成后的状态:', {
            reviewMode: this.data.todayWords[index].reviewMode,
            spellingMode: this.data.todayWords[index].spellingMode,
            questionsCount: this.data.todayWords[index].questions ? this.data.todayWords[index].questions.length : 0
          });
        });
      })
      .catch(err => {
        wx.hideLoading();
        console.error('生成题目失败:', err);
        wx.showToast({
          title: err.message || '生成题目失败',
          icon: 'none',
          duration: 2000
        });
      });
  },

  /**
   * 填空题输入（只允许英文字母，防止输入法联想补全）
   */
  onFillBlankInput(e) {
    const index = e.currentTarget.dataset.index;
    const qIndex = e.currentTarget.dataset.qIndex;
    let newValue = e.detail.value;
    const words = this.data.todayWords;
    
    // 获取当前已有的答案
    const oldValue = (words[index].userAnswers && words[index].userAnswers[qIndex]) || '';
    
    // 先过滤掉所有非英文字母
    newValue = newValue.replace(/[^a-zA-Z]/g, '');
    
    // 防止输入法联想补全的智能处理
    if (oldValue) {
      const oldLen = oldValue.length;
      const newLen = newValue.length;
      
      if (newLen > oldLen + 1) {
        // 输入了多个字符，可能是联想补全
        // 策略：只保留旧值 + 第一个新增的字符（用户实际输入的）
        if (newValue.indexOf(oldValue) === 0) {
          // 新值以旧值开头，提取第一个新增字符
          const firstNewChar = newValue[oldLen];
          if (firstNewChar && /[a-zA-Z]/.test(firstNewChar)) {
            newValue = oldValue + firstNewChar;
          } else {
            newValue = oldValue;
          }
        } else {
          // 新值不以旧值开头，可能是选择了联想词，只保留最后一个字符
          const lastChar = newValue.slice(-1);
          if (/[a-zA-Z]/.test(lastChar)) {
            newValue = lastChar;
          } else {
            newValue = oldValue;
          }
        }
      } else if (newLen === oldLen + 1) {
        // 正常输入一个字符
        const addedChar = newValue.slice(oldLen);
        if (!/[a-zA-Z]/.test(addedChar)) {
          newValue = oldValue;
        }
        // 否则 newValue 已经是正确的值
      }
      // 如果 newLen <= oldLen，说明是删除操作，newValue 已经是正确的值
    } else {
      // 首次输入
      if (newValue.length > 1) {
        // 首次输入多个字符，可能是联想补全，只保留第一个字符
        const firstChar = newValue[0];
        if (firstChar && /[a-zA-Z]/.test(firstChar)) {
          newValue = firstChar;
        } else {
          newValue = '';
        }
      }
    }
    
    // 最终确保只包含英文字母
    newValue = newValue.replace(/[^a-zA-Z]/g, '');
    
    if (!words[index].userAnswers) {
      words[index].userAnswers = [];
    }
    words[index].userAnswers[qIndex] = newValue;
    
    // 更新是否可以提交的状态
    this.updateCanSubmitStatus(words, index);
    
    this.setData({
      todayWords: words
    });
  },

  /**
   * 填空题失去焦点时，再次清理输入内容
   */
  onFillBlankBlur(e) {
    const index = e.currentTarget.dataset.index;
    const qIndex = e.currentTarget.dataset.qIndex;
    const words = this.data.todayWords;
    
    if (words[index].userAnswers && words[index].userAnswers[qIndex]) {
      // 确保只包含英文字母
      let value = words[index].userAnswers[qIndex];
      value = value.replace(/[^a-zA-Z]/g, '');
      words[index].userAnswers[qIndex] = value;
      
      this.setData({
        todayWords: words
      });
    }
  },

  /**
   * 选择题选择选项
   */
  selectOption(e) {
    const index = e.currentTarget.dataset.index;
    const qIndex = e.currentTarget.dataset.qIndex;
    const option = e.currentTarget.dataset.option;
    
    if (index === undefined || qIndex === undefined || option === undefined) {
      console.error('selectOption: 缺少必要参数', { index, qIndex, option });
      return;
    }
    
    const words = this.data.todayWords;
    
    if (!words[index]) {
      console.error('selectOption: 单词不存在', { index, wordsLength: words.length });
      return;
    }
    
    // 如果已经显示答案，不允许修改
    if (words[index].showAnswers && words[index].showAnswers[qIndex]) {
      return;
    }
    
    if (!words[index].userAnswers) {
      words[index].userAnswers = [];
    }
    words[index].userAnswers[qIndex] = option;
    
    console.log('选择选项:', { index, qIndex, option, userAnswer: words[index].userAnswers[qIndex] });
    
    // 更新是否可以提交的状态
    this.updateCanSubmitStatus(words, index);
    
    this.setData({
      todayWords: words
    });
  },

  /**
   * 更新是否可以提交的状态
   */
  updateCanSubmitStatus(words, index) {
    const word = words[index];
    if (!word.questions || word.questions.length === 0) {
      word.canSubmit = false;
      return;
    }
    
    // 检查是否所有题目都已作答
    const allAnswered = word.questions.every((q, qIndex) => {
      return word.userAnswers && word.userAnswers[qIndex] && word.userAnswers[qIndex].trim() !== '';
    });
    
    // 检查是否已经显示答案（如果已经提交，就不能再提交）
    const hasShownAnswers = word.showAnswers && word.showAnswers.some(show => show === true);
    
    word.canSubmit = allAnswered && !hasShownAnswers;
  },

  /**
   * 检查所有答案
   */
  checkAllAnswers(e) {
    const index = e.currentTarget.dataset.index;
    const words = this.data.todayWords;
    const word = words[index];
    
    // 检查是否所有题目都已作答
    const allAnswered = word.questions.every((q, qIndex) => {
      return word.userAnswers && word.userAnswers[qIndex] && word.userAnswers[qIndex].trim() !== '';
    });
    
    if (!allAnswered) {
      wx.showToast({
        title: '请完成所有题目',
        icon: 'none',
        duration: 1500
      });
      return;
    }
    
    // 显示所有答案
    if (!word.showAnswers) {
      word.showAnswers = [];
    }
    word.questions.forEach((q, qIndex) => {
      word.showAnswers[qIndex] = true;
    });
    
    // 设置为不能提交（已提交）
    word.canSubmit = false;
    
    // 计算正确数量（不区分大小写比较）
    let allCorrect = true; // 所有题目是否都正确
    
    // 检查所有题目的答案
    word.questions.forEach((q, qIndex) => {
      let userAnswer = (word.userAnswers[qIndex] || '').trim().toLowerCase();
      let correctAnswer = (q.correct_answer || '').trim().toLowerCase();
      
      // 移除所有非英文字母字符进行比较
      userAnswer = userAnswer.replace(/[^a-zA-Z]/g, '');
      correctAnswer = correctAnswer.replace(/[^a-zA-Z]/g, '');
      
      const isCorrect = userAnswer === correctAnswer;
      
      // 如果任何一题错误，则不是全部正确
      if (!isCorrect) {
        allCorrect = false;
      }
    });
    
    // 计算答对的数量（用于显示）
    const correctCount = word.questions.filter((q, qIndex) => {
      let userAnswer = (word.userAnswers[qIndex] || '').trim().toLowerCase();
      let correctAnswer = (q.correct_answer || '').trim().toLowerCase();
      
      // 移除所有非英文字母字符进行比较
      userAnswer = userAnswer.replace(/[^a-zA-Z]/g, '');
      correctAnswer = correctAnswer.replace(/[^a-zA-Z]/g, '');
      
      return userAnswer === correctAnswer;
    }).length;
    
    this.setData({
      todayWords: words
    });
    
    // 标记是否答错（用于显示重新生成按钮）
    words[index].hasAnswered = true;
    words[index].allCorrect = allCorrect;
    
    this.setData({
      todayWords: words
    });
    
    // 显示结果
    wx.showToast({
      title: `答对 ${correctCount}/${word.questions.length} 题`,
      icon: allCorrect ? 'success' : 'none',
      duration: 2000
    });
    
    // 如果所有题目都答对了，自动标记为已复习
    if (allCorrect && word.questions.length > 0) {
      console.log('所有题目都答对了，自动标记为已复习', {
        wordId: word.id,
        correctCount: correctCount,
        totalQuestions: word.questions.length
      });
      setTimeout(() => {
        this.autoMarkAsReviewed(word.id, index);
      }, 2000); // 等待toast显示完成
    } else {
      console.log('答案不正确，不自动标记为已复习', {
        allCorrect: allCorrect,
        correctCount: correctCount,
        totalQuestions: word.questions.length
      });
    }
  },

  /**
   * 重新生成题目
   */
  regenerateQuestions(e) {
    const index = e.currentTarget.dataset.index;
    const word = this.data.todayWords[index];
    
    // 获取年级信息
    const grade = wx.getStorageSync('grade') || '';
    
    wx.showLoading({
      title: '重新生成题目中...',
      mask: true
    });

    // 调用后端API重新生成题目
    api.generateReviewQuestions(this.data.activePlanId, word.id, grade)
      .then(res => {
        wx.hideLoading();
        
        console.log('重新生成题目API响应:', res);
        console.log('题目数据:', res.questions);
        
        const words = this.data.todayWords;
        
        // 处理题目数据，确保字段名正确
        const questions = (res.questions || []).map((q, idx) => {
          const question = {
            type: q.type || q.Type || '',
            question: q.question || q.Question || '',
            options: q.options || q.Options || [],
            correct_answer: q.correct_answer || q.correctAnswer || q.CorrectAnswer || ''
          };
          console.log(`题目 ${idx + 1}:`, question);
          return question;
        });
        
        if (questions.length === 0) {
          console.error('没有获取到题目数据，原始响应:', res);
          wx.showToast({
            title: '未获取到题目',
            icon: 'none',
            duration: 2000
          });
          return;
        }
        
        // 重置题目相关状态
        words[index].questions = questions;
        words[index].currentQuestionIndex = 0;
        words[index].userAnswers = [];
        words[index].showAnswers = [];
        words[index].canSubmit = true; // 可以提交答案
        words[index].hasAnswered = false; // 重置答题状态
        words[index].allCorrect = false; // 重置正确状态
        
        console.log('重新生成后的题目:', questions);
        
        this.setData({
          todayWords: words
        });
        
        wx.showToast({
          title: '已重新生成题目',
          icon: 'success',
          duration: 1500
        });
      })
      .catch(err => {
        wx.hideLoading();
        console.error('重新生成题目失败:', err);
        wx.showToast({
          title: err.message || '重新生成题目失败',
          icon: 'none',
          duration: 2000
        });
      });
  },

  /**
   * 自动标记为已复习
   */
  autoMarkAsReviewed(wordId, index) {
    wx.showLoading({
      title: '标记已复习...',
      mask: true
    });

      api.markWordReviewed(this.data.activePlanId, wordId)
      .then(res => {
        wx.hideLoading();
        wx.showToast({
          title: '已自动标记为复习',
          icon: 'success',
          duration: 1500
        });

        // 更新单词状态
        const words = this.data.todayWords;
        if (words[index]) {
          const today = new Date();
          const todayStr = ebbinghaus.formatDate(today);
          words[index].review_count = (words[index].review_count || 0) + 1;
          words[index].last_review_date = todayStr;
          words[index].isReviewedToday = true; // 标记为今日已复习
          
          // 立即更新状态，显示"今日已复习"标识
          this.setData({
            todayWords: words
          });
        }

        // 重新加载列表以获取最新状态
        setTimeout(() => {
          this.loadTodayWords();
        }, 500);
      })
      .catch(err => {
        wx.hideLoading();
        console.error('自动标记失败:', err);
        wx.showToast({
          title: '标记失败: ' + (err.message || '未知错误'),
          icon: 'none',
          duration: 2000
        });
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
   * 切换单词显示/隐藏
   */
  toggleWordDisplay(e) {
    const index = e.currentTarget.dataset.index;
    const words = this.data.todayWords;
    
    if (words[index]) {
      words[index].showWord = !words[index].showWord;
      
      this.setData({
        todayWords: words
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
    words[index].reviewMode = false;
    words[index].questions = [];
    words[index].currentQuestionIndex = 0;
    words[index].userAnswers = [];
    words[index].showAnswers = [];
    words[index].canSubmit = false;
    words[index].userAnswer = '';
    words[index].showAnswer = false;
    words[index].hasAnswered = false; // 重置答题状态
    words[index].allCorrect = false; // 重置正确状态
    // 清除掩盖信息
    delete words[index].maskedWord;
    delete words[index].maskedMeaning;
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
      api.markWordReviewed(this.data.activePlanId, wordId)
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
          title: err.message || '标记失败',
          icon: 'none'
        });
      })
      .finally(() => {
        wx.hideLoading();
      });
  },

  /**
   * 为所有单词加载混淆词列表
   */
  loadConfusedWordsForAll(words) {
    const promises = words.map((word, index) => {
      // 如果已经有混淆词列表，跳过
      if (word.confused_words && word.confused_words.length > 0) {
        return Promise.resolve(word);
      }
      
      return api.getConfusedWords(this.data.activePlanId, word.id)
        .then(res => {
          words[index].confused_words = res.confused_words || [];
          return words[index];
        })
        .catch(err => {
          console.warn(`加载单词 ${word.id} 的混淆词失败:`, err);
          words[index].confused_words = [];
          return words[index];
        });
    });

    Promise.all(promises).then(() => {
      this.setData({
        todayWords: words,
        loading: false
      });
    });
  },

  /**
   * 显示添加混淆词弹窗
   */
  showAddConfusedWord(e) {
    const index = e.currentTarget.dataset.index;
    const wordId = e.currentTarget.dataset.wordId;
    
    this.setData({
      showConfusedWordModal: true,
      currentWordIndex: index,
      currentWordId: wordId,
      confusedWordKeyword: '',
      confusedWordSearchResults: []
    });
  },

  /**
   * 隐藏添加混淆词弹窗
   */
  hideAddConfusedWord() {
    this.setData({
      showConfusedWordModal: false,
      currentWordIndex: -1,
      currentWordId: null,
      confusedWordKeyword: '',
      confusedWordSearchResults: []
    });
  },

  /**
   * 阻止事件冒泡
   */
  stopPropagation() {
    // 空函数，用于阻止点击弹窗内容时关闭弹窗
  },

  /**
   * 混淆词搜索输入
   */
  onConfusedWordSearch(e) {
    this.setData({
      confusedWordKeyword: e.detail.value
    });
  },

  /**
   * 搜索混淆词
   */
  searchConfusedWords() {
    const keyword = this.data.confusedWordKeyword.trim();
    if (!keyword) {
      wx.showToast({
        title: '请输入搜索关键词',
        icon: 'none'
      });
      return;
    }

    wx.showLoading({
      title: '搜索中...',
      mask: true
    });

    api.searchWords(this.data.activePlanId, keyword, 20)
      .then(res => {
        // 过滤掉当前单词本身
        const currentWordId = this.data.currentWordId;
        const results = (res.words || []).filter(word => word.id !== currentWordId);
        
        this.setData({
          confusedWordSearchResults: results
        });
        
        wx.hideLoading();
        
        if (results.length === 0) {
          wx.showToast({
            title: '未找到匹配的单词',
            icon: 'none'
          });
        }
      })
      .catch(err => {
        console.error('搜索失败:', err);
        wx.hideLoading();
        wx.showToast({
          title: err.message || '搜索失败',
          icon: 'none'
        });
      });
  },

  /**
   * 添加混淆词
   */
  addConfusedWord(e) {
    const confusedWordId = e.currentTarget.dataset.confusedWordId;
    const wordId = this.data.currentWordId;
    const index = this.data.currentWordIndex;

    wx.showLoading({
      title: '添加中...',
      mask: true
    });

    api.addConfusedWord(this.data.activePlanId, wordId, confusedWordId)
      .then(res => {
        wx.showToast({
          title: '添加成功',
          icon: 'success'
        });

        // 更新单词的混淆词列表
        const words = this.data.todayWords;
        if (words[index]) {
          // 重新加载混淆词列表
          api.getConfusedWords(this.data.activePlanId, wordId)
            .then(confusedRes => {
              words[index].confused_words = confusedRes.confused_words || [];
              this.setData({
                todayWords: words
              });
            })
            .catch(err => {
              console.warn('重新加载混淆词失败:', err);
            });
        }

        // 从搜索结果中移除已添加的单词
        const results = this.data.confusedWordSearchResults.filter(w => w.id !== confusedWordId);
        this.setData({
          confusedWordSearchResults: results
        });

        wx.hideLoading();
      })
      .catch(err => {
        console.error('添加混淆词失败:', err);
        wx.hideLoading();
        wx.showToast({
          title: err.message || '添加失败',
          icon: 'none'
        });
      });
  },

  /**
   * 播放单词发音
   */
  playAudio(e) {
    const wordId = e.currentTarget.dataset.wordId;
    const index = e.currentTarget.dataset.index;
    const word = this.data.todayWords[index];
    
    // 检查是否有音频 URL
    if (!word.audio_urls || word.audio_urls.length === 0) {
      wx.showToast({
        title: '暂无发音',
        icon: 'none',
        duration: 1500
      });
      return;
    }

    // 确保 audioContext 已初始化
    if (!this.data.audioContext) {
      try {
        this.data.audioContext = wx.createInnerAudioContext();
        if (!this.data.audioContext) {
          console.error('创建音频上下文失败: 返回 null');
          wx.showToast({
            title: '音频初始化失败',
            icon: 'none',
            duration: 1500
          });
          return;
        }
        // 设置事件监听器
        this.data.audioContext.onEnded(() => {
          this.setData({
            playingWordId: null
          });
        });
        this.data.audioContext.onError((err) => {
          console.error('音频播放失败:', err);
          wx.showToast({
            title: '播放失败',
            icon: 'none',
            duration: 1500
          });
          this.setData({
            playingWordId: null
          });
        });
        console.log('音频上下文已创建并初始化');
      } catch (err) {
        console.error('创建音频上下文异常:', err);
        this.data.audioContext = null;
        wx.showToast({
          title: '音频初始化失败',
          icon: 'none',
          duration: 1500
        });
        return;
      }
    }

    // 如果正在播放同一个单词，则停止播放
    if (this.data.playingWordId === wordId) {
      if (this.data.audioContext) {
        try {
          this.data.audioContext.stop();
        } catch (err) {
          console.error('停止音频失败:', err);
        }
      }
      this.setData({
        playingWordId: null
      });
      return;
    }

    // 停止当前播放的音频
    if (this.data.audioContext) {
      try {
        this.data.audioContext.stop();
      } catch (err) {
        console.error('停止音频失败:', err);
      }
    }

    // 优先选择 Cambridge 字典的音频，否则使用第一个
    let audioUrl = null;
    const cambridgeUrl = word.audio_urls.find(url => 
      url && typeof url === 'string' && url.indexOf('http://dictionary.cambridge.org') === 0
    );
    if (cambridgeUrl) {
      audioUrl = cambridgeUrl;
    } else if (word.audio_urls.length > 0) {
      audioUrl = word.audio_urls[0];
    }
    
    if (!audioUrl) {
      wx.showToast({
        title: '暂无发音',
        icon: 'none',
        duration: 1500
      });
      return;
    }
    
    // 再次确保 audioContext 存在且有效
    if (!this.data.audioContext) {
      console.error('audioContext 未初始化');
      wx.showToast({
        title: '音频初始化失败',
        icon: 'none',
        duration: 1500
      });
      return;
    }
    
    // 设置音频源并播放
    try {
      // 再次验证 audioContext 存在且有效
      if (!this.data.audioContext) {
        console.error('audioContext 在播放前变为 null，尝试重新创建');
        try {
          this.data.audioContext = wx.createInnerAudioContext();
          if (!this.data.audioContext) {
            throw new Error('创建音频上下文失败');
          }
          this.data.audioContext.onEnded(() => {
            this.setData({
              playingWordId: null
            });
          });
          this.data.audioContext.onError((err) => {
            console.error('音频播放失败:', err);
            this.setData({
              playingWordId: null
            });
          });
        } catch (createErr) {
          console.error('重新创建音频上下文失败:', createErr);
          wx.showToast({
            title: '音频初始化失败',
            icon: 'none',
            duration: 1500
          });
          return;
        }
      }
      
      // 先停止之前的播放（如果有）
      if (this.data.audioContext.src) {
        try {
          this.data.audioContext.stop();
        } catch (stopErr) {
          console.warn('停止之前的音频失败:', stopErr);
        }
      }
      
      // 设置新的音频源
      this.data.audioContext.src = audioUrl;
      
      // 设置播放状态
      this.setData({
        playingWordId: wordId
      });
      
      // 延迟播放，确保音频源设置完成
      setTimeout(() => {
        try {
          // 再次验证 audioContext 仍然有效
          if (!this.data.audioContext) {
            console.error('audioContext 在延迟播放时变为 null');
            this.setData({
              playingWordId: null
            });
            return;
          }
          
          // 验证音频源是否正确设置
          if (this.data.audioContext.src !== audioUrl) {
            console.warn('音频源不匹配，重新设置');
            this.data.audioContext.src = audioUrl;
          }
          
          // 播放音频
          this.data.audioContext.play();
          console.log('开始播放音频:', audioUrl);
        } catch (playErr) {
          console.error('播放音频异常:', playErr);
          wx.showToast({
            title: '播放失败',
            icon: 'none',
            duration: 1500
          });
          this.setData({
            playingWordId: null
          });
        }
      }, 100); // 延迟100ms确保音频源设置完成
      
    } catch (err) {
      console.error('设置音频源失败:', err);
      wx.showToast({
        title: '播放失败',
        icon: 'none',
        duration: 1500
      });
      this.setData({
        playingWordId: null
      });
    }
  },

  /**
   * 删除混淆词
   */
  removeConfusedWord(e) {
    const wordId = e.currentTarget.dataset.wordId;
    const confusedWordId = e.currentTarget.dataset.confusedWordId;
    const index = e.currentTarget.dataset.index;

    wx.showModal({
      title: '确认删除',
      content: '确定要删除这个混淆词吗？',
      success: (res) => {
        if (res.confirm) {
          wx.showLoading({
            title: '删除中...',
            mask: true
          });

          api.removeConfusedWord(wordId, confusedWordId)
            .then(() => {
              wx.showToast({
                title: '删除成功',
                icon: 'success'
              });

              // 更新单词的混淆词列表
              const words = this.data.todayWords;
              if (words[index]) {
                words[index].confused_words = words[index].confused_words.filter(
                  w => w.id !== confusedWordId
                );
                this.setData({
                  todayWords: words
                });
              }

              wx.hideLoading();
            })
            .catch(err => {
              console.error('删除混淆词失败:', err);
              wx.hideLoading();
              wx.showToast({
                title: err.message || '删除失败',
                icon: 'none'
              });
            });
        }
      }
    });
  }
});
