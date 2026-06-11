const { request } = require('../../utils/request')

Page({
  data: { items: [], stores: [], storeIndex: -1, storeId: 0, storeName: '', total: '0.00' },
  onShow() {
    this.loadCart()
    this.loadStores()
  },
  loadCart() {
    request({ url: '/wx/cart' }).then((res) => {
      const items = res.code === 0 ? (res.data || []) : []
      const total = items.filter((x) => x.checked === 1).reduce((sum, x) => sum + x.price * x.quantity, 0)
      this.setData({ items, total: total.toFixed(2) })
    })
  },
  loadStores() {
    request({ url: '/wx/stores' }).then((res) => {
      if (res.code === 0) {
        const stores = res.data || []
        let storeIndex = this.data.storeIndex
        if (this.data.storeId) {
          storeIndex = stores.findIndex((store) => store.id === this.data.storeId)
        }
        this.setData({
          stores,
          storeIndex,
          storeName: storeIndex >= 0 && stores[storeIndex] ? stores[storeIndex].name : ''
        })
      }
    })
  },
  selectStore(e) {
    const index = Number(e.detail.value)
    const store = this.data.stores[index]
    this.setData({ storeIndex: index, storeId: store.id, storeName: store.name })
  },
  toggleCheck(e) {
    request({
      url: '/wx/cart/' + e.currentTarget.dataset.id,
      method: 'PUT',
      data: { checked: e.currentTarget.dataset.checked === 1 ? 0 : 1 }
    }).then(() => this.loadCart())
  },
  changeQty(e) {
    request({
      url: '/wx/cart/' + e.currentTarget.dataset.id,
      method: 'PUT',
      data: { quantity: Number(e.currentTarget.dataset.qty) }
    }).then(() => this.loadCart())
  },
  removeItem(e) {
    request({ url: '/wx/cart/' + e.currentTarget.dataset.id, method: 'DELETE' }).then(() => this.loadCart())
  },
  submitOrder() {
    if (this.data.storeIndex < 0) {
      wx.showToast({ title: '请选择自提门店', icon: 'none' })
      return
    }
    const items = this.data.items.filter((x) => x.checked === 1)
    if (!items.length) {
      wx.showToast({ title: '请选择商品', icon: 'none' })
      return
    }
    request({
      url: '/wx/orders',
      method: 'POST',
      data: { items, store_id: this.data.storeId || this.data.stores[this.data.storeIndex].id }
    }).then((res) => {
      wx.showToast({ title: res.msg || '已提交', icon: 'none' })
      if (res.code === 0) wx.switchTab({ url: '/pages/orders/orders' })
    })
  }
})
