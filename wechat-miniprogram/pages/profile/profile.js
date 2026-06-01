const { request } = require('../../utils/request')

Page({
  data: { profile: {}, reminds: [] },
  onShow() {
    this.loadProfile()
    this.loadReminds()
  },
  loadProfile() {
    request({ url: '/wx/profile' }).then((res) => {
      if (res.code === 0) this.setData({ profile: res.data || {} })
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
    request({ url: '/wx/profile', method: 'PUT', data: this.data.profile }).then((res) => {
      wx.showToast({ title: res.msg || '已保存', icon: 'none' })
    })
  },
  deleteRemind(e) {
    request({ url: '/wx/groupon/reminds/' + e.currentTarget.dataset.id, method: 'DELETE' }).then(() => this.loadReminds())
  }
})
