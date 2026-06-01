App({
  globalData: {
    baseUrl: 'http://127.0.0.1:8080',
    token: '',
    user: null
  },
  onLaunch() {
    const token = wx.getStorageSync('token')
    const user = wx.getStorageSync('user')
    if (token) {
      this.globalData.token = token
      this.globalData.user = user
      return
    }
  }
})
