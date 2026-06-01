const { request } = require('../../utils/request')

Page({
  data: {
    orders: [],
    statusMap: { 0: '待付款', 1: '已付款', 2: '待取货', 3: '已取货', 4: '已取消' }
  },
  onShow() {
    this.loadOrders()
  },
  loadOrders() {
    request({ url: '/wx/orders' }).then((res) => {
      if (res.code === 0) this.setData({ orders: res.data || [] })
    })
  },
  confirmOrder(e) {
    request({ url: '/wx/orders/' + e.currentTarget.dataset.id + '/confirm', method: 'PUT' }).then((res) => {
      wx.showToast({ title: res.msg || '已确认', icon: 'none' })
      this.loadOrders()
    })
  }
})
