// pages/settings/settings.js
const api = require('../../utils/api.js');
const auth = require('../../utils/auth.js');

Page({
  data: {
    studentId: '',
    studentName: '',
    studentNo: '',
    apiUrl: '',
    loading: false,
    isAutoLogin: false
  },

  onLoad() {
    const studentId = api.getStudentId();
    const apiUrl = wx.getStorageSync('apiUrl') || api.API_BASE_URL;
    
    if (studentId) {
      this.setData({
        studentId: studentId,
        apiUrl: apiUrl
      });
      this.loadStudentInfo();
    } else {
      this.setData({
        apiUrl: apiUrl
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
   * 输入学生ID
   */
  onStudentIdInput(e) {
    this.setData({
      studentId: e.detail.value
    });
  },

  /**
   * 输入API地址
   */
  onApiUrlInput(e) {
    this.setData({
      apiUrl: e.detail.value
    });
  },

  /**
   * 保存学生ID
   */
  saveStudentId() {
    const studentId = this.data.studentId.trim();
    if (!studentId) {
      wx.showToast({
        title: '请输入学生ID',
        icon: 'none'
      });
      return;
    }

    // 验证学生是否存在
    this.setData({ loading: true });
    api.getStudent(studentId)
      .then(res => {
        api.setStudentId(studentId);
        this.setData({
          studentName: res.student.name,
          studentNo: res.student.student_no
        });
        wx.showToast({
          title: '保存成功',
          icon: 'success'
        });
      })
      .catch(err => {
        wx.showToast({
          title: err.message || '学生不存在',
          icon: 'none'
        });
        console.error(err);
      })
      .finally(() => {
        this.setData({ loading: false });
      });
  },

  /**
   * 保存API地址
   */
  saveApiUrl() {
    const apiUrl = this.data.apiUrl.trim();
    if (!apiUrl) {
      wx.showToast({
        title: '请输入API地址',
        icon: 'none'
      });
      return;
    }

    wx.setStorageSync('apiUrl', apiUrl);
    wx.showToast({
      title: '保存成功',
      icon: 'success'
    });
    
    // 提示需要重启应用
    setTimeout(() => {
      wx.showModal({
        title: '提示',
        content: 'API地址已更新，请重启小程序生效',
        showCancel: false
      });
    }, 1000);
  }
});
