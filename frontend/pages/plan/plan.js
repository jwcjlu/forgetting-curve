// pages/plan/plan.js
const api = require('../../utils/api.js');

Page({
  data: {
    plans: [],
    loading: false,
    showCreateModal: false,
    showEditModal: false,
    editingPlanId: null,
    planName: '',
    planGrade: '',
    planGradeIndex: 0,
    grades: ['小学一年级', '小学二年级', '小学三年级', '小学四年级', '小学五年级', '小学六年级', '初中一年级', '初中二年级', '初中三年级', '高中一年级', '高中二年级', '高中三年级', '大学', '其他']
  },

  onLoad() {
    this.loadPlans();
  },

  onShow() {
    this.loadPlans();
  },

  /**
   * 加载计划列表
   */
  loadPlans() {
    this.setData({ loading: true });

    api.getPlans()
      .then(res => {
        console.log('获取计划列表成功:', res);
        const plans = res.plans || [];
        
        // 获取当前激活的计划ID
        const activePlanId = wx.getStorageSync('activePlanId');
        
        // 为每个计划加载单词数量
        const planPromises = plans.map(plan => {
          return api.getPlanWords(plan.id, 1, 1)
            .then(wordRes => {
              plan.wordCount = wordRes.total || 0;
              plan.is_active = (plan.id == activePlanId) || plan.is_active;
              return plan;
            })
            .catch(err => {
              console.warn(`获取计划 ${plan.id} 的单词数量失败:`, err);
              plan.wordCount = 0;
              plan.is_active = (plan.id == activePlanId) || plan.is_active;
              return plan;
            });
        });
        
        return Promise.all(planPromises);
      })
      .then(plans => {
        this.setData({
          plans: plans,
          loading: false
        });
      })
      .catch(err => {
        console.error('获取计划列表失败:', err);
        wx.showToast({
          title: err.message || '加载失败',
          icon: 'none',
          duration: 2000
        });
        this.setData({ loading: false });
      });
  },

  /**
   * 显示创建计划弹窗
   */
  showCreate() {
    this.setData({
      showCreateModal: true,
      planName: '',
      planGrade: '',
      planGradeIndex: 0
    });
  },

  /**
   * 关闭创建计划弹窗
   */
  hideCreate() {
    this.setData({ showCreateModal: false });
  },

  /**
   * 输入计划名称
   */
  onPlanNameInput(e) {
    this.setData({ planName: e.detail.value });
  },

  /**
   * 选择年级
   */
  onGradeChange(e) {
    const index = parseInt(e.detail.value);
    this.setData({ 
      planGradeIndex: index,
      planGrade: this.data.grades[index] || '' 
    });
  },

  /**
   * 创建计划
   */
  createPlan() {
    const name = this.data.planName.trim();
    if (!name) {
      wx.showToast({
        title: '请输入计划名称',
        icon: 'none',
        duration: 2000
      });
      return;
    }

    wx.showLoading({
      title: '创建中...',
      mask: true
    });

    api.createPlan(name, this.data.planGrade)
      .then(res => {
        wx.hideLoading();
        console.log('创建计划成功:', res);
        wx.showToast({
          title: '创建成功',
          icon: 'success',
          duration: 1500
        });
        this.hideCreate();
        this.loadPlans();
      })
      .catch(err => {
        wx.hideLoading();
        console.error('创建计划失败:', err);
        wx.showToast({
          title: err.message || '创建失败',
          icon: 'none',
          duration: 2000
        });
      });
  },

  /**
   * 选择计划（设置为当前激活的计划）
   */
  selectPlan(e) {
    const planId = e.currentTarget.dataset.planId;
    const plan = this.data.plans.find(p => p.id === planId);
    
    if (!plan) {
      return;
    }

    wx.showLoading({
      title: '切换中...',
      mask: true
    });

    api.selectPlan(planId)
      .then(res => {
        wx.hideLoading();
        console.log('选择计划成功:', res);
        wx.showToast({
          title: '已切换到计划：' + plan.name,
          icon: 'success',
          duration: 2000
        });
        
        // 保存当前激活的计划ID到本地存储
        wx.setStorageSync('activePlanId', planId);
        
        // 重新加载计划列表
        this.loadPlans();
        
        // 返回上一页（通常是首页）
        setTimeout(() => {
          wx.navigateBack();
        }, 500);
      })
      .catch(err => {
        wx.hideLoading();
        console.error('选择计划失败:', err);
        wx.showToast({
          title: err.message || '切换失败',
          icon: 'none',
          duration: 2000
        });
      });
  },

  /**
   * 删除计划
   */
  deletePlan(e) {
    const planId = e.currentTarget.dataset.planId;
    const plan = this.data.plans.find(p => p.id === planId);
    
    if (!plan) {
      return;
    }

    wx.showModal({
      title: '确认删除',
      content: `确定要删除计划"${plan.name}"吗？删除后无法恢复。`,
      success: (res) => {
        if (res.confirm) {
          wx.showLoading({
            title: '删除中...',
            mask: true
          });

          api.deletePlan(planId)
            .then(() => {
              wx.hideLoading();
              wx.showToast({
                title: '删除成功',
                icon: 'success',
                duration: 1500
              });
              
              // 如果删除的是当前激活的计划，清除本地存储
              const activePlanId = wx.getStorageSync('activePlanId');
              if (activePlanId == planId) {
                wx.removeStorageSync('activePlanId');
              }
              
              this.loadPlans();
            })
            .catch(err => {
              wx.hideLoading();
              console.error('删除计划失败:', err);
              wx.showToast({
                title: err.message || '删除失败',
                icon: 'none',
                duration: 2000
              });
            });
        }
      }
    });
  },

  /**
   * 编辑计划
   */
  editPlan(e) {
    const planId = e.currentTarget.dataset.planId;
    const plan = this.data.plans.find(p => p.id === planId);
    
    if (!plan) {
      return;
    }

    // 找到年级索引
    const gradeIndex = this.data.grades.findIndex(g => g === plan.grade);
    
    this.setData({
      showEditModal: true,
      editingPlanId: planId,
      planName: plan.name,
      planGrade: plan.grade || '',
      planGradeIndex: gradeIndex >= 0 ? gradeIndex : 0
    });
  },

  /**
   * 关闭编辑计划弹窗
   */
  hideEdit() {
    this.setData({ 
      showEditModal: false,
      editingPlanId: null,
      planName: '',
      planGrade: '',
      planGradeIndex: 0
    });
  },

  /**
   * 更新计划
   */
  updatePlan() {
    const name = this.data.planName.trim();
    if (!name) {
      wx.showToast({
        title: '请输入计划名称',
        icon: 'none',
        duration: 2000
      });
      return;
    }

    if (!this.data.editingPlanId) {
      return;
    }

    wx.showLoading({
      title: '更新中...',
      mask: true
    });

    api.updatePlan(this.data.editingPlanId, name, this.data.planGrade)
      .then(res => {
        wx.hideLoading();
        console.log('更新计划成功:', res);
        wx.showToast({
          title: '更新成功',
          icon: 'success',
          duration: 1500
        });
        this.hideEdit();
        this.loadPlans();
      })
      .catch(err => {
        wx.hideLoading();
        console.error('更新计划失败:', err);
        wx.showToast({
          title: err.message || '更新失败',
          icon: 'none',
          duration: 2000
        });
      });
  },

  /**
   * 管理计划中的单词（添加单词到计划）
   */
  managePlanWords(e) {
    const planId = e.currentTarget.dataset.planId;
    // 跳转到计划单词管理页面
    wx.navigateTo({
      url: `/pages/plan-words/plan-words?planId=${planId}`
    });
  },

  /**
   * 查看计划中的单词
   */
  viewPlanWords(e) {
    const planId = e.currentTarget.dataset.planId;
    // 跳转到计划单词管理页面（查看模式）
    wx.navigateTo({
      url: `/pages/plan-words/plan-words?planId=${planId}`
    });
  }
});

