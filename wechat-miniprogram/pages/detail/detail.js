const { request } = require('../../utils/request')

Page({
  data: { id: 0, goods: null },
  onLoad(query) {
    this.setData({ id: query.id })
    this.loadGoods()
  },
  loadGoods() {
    request({ url: '/wx/commodities/' + this.data.id }).then((res) => {
      if (res.code === 0) this.setData({ goods: res.data })
    })
  },
  addCart() {
    request({
      url: '/wx/cart',
      method: 'POST',
      data: { commodity_id: Number(this.data.id), quantity: 1 }
    }).then((res) => wx.showToast({ title: res.msg || '已加入', icon: 'none' }))
  },
  buyNow() {
    this.addCart()
    setTimeout(() => wx.switchTab({ url: '/pages/cart/cart' }), 500)
  }
})
