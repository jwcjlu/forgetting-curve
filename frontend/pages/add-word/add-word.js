// pages/add-word/add-word.js
const api = require('../../utils/api.js');
const auth = require('../../utils/auth.js');
const ebbinghaus = require('../../utils/ebbinghaus.js');

Page({
  data: {
    word: '',
    meaning: '',
    startDate: '',
    minDate: '',
    words: [],
    totalWords: 0,
    page: 1,
    pageSize: 20,
    hasMore: false,
    loading: false,
    loadingMore: false,
    submitting: false,
    // OCR相关
    recognizing: false,
    showOCRResult: false,
    recognizedWords: [],
    // 音频播放相关
    audioContext: null,
    playingWordId: null
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
    
    this.checkLoginAndLoadWords();
  },

  onUnload() {
    // 页面卸载时销毁音频上下文
    if (this.data.audioContext) {
      this.data.audioContext.destroy();
      this.data.audioContext = null;
    }
  },

  onShow() {
    // 每次显示页面时重新加载单词列表
    this.checkLoginAndLoadWords();
  },

  /**
   * 检查登录并加载单词
   */
  checkLoginAndLoadWords() {
    const studentId = api.getStudentId();
    if (!studentId) {
      // 尝试自动登录
      auth.autoLogin()
        .then(() => {
          this.loadAllWords();
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
    } else {
      this.loadAllWords();
    }
  },

  /**
   * 加载所有单词
   */
  loadAllWords() {
    this.setData({ loading: true, page: 1 });
    
    api.getStudentWords(1, this.data.pageSize)
      .then(res => {
        // 处理单词列表，兼容 audio_urls 和 audioUrls 两种字段名
        const words = (res.words || []).map(word => {
          // 兼容两种字段名：audio_urls（下划线）和 audioUrls（驼峰）
          const audioUrls = word.audio_urls || word.audioUrls || [];
          
          // 调试日志：检查音频字段
          if (audioUrls.length > 0) {
            console.log(`单词 "${word.word}" 有 ${audioUrls.length} 个音频:`, audioUrls);
          }
          
          return {
            ...word,
            audio_urls: audioUrls // 统一使用下划线命名
          };
        });
        
        this.setData({
          words: words,
          totalWords: res.total || 0,
          hasMore: words.length < (res.total || 0),
          loading: false
        });
      })
      .catch(err => {
        console.error('加载单词失败:', err);
        wx.showToast({
          title: err.message || '加载失败',
          icon: 'none'
        });
        this.setData({
          loading: false
        });
      });
  },

  /**
   * 加载更多单词
   */
  loadMoreWords() {
    if (this.data.loadingMore || !this.data.hasMore) {
      return;
    }

    this.setData({ loadingMore: true });
    const nextPage = this.data.page + 1;

    api.getStudentWords(nextPage, this.data.pageSize)
      .then(res => {
        const newWords = res.words || [];
        this.setData({
          words: [...this.data.words, ...newWords],
          page: nextPage,
          hasMore: newWords.length === this.data.pageSize && 
                   (this.data.words.length + newWords.length) < (res.total || 0),
          loadingMore: false
        });
      })
      .catch(err => {
        console.error('加载更多失败:', err);
        wx.showToast({
          title: err.message || '加载失败',
          icon: 'none'
        });
        this.setData({
          loadingMore: false
        });
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
    let word = this.data.word.trim();
    let meaning = this.data.meaning.trim();
    const startDate = this.data.startDate;

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

    this.setData({ submitting: true });

    // 调用后端API批量添加单词
    api.batchAddWords([{
      word: word,
      meaning: meaning,
      start_date: startDate
    }])
      .then(res => {
        wx.showToast({
          title: '添加成功',
          icon: 'success'
        });

        // 清空表单
        this.setData({
          word: '',
          meaning: '',
          submitting: false
        });

        // 重新加载列表
        this.loadAllWords();
      })
      .catch(err => {
        console.error('添加失败:', err);
        wx.showToast({
          title: err.message || '添加失败',
          icon: 'none'
        });
        this.setData({
          submitting: false
        });
      });
  },

  /**
   * 选择图片（拍照或从相册）
   */
  chooseImage() {
    const studentId = api.getStudentId();
    if (!studentId) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      });
      return;
    }

    wx.chooseMedia({
      count: 1,
      mediaType: ['image'],
      sourceType: ['camera', 'album'],
      camera: 'back',
      success: (res) => {
        const tempFilePath = res.tempFiles[0].tempFilePath;
        this.recognizeImage(tempFilePath);
      },
      fail: (err) => {
        console.error('选择图片失败:', err);
        wx.showToast({
          title: '选择图片失败',
          icon: 'none'
        });
      }
    });
  },

  /**
   * 识别图片中的单词
   */
  recognizeImage(imagePath) {
    this.setData({ recognizing: true });

    // 将图片转换为base64
    const fs = wx.getFileSystemManager();
    fs.readFile({
      filePath: imagePath,
      encoding: 'base64',
      success: (res) => {
        const imageBase64 = 'data:image/jpeg;base64,' + res.data;
        
        // 调用OCR API
        api.recognizeWordsFromImage(imageBase64, this.data.startDate)
          .then(result => {
            const words = (result.recognized_words || []).map(w => {
              const confidence = w.confidence || 0.8;
              return {
                word: w.word,
                meaning: w.meaning || '',
                confidence: confidence,
                confidence_text: Math.round(confidence * 100) + '%' // 预处理百分比字符串
              };
            });

            this.setData({
              recognizedWords: words,
              showOCRResult: true,
              recognizing: false
            });

            if (words.length === 0) {
              wx.showToast({
                title: '未识别到单词',
                icon: 'none'
              });
            }
          })
          .catch(err => {
            console.error('OCR识别失败:', err);
            wx.showToast({
              title: err.message || '识别失败',
              icon: 'none'
            });
            this.setData({ recognizing: false });
          });
      },
      fail: (err) => {
        console.error('读取图片失败:', err);
        wx.showToast({
          title: '读取图片失败',
          icon: 'none'
        });
        this.setData({ recognizing: false });
      }
    });
  },

  /**
   * OCR释义输入
   */
  onOCRMeaningInput(e) {
    const index = e.currentTarget.dataset.index;
    const value = e.detail.value;
    const words = this.data.recognizedWords;
    words[index].meaning = value;
    this.setData({
      recognizedWords: words
    });
  },

  /**
   * 添加识别到的单词
   */
  addRecognizedWords() {
    const words = this.data.recognizedWords.filter(w => w.word.trim() !== '');
    if (words.length === 0) {
      wx.showToast({
        title: '没有可添加的单词',
        icon: 'none'
      });
      return;
    }

    this.setData({ submitting: true });

    // 转换为批量添加格式
    const wordItems = words.map(w => ({
      word: w.word.trim(),
      meaning: w.meaning.trim() || '（待补充）',
      start_date: this.data.startDate
    }));

    // 批量添加单词
    api.batchAddWords(wordItems)
      .then(res => {
        wx.showToast({
          title: `成功添加 ${words.length} 个单词`,
          icon: 'success',
          duration: 2000
        });

        // 关闭弹窗并重新加载列表
        this.setData({
          showOCRResult: false,
          recognizedWords: [],
          submitting: false
        });

        setTimeout(() => {
          this.loadAllWords();
        }, 500);
      })
      .catch(err => {
        console.error('添加失败:', err);
        wx.showToast({
          title: err.message || '添加失败',
          icon: 'none'
        });
        this.setData({ submitting: false });
      });
  },

  /**
   * 隐藏OCR结果弹窗
   */
  hideOCRResult() {
    this.setData({
      showOCRResult: false,
      recognizedWords: []
    });
  },

  /**
   * 播放单词发音
   */
  playAudio(e) {
    const wordId = e.currentTarget.dataset.wordId;
    const index = e.currentTarget.dataset.index;
    const word = this.data.words[index];
    
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
   * 阻止事件冒泡
   */
  stopPropagation() {
    // 空函数，用于阻止点击弹窗内容时关闭弹窗
  }
});
