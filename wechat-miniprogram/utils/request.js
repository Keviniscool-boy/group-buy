let loginTask = null

function login() {
  if (loginTask) return loginTask
  const app = getApp()
  loginTask = new Promise((resolve, reject) => {
    wx.request({
      url: app.globalData.baseUrl + '/wx/login',
      method: 'POST',
      header: { 'content-type': 'application/json' },
      data: {
        code: 'wx_demo_' + Date.now(),
        nickname: '社区用户' + Math.floor(Math.random() * 9000 + 1000),
        avatar: ''
      },
      success(res) {
        const data = res.data || {}
        if (data.code === 0) {
          app.globalData.token = data.data.token
          app.globalData.user = data.data
          wx.setStorageSync('token', data.data.token)
          wx.setStorageSync('user', data.data)
        }
        resolve(data)
      },
      fail(err) {
        reject(err)
      },
      complete() {
        loginTask = null
      }
    })
  })
  return loginTask
}

function request(options) {
  const app = getApp()
  const baseUrl = app.globalData.baseUrl
  const header = Object.assign({ 'content-type': 'application/json' }, options.header || {})
  if (!options.noAuth) {
    const token = app.globalData.token || wx.getStorageSync('token')
    if (token) {
      header.Authorization = 'Bearer ' + token
    } else {
      return login().then(() => request(options))
    }
  }
  return new Promise((resolve, reject) => {
    wx.request({
      url: baseUrl + options.url,
      method: options.method || 'GET',
      data: options.data || {},
      header,
      success(res) {
        resolve(res.data)
      },
      fail(err) {
        wx.showToast({ title: '服务连接失败', icon: 'none' })
        reject(err)
      }
    })
  })
}

module.exports = { request }
