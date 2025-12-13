// pages/plan-words/plan-words.js
const api = require('../../utils/api.js');

Page({
  data: {
    planId: null,
    planName: '',
    words: [],
    planWords: [], // 计划中已有的单词
    selectedWordIds: [], // 选中的单词ID
    totalWords: 0,
    page: 1,
    pageSize: 20,
    hasMore: false,
    loading: false,
    loadingMore: false,
    submitting: false
  },

  onLoad(options) {
    const planId = options.planId;
    if (!planId) {
      wx.showToast({
        title: '计划ID不能为空',
        icon: 'none'
      });
      setTimeout(() => {
        wx.navigateBack();
      }, 1500);
      return;
    }

    this.setData({ planId: parseInt(planId) });
    this.loadPlanInfo();
    this.loadPlanWords();
    this.loadWords();
  },

  /**
   * 加载计划信息
   */
  loadPlanInfo() {
    api.getPlans()
      .then(res => {
        const plan = (res.plans || []).find(p => p.id === this.data.planId);
        if (plan) {
          this.setData({ planName: plan.name });
          wx.setNavigationBarTitle({
            title: `添加单词 - ${plan.name}`
          });
        }
      })
      .catch(err => {
        console.error('加载计划信息失败:', err);
      });
  },

  /**
   * 加载计划中已有的单词
   */
  loadPlanWords() {
    // 使用分页接口获取计划中的单词
    api.getPlanWords(this.data.planId, 1, 1000)
      .then(res => {
        const planWordIds = (res.words || []).map(w => w.id);
        this.setData({ 
          planWords: res.words || [],
          selectedWordIds: planWordIds
        });
      })
      .catch(err => {
        console.error('加载计划单词失败:', err);
      });
  },

  /**
   * 加载所有单词列表
   */
  loadWords() {
    this.setData({ loading: true });

    // 使用计划的单词列表接口
    api.getPlanWords(this.data.planId, 1, this.data.pageSize)
      .then(res => {
        const words = res.words || [];
        this.setData({
          words: words,
          totalWords: res.total || 0,
          hasMore: words.length === this.data.pageSize && words.length < (res.total || 0),
          loading: false
        });
      })
      .catch(err => {
        console.error('加载单词列表失败:', err);
        wx.showToast({
          title: err.message || '加载失败',
          icon: 'none'
        });
        this.setData({ loading: false });
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

    api.getPlanWords(this.data.planId, nextPage, this.data.pageSize)
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
        this.setData({ loadingMore: false });
      });
  },

  /**
   * 切换单词选择状态
   */
  toggleWordSelect(e) {
    const wordId = e.currentTarget.dataset.wordId;
    const selectedWordIds = [...this.data.selectedWordIds];
    const index = selectedWordIds.indexOf(wordId);
    
    if (index >= 0) {
      selectedWordIds.splice(index, 1);
    } else {
      selectedWordIds.push(wordId);
    }
    
    this.setData({ selectedWordIds });
  },

  /**
   * 保存选中的单词到计划
   */
  saveWordsToPlan() {
    if (this.data.selectedWordIds.length === 0) {
      wx.showToast({
        title: '请至少选择一个单词',
        icon: 'none',
        duration: 2000
      });
      return;
    }

    this.setData({ submitting: true });

    api.addWordsToPlan(this.data.planId, this.data.selectedWordIds)
      .then(res => {
        wx.hideLoading();
        wx.showToast({
          title: `成功添加 ${res.added_count || this.data.selectedWordIds.length} 个单词`,
          icon: 'success',
          duration: 2000
        });
        
        // 重新加载计划单词
        this.loadPlanWords();
        
        // 延迟返回
        setTimeout(() => {
          wx.navigateBack();
        }, 2000);
      })
      .catch(err => {
        wx.hideLoading();
        console.error('添加单词到计划失败:', err);
        wx.showToast({
          title: err.message || '添加失败',
          icon: 'none',
          duration: 2000
        });
        this.setData({ submitting: false });
      });
  },

  /**
   * 从计划中移除单词
   */
  removeWordFromPlan(e) {
    const wordId = e.currentTarget.dataset.wordId;
    const word = this.data.planWords.find(w => w.id === wordId);
    
    if (!word) {
      return;
    }

    wx.showModal({
      title: '确认移除',
      content: `确定要从计划中移除单词"${word.word}"吗？`,
      success: (res) => {
        if (res.confirm) {
          wx.showLoading({
            title: '移除中...',
            mask: true
          });

          api.removeWordsFromPlan(this.data.planId, wordId)
            .then(() => {
              wx.hideLoading();
              wx.showToast({
                title: '移除成功',
                icon: 'success',
                duration: 1500
              });
              
              // 从选中列表中移除
              const selectedWordIds = this.data.selectedWordIds.filter(id => id !== wordId);
              this.setData({ selectedWordIds });
              
              // 重新加载计划单词
              this.loadPlanWords();
            })
            .catch(err => {
              wx.hideLoading();
              console.error('移除单词失败:', err);
              wx.showToast({
                title: err.message || '移除失败',
                icon: 'none',
                duration: 2000
              });
            });
        }
      }
    });
  }
});

