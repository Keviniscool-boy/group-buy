const { request } = require('../../utils/request')

Page({
  data: {
    ads: [],
    notice: null,
    categories: [],
    currentCat: '',
    keyword: '',
    hotGoods: [],
    goods: []
  },
  onShow() {
    this.loadAll()
  },
  onPullDownRefresh() {
    this.loadAll().finally(() => wx.stopPullDownRefresh())
  },
  loadAll() {
    return Promise.all([
      this.loadAds(),
      this.loadNotices(),
      this.loadCategories(),
      this.loadHotGoods(),
      this.loadGoods()
    ])
  },
  loadAds() {
    return request({ url: '/wx/ads' }).then((res) => {
      if (res.code === 0) this.setData({ ads: res.data || [] })
    })
  },
  loadNotices() {
    return request({ url: '/wx/announcements' }).then((res) => {
      if (res.code === 0) this.setData({ notice: (res.data || [])[0] || null })
    })
  },
  loadCategories() {
    return request({ url: '/wx/categories' }).then((res) => {
      if (res.code === 0) this.setData({ categories: res.data || [] })
    })
  },
  loadHotGoods() {
    return request({ url: '/wx/commodities?is_groupon=1' }).then((res) => {
      if (res.code === 0) this.setData({ hotGoods: res.data || [] })
    })
  },
  loadGoods() {
    let url = '/wx/commodities'
    const params = []
    if (this.data.currentCat) params.push('category_id=' + this.data.currentCat)
    if (this.data.keyword) params.push('keyword=' + encodeURIComponent(this.data.keyword))
    if (params.length) url += '?' + params.join('&')
    return request({ url }).then((res) => {
      if (res.code === 0) this.setData({ goods: res.data || [] })
    })
  },
  onKeyword(e) {
    this.setData({ keyword: e.detail.value })
  },
  searchGoods() {
    this.loadGoods()
  },
  selectCat(e) {
    this.setData({ currentCat: e.currentTarget.dataset.id || '' }, () => this.loadGoods())
  },
  goDetail(e) {
    wx.navigateTo({ url: '/pages/detail/detail?id=' + e.currentTarget.dataset.id })
  }
})
