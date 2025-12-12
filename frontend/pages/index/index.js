// pages/index/index.js
const api = require('../../utils/api.js');
const auth = require('../../utils/auth.js');
const ebbinghaus = require('../../utils/ebbinghaus.js');

Page({
  data: {
    todayWords: [],
    todayDate: '',
    loading: false,
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
    this.loadTodayWords();
    // 创建音频上下文
    this.data.audioContext = wx.createInnerAudioContext();
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
  },

  onUnload() {
    // 页面卸载时销毁音频上下文
    if (this.data.audioContext) {
      this.data.audioContext.destroy();
      this.data.audioContext = null;
    }
  },

  onShow() {
    // 每次显示页面时重新加载，确保数据是最新的
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

  /**
   * 加载今天需要背诵的单词
   */
  loadTodayWords() {
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

    this.setData({ loading: true });

    const today = new Date();
    const todayStr = ebbinghaus.formatDate(today);

    // 从后端获取今日单词
    api.getTodayWords(todayStr)
      .then(res => {
        // 为每个单词添加拼写模式相关状态
        const wordsWithState = (res.words || []).map(word => {
          // 兼容两种字段名：audio_urls（下划线）和 audioUrls（驼峰）
          const audioUrls = word.audio_urls || word.audioUrls || [];
          
          // 调试日志：检查音频字段
          if (audioUrls.length > 0) {
            console.log(`单词 "${word.word}" 有 ${audioUrls.length} 个音频:`, audioUrls);
          }
          
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
            isCorrect: false
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

        this.setData({
          todayWords: wordsWithState,
          todayDate: res.date || todayStr,
          loading: false
        });
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
    api.generateReviewQuestions(word.id, grade)
      .then(res => {
        wx.hideLoading();
        
        const words = this.data.todayWords;
        words[index].reviewMode = true;
        words[index].questions = res.questions || [];
        words[index].currentQuestionIndex = 0;
        words[index].userAnswers = [];
        words[index].showAnswers = [];
        words[index].canSubmit = true; // 可以提交答案
        
        this.setData({
          todayWords: words
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
   * 填空题输入
   */
  onFillBlankInput(e) {
    const index = e.currentTarget.dataset.index;
    const qIndex = e.currentTarget.dataset.qIndex;
    const value = e.detail.value;
    const words = this.data.todayWords;
    
    if (!words[index].userAnswers) {
      words[index].userAnswers = [];
    }
    words[index].userAnswers[qIndex] = value;
    
    // 更新是否可以提交的状态
    this.updateCanSubmitStatus(words, index);
    
    this.setData({
      todayWords: words
    });
  },

  /**
   * 选择题选择选项
   */
  selectOption(e) {
    const index = e.currentTarget.dataset.index;
    const qIndex = e.currentTarget.dataset.qIndex;
    const option = e.currentTarget.dataset.option;
    const words = this.data.todayWords;
    
    // 如果已经显示答案，不允许修改
    if (words[index].showAnswers && words[index].showAnswers[qIndex]) {
      return;
    }
    
    if (!words[index].userAnswers) {
      words[index].userAnswers = [];
    }
    words[index].userAnswers[qIndex] = option;
    
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
    
    // 计算正确数量
    const correctCount = word.questions.filter((q, qIndex) => {
      return word.userAnswers[qIndex] === q.correct_answer;
    }).length;
    
    this.setData({
      todayWords: words
    });
    
    // 显示结果
    wx.showToast({
      title: `答对 ${correctCount}/${word.questions.length} 题`,
      icon: correctCount === word.questions.length ? 'success' : 'none',
      duration: 2000
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
    words[index].reviewMode = false;
    words[index].questions = [];
    words[index].currentQuestionIndex = 0;
    words[index].userAnswers = [];
    words[index].showAnswers = [];
    words[index].canSubmit = false;
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
      
      return api.getConfusedWords(word.id)
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

    api.searchWords(keyword, 20)
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

    api.addConfusedWord(wordId, confusedWordId)
      .then(res => {
        wx.showToast({
          title: '添加成功',
          icon: 'success'
        });

        // 更新单词的混淆词列表
        const words = this.data.todayWords;
        if (words[index]) {
          // 重新加载混淆词列表
          api.getConfusedWords(wordId)
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

    // 如果正在播放同一个单词，则停止播放
    if (this.data.playingWordId === wordId && this.data.audioContext) {
      this.data.audioContext.stop();
      this.setData({
        playingWordId: null
      });
      return;
    }

    // 停止当前播放的音频
    if (this.data.audioContext) {
      this.data.audioContext.stop();
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
    
    // 设置音频源并播放
    this.data.audioContext.src = audioUrl;
    this.data.audioContext.play();
    
    this.setData({
      playingWordId: wordId
    });
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
