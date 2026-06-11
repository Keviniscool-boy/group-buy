const { request, login } = require('../../utils/request')

Page({
  data: {
    profile: {},
    reminds: [],
    loggedIn: false,
    loggingIn: false,
    loginText: '微信一键登录',
    loginStatus: '未登录'
  },
  onShow() {
    this.refreshLoginState()
  },
  refreshLoginState() {
    const token = wx.getStorageSync('token')
    const user = wx.getStorageSync('user')
    this.setData({
      loggedIn: !!token,
      loginText: token ? '重新登录' : '微信一键登录',
      loginStatus: token ? '已登录' : '未登录',
      profile: user ? Object.assign({}, this.data.profile, user) : this.data.profile
    })
    if (token) {
      this.loadProfile()
      this.loadReminds()
    }
  },
  wxLogin() {
    if (this.data.loggingIn) return
    this.setData({ loggingIn: true, loginText: '登录中...', loginStatus: '正在登录' })
    login(true)
      .then((userInfo) => {
        this.setData({
          loggedIn: true,
          loggingIn: false,
          loginText: '重新登录',
          loginStatus: '已登录',
          profile: Object.assign({}, this.data.profile, userInfo || {})
        })
        this.loadProfile()
        this.loadReminds()
        wx.showToast({ title: '登录成功', icon: 'success' })
      })
      .catch(() => {
        this.setData({ loggingIn: false, loginText: '微信一键登录', loginStatus: '登录失败' })
        wx.showToast({ title: '登录失败，请确认后端已启动', icon: 'none' })
      })
  },
  loadProfile() {
    request({ url: '/wx/profile' }).then((res) => {
      if (res.code === 0) {
        this.setData({ profile: res.data || {} })
        wx.setStorageSync('user', Object.assign({}, wx.getStorageSync('user') || {}, res.data || {}))
      }
    })
  },
  loadReminds() {
    request({ url: '/wx/groupon/reminds' }).then((res) => {
      if (res.code === 0) this.setData({ reminds: res.data || [] })
    })
  },
  onNickname(e) {
    this.setData({ 'profile.nickname': e.detail.value })
  },
  onPhone(e) {
    this.setData({ 'profile.phone': e.detail.value })
  },
  saveProfile() {
    if (!this.data.loggedIn) {
      wx.showToast({ title: '请先微信一键登录', icon: 'none' })
      return
    }
    request({ url: '/wx/profile', method: 'PUT', data: this.data.profile }).then((res) => {
      wx.showToast({ title: res.msg || '已保存', icon: 'none' })
    })
  },
  deleteRemind(e) {
    request({ url: '/wx/groupon/reminds/' + e.currentTarget.dataset.id, method: 'DELETE' }).then(() => this.loadReminds())
  }
})
