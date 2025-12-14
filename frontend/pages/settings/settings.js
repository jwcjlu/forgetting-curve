// pages/settings/settings.js
const api = require('../../utils/api.js');
const auth = require('../../utils/auth.js');

Page({
  data: {
    studentId: '',
    studentName: '',
    studentNo: '',
    loading: false,
    isAutoLogin: false,
    grade: '',
    gradeIndex: 0,
    gradeOptions: [
      { value: '小学一年级', label: '小学一年级' },
      { value: '小学二年级', label: '小学二年级' },
      { value: '小学三年级', label: '小学三年级' },
      { value: '小学四年级', label: '小学四年级' },
      { value: '小学五年级', label: '小学五年级' },
      { value: '小学六年级', label: '小学六年级' },
      { value: '初中一年级', label: '初中一年级' },
      { value: '初中二年级', label: '初中二年级' },
      { value: '初中三年级', label: '初中三年级' },
      { value: '高中一年级', label: '高中一年级' },
      { value: '高中二年级', label: '高中二年级' },
      { value: '高中三年级', label: '高中三年级' },
      { value: '大学', label: '大学' },
      { value: '其他', label: '其他' }
    ]
  },

  onLoad() {
    const studentId = api.getStudentId();
    const grade = wx.getStorageSync('grade') || '';
    
    // 计算年级索引
    const gradeIndex = this.data.gradeOptions.findIndex(opt => opt.value === grade);
    
    if (studentId) {
      this.setData({
        studentId: studentId,
        grade: grade,
        gradeIndex: gradeIndex >= 0 ? gradeIndex : 0
      });
      this.loadStudentInfo();
    } else {
      this.setData({
        grade: grade,
        gradeIndex: gradeIndex >= 0 ? gradeIndex : 0
      });
      // 尝试自动登录
      this.tryAutoLogin();
    }
  },

  /**
   * 尝试自动登录
   */
  tryAutoLogin() {
    this.setData({ loading: true, isAutoLogin: true });
    auth.autoLogin()
      .then(result => {
        this.setData({
          studentId: result.student.id,
          studentName: result.student.name,
          studentNo: result.student.student_no
        });
        if (result.isNew) {
          wx.showToast({
            title: '注册成功',
            icon: 'success'
          });
        }
      })
      .catch(err => {
        console.error('自动登录失败:', err);
        wx.showToast({
          title: '自动登录失败',
          icon: 'none'
        });
      })
      .finally(() => {
        this.setData({ loading: false, isAutoLogin: false });
      });
  },

  /**
   * 加载学生信息
   */
  loadStudentInfo() {
    const studentId = this.data.studentId;
    if (!studentId) {
      return;
    }

    this.setData({ loading: true });
    api.getStudent(studentId)
      .then(res => {
        if (res.student) {
          this.setData({
            studentName: res.student.name,
            studentNo: res.student.student_no
          });
        }
      })
      .catch(err => {
        console.error('加载失败:', err);
        wx.showToast({
          title: err.message || '加载失败',
          icon: 'none'
        });
      })
      .finally(() => {
        this.setData({ loading: false });
      });
  },

  /**
   * 输入学生姓名
   */
  onStudentNameInput(e) {
    this.setData({
      studentName: e.detail.value
    });
  },

  /**
   * 保存学生姓名
   */
  saveStudentName() {
    const studentName = this.data.studentName.trim();
    if (!studentName) {
      wx.showToast({
        title: '请输入姓名',
        icon: 'none'
      });
      return;
    }

    // 通过重新登录来更新姓名（使用现有的openid）
    this.setData({ loading: true });
    
    // 获取本地存储的openid
    const openid = api.getOpenid();
    if (!openid) {
      wx.showToast({
        title: '请先登录',
        icon: 'none'
      });
      this.setData({ loading: false });
      return;
    }

    // 调用登录接口更新姓名
    api.getOrCreateStudentByOpenid(openid, null, studentName)
      .then(res => {
        if (res && res.student) {
          this.setData({
            studentName: res.student.name,
            studentNo: res.student.student_no
          });
          wx.showToast({
            title: '保存成功',
            icon: 'success'
          });
        } else {
          throw new Error('更新失败');
        }
      })
      .catch(err => {
        wx.showToast({
          title: err.message || '保存失败',
          icon: 'none'
        });
        console.error(err);
      })
      .finally(() => {
        this.setData({ loading: false });
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
   * 选择年级
   */
  onGradeChange(e) {
    const index = parseInt(e.detail.value);
    const grade = this.data.gradeOptions[index].value;
    this.setData({
      grade: grade,
      gradeIndex: index
    });
    wx.setStorageSync('grade', grade);
    wx.showToast({
      title: '年级已保存',
      icon: 'success',
      duration: 1500
    });
  }
});
