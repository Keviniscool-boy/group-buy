let loginTask = null

function login(force) {
  if (loginTask) return loginTask
  const app = getApp()
  const cachedToken = app.globalData.token || wx.getStorageSync('token')
  const cachedUser = app.globalData.user || wx.getStorageSync('user')
  if (!force && cachedToken) {
    app.globalData.token = cachedToken
    app.globalData.user = cachedUser || null
    return Promise.resolve(cachedUser)
  }

  loginTask = new Promise((resolve, reject) => {
    wx.login({
      success(loginRes) {
        const code = createLoginCode()
        doLogin(app, code, '', '', resolve, reject)
      },
      fail() {
        doLogin(app, createLoginCode(), '', '', resolve, reject)
      }
    })
  })
  return loginTask
}

function doLogin(app, code, nickname, avatar, resolve, reject) {
  wx.request({
    url: app.globalData.baseUrl + '/wx/login',
    method: 'POST',
    header: { 'content-type': 'application/json' },
    data: { code, nickname, avatar },
    success(res) {
      const data = res.data || {}
      if (data.code === 0) {
        const user = normalizeLoginUser(data.data)
        app.globalData.token = user.token
        app.globalData.user = user
        wx.setStorageSync('token', user.token)
        wx.setStorageSync('user', user)
        resolve(user)
        return
      }
      reject(data)
    },
    fail(err) {
      reject(err)
    },
    complete() {
      loginTask = null
    }
  })
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
        const data = res.data || {}
        // token 失效时自动清除并重新登录
        if (!options.noAuth && data.code === 401) {
          app.globalData.token = ''
          app.globalData.user = null
          wx.removeStorageSync('token')
          wx.removeStorageSync('user')
          login().then(() => request(options)).then(resolve).catch(reject)
          return
        }
        resolve(normalizeResponseAssets(data, baseUrl))
      },
      fail(err) {
        wx.showToast({ title: '服务连接失败', icon: 'none' })
        reject(err)
      }
    })
  })
}

function createLoginCode() {
  let code = wx.getStorageSync('mockLoginCode')
  if (!code) {
    code = 'DEV_WX_' + Date.now() + '_' + Math.floor(Math.random() * 10000)
    wx.setStorageSync('mockLoginCode', code)
  }
  return code
}

function normalizeLoginUser(data) {
  const user = data || {}
  return Object.assign({}, user, {
    token: user.token || '',
    nickname: user.nickname || user.name || '',
    name: user.name || user.nickname || '',
    avatar: user.avatar || ''
  })
}

function normalizeResponseAssets(payload, baseUrl) {
  if (Array.isArray(payload)) {
    return payload.map((item) => normalizeResponseAssets(item, baseUrl))
  }
  if (!payload || typeof payload !== 'object') return payload
  Object.keys(payload).forEach((key) => {
    const value = payload[key]
    if (value && typeof value === 'object') {
      payload[key] = normalizeResponseAssets(value, baseUrl)
      return
    }
    if (typeof value === 'string' && shouldNormalizeAsset(key, value)) {
      payload[key] = baseUrl + value
    }
  })
  return payload
}

function shouldNormalizeAsset(key, value) {
  return ['image', 'images', 'avatar'].indexOf(key) !== -1 && value.indexOf('/') === 0 && value.indexOf('//') !== 0
}

module.exports = { request, login }
