const { request } = require('../../utils/request')

Page({
  data: { messages: [] },
  onShow() {
    this.loadMessages()
  },
  loadMessages() {
    request({ url: '/wx/messages' }).then((res) => {
      if (res.code === 0) this.setData({ messages: res.data || [] })
    })
  },
  readMsg(e) {
    request({ url: '/wx/messages/' + e.currentTarget.dataset.id + '/read', method: 'PUT' }).then(() => this.loadMessages())
  }
})
