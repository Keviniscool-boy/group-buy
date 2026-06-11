package main

import (
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"zhixiang-group-buying/config"
	"zhixiang-group-buying/handlers"
	"zhixiang-group-buying/middleware"
	"zhixiang-group-buying/models"
)

func main() {
	// ── 连线资料库 ──────────────────────────────
	dsn := config.GetDSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("资料库连线失败: %v", err)
	}
	log.Println("资料库连线成功")

	// ── 连线 Redis ──────────────────────────────
	config.InitRedis()

	// ── 自动迁移（建表） ─────────────────────────
	err = db.AutoMigrate(
		&models.Admin{},
		&models.User{},
		&models.Role{},
		&models.Menu{},
		&models.Authority{},
		&models.Commodity{},
		&models.CommodityCategory{},
		&models.ShoppingCart{},
		&models.CartItem{},
		&models.Order{},
		&models.OrderItem{},
		&models.Store{},
		&models.GrouponRemind{},
		&models.Message{},
		&models.Ads{},
		&models.Announcement{},
		&models.OperationLog{},
		&models.Coupon{},
		&models.UserCoupon{},
		&models.StockLog{},
	)
	if err != nil {
		log.Fatalf("资料库迁移失败: %v", err)
	}
	log.Println("资料库表迁移完成")

	// ── 初始化种子数据 ──────────────────────────
	seedData(db)

	// ── 建立 Gin 引擎 ──────────────────────────
	r := gin.Default()

	// CORS 跨域设定
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// ── 静态档案与模板 ──────────────────────────
	r.Static("/static", "./static")
	r.LoadHTMLGlob("templates/*")

	// ── 页面路由 ──────────────────────────────
	r.GET("/admin/login", func(c *gin.Context) {
		c.HTML(200, "admin_login.html", nil)
	})
	r.GET("/admin", func(c *gin.Context) {
		c.HTML(200, "admin_dashboard.html", nil)
	})
	r.GET("/admin/commodity", func(c *gin.Context) {
		c.HTML(200, "admin_dashboard.html", nil)
	})
	r.GET("/admin/category", func(c *gin.Context) {
		c.HTML(200, "admin_dashboard.html", nil)
	})
	r.GET("/admin/order", func(c *gin.Context) {
		c.HTML(200, "admin_dashboard.html", nil)
	})
	r.GET("/admin/user", func(c *gin.Context) {
		c.HTML(200, "admin_dashboard.html", nil)
	})
	r.GET("/admin/admin-manager", func(c *gin.Context) {
		c.HTML(200, "admin_dashboard.html", nil)
	})
	r.GET("/admin/role", func(c *gin.Context) {
		c.HTML(200, "admin_dashboard.html", nil)
	})
	r.GET("/admin/store", func(c *gin.Context) {
		c.HTML(200, "admin_dashboard.html", nil)
	})
	r.GET("/admin/announcement", func(c *gin.Context) {
		c.HTML(200, "admin_dashboard.html", nil)
	})
	r.GET("/admin/ads", func(c *gin.Context) {
		c.HTML(200, "admin_dashboard.html", nil)
	})
	r.GET("/admin/message", func(c *gin.Context) {
		c.HTML(200, "admin_dashboard.html", nil)
	})
	r.GET("/admin/operation-log", func(c *gin.Context) {
		c.HTML(200, "admin_dashboard.html", nil)
	})
	r.GET("/admin/coupon", func(c *gin.Context) {
		c.HTML(200, "admin_dashboard.html", nil)
	})
	r.GET("/admin/stock-log", func(c *gin.Context) {
		c.HTML(200, "admin_dashboard.html", nil)
	})
	r.GET("/admin/verify", func(c *gin.Context) {
		c.HTML(200, "admin_dashboard.html", nil)
	})
	r.GET("/mock-wx", func(c *gin.Context) {
		c.HTML(200, "mock_wx.html", nil)
	})

	// ═══════════════════════════════════════════
	//  后台管理 API (/admin)
	// ═══════════════════════════════════════════

	admin := r.Group("/admin")
	{
		// 公开路由
		admin.POST("/login", handlers.AdminLogin(db))
		admin.POST("/register", handlers.AdminRegister(db))
		admin.GET("/captcha", handlers.GetCaptcha(db))

		// 需 JWT 认证
		auth := admin.Group("")
		auth.Use(middleware.AdminAuth())
		auth.Use(handlers.AdminAuditLog(db))
		{
			// 仪表板
			auth.GET("/dashboard", handlers.AdminDashboard(db))

			// 管理员管理
			auth.GET("/admins", handlers.ListAdmins(db))
			auth.POST("/admins", handlers.CreateAdmin(db))
			auth.PUT("/admins/:id", handlers.UpdateAdmin(db))
			auth.DELETE("/admins/:id", handlers.DeleteAdmin(db))

			// 用户管理
			auth.GET("/users", handlers.ListUsers(db))
			auth.PUT("/users/:id/toggle", handlers.ToggleUser(db))
			auth.DELETE("/users/:id", handlers.DeleteUser(db))

			// 商品分类
			auth.GET("/categories", handlers.ListCategories(db))
			auth.POST("/categories", handlers.CreateCategory(db))
			auth.PUT("/categories/:id", handlers.UpdateCategory(db))
			auth.DELETE("/categories/:id", handlers.DeleteCategory(db))

			// 商品管理
			auth.GET("/commodities", handlers.AdminListCommodities(db))
			auth.POST("/commodities", handlers.CreateCommodity(db))
			auth.PUT("/commodities/:id", handlers.UpdateCommodity(db))
			auth.PUT("/commodities/:id/toggle", handlers.ToggleCommodity(db))
			auth.DELETE("/commodities/:id", handlers.DeleteCommodity(db))

			// 订单管理
			auth.GET("/orders", handlers.AdminListOrders(db))
			auth.GET("/orders/:id", handlers.AdminGetOrder(db))
			auth.GET("/orders/:id/items", handlers.AdminListOrderItems(db))
			auth.PUT("/orders/:id/status", handlers.AdminUpdateOrderStatus(db))
			auth.POST("/orders/verify", handlers.AdminVerifyPickupCode(db))

			// 自提门店
			auth.GET("/stores", handlers.ListStores(db))
			auth.POST("/stores", handlers.CreateStore(db))
			auth.PUT("/stores/:id", handlers.UpdateStore(db))
			auth.DELETE("/stores/:id", handlers.DeleteStore(db))

			// 消息
			auth.POST("/messages", handlers.CreateMessage(db))

			// 广告轮播图
			auth.GET("/ads/all", handlers.AdminListAds(db))
			auth.POST("/ads", handlers.CreateAds(db))
			auth.PUT("/ads/:id", handlers.UpdateAds(db))
			auth.DELETE("/ads/:id", handlers.DeleteAds(db))

			// 公告
			auth.GET("/announcements/all", handlers.AdminListAnnouncements(db))
			auth.POST("/announcements", handlers.CreateAnnouncement(db))
			auth.PUT("/announcements/:id", handlers.UpdateAnnouncement(db))
			auth.DELETE("/announcements/:id", handlers.DeleteAnnouncement(db))

			// 角色管理
			auth.GET("/roles", handlers.ListRoles(db))
			auth.POST("/roles", handlers.CreateRole(db))
			auth.PUT("/roles/:id", handlers.UpdateRole(db))
			auth.DELETE("/roles/:id", handlers.DeleteRole(db))

			// 菜单管理
			auth.GET("/menus", handlers.ListMenus(db))
			auth.POST("/menus", handlers.CreateMenu(db))
			auth.DELETE("/menus/:id", handlers.DeleteMenu(db))

			// 权限管理
			auth.GET("/authorities", handlers.ListAuthorities(db))
			auth.POST("/authorities", handlers.SaveAuthorities(db))

			// 团长销量
			auth.GET("/leader-sales", handlers.LeaderSales(db))

			// 操作审计日志
			auth.GET("/operation-logs", handlers.ListOperationLogs(db))

			// 营销、库存与核销
			auth.GET("/coupons", handlers.AdminListCoupons(db))
			auth.POST("/coupons", handlers.CreateCoupon(db))
			auth.PUT("/coupons/:id", handlers.UpdateCoupon(db))
			auth.DELETE("/coupons/:id", handlers.DeleteCoupon(db))
			auth.GET("/stock-logs", handlers.AdminListStockLogs(db))
		}
	}

	// ═══════════════════════════════════════════
	//  微信小程序 API (/wx)
	// ═══════════════════════════════════════════

	wx := r.Group("/wx")
	{
		// 公开路由
		wx.POST("/login", handlers.WxLogin(db))
		wx.GET("/categories", handlers.ListCategories(db))
		wx.GET("/commodities", handlers.WxListCommodities(db))
		wx.GET("/commodities/:id", handlers.WxGetCommodity(db))
		wx.GET("/stores", handlers.WxListStores(db))
		wx.GET("/ads", handlers.ListAds(db))
		wx.GET("/announcements", handlers.ListAnnouncements(db))
		wx.GET("/coupons", handlers.WxListCoupons(db))

		// 需 JWT 认证
		auth := wx.Group("")
		auth.Use(middleware.WxAuth())
		{
			// 个人中心
			auth.GET("/profile", handlers.WxGetProfile(db))
			auth.PUT("/profile", handlers.WxUpdateProfile(db))

			// 购物车
			auth.POST("/cart", handlers.WxAddToCart(db))
			auth.GET("/cart", handlers.WxListCart(db))
			auth.PUT("/cart/:id", handlers.WxUpdateCartItem(db))
			auth.DELETE("/cart/:id", handlers.WxDeleteCartItem(db))
			auth.PUT("/cart/check-all", handlers.WxCheckAllCart(db))
			auth.DELETE("/cart/clear", handlers.WxClearCart(db))

			// 订单
			auth.POST("/orders", handlers.WxCreateOrder(db))
			auth.GET("/orders", handlers.WxListOrders(db))
			auth.PUT("/orders/:id/confirm", handlers.WxConfirmOrder(db))

			// 优惠券
			auth.POST("/coupons/:id/receive", handlers.WxReceiveCoupon(db))
			auth.GET("/my-coupons", handlers.WxListMyCoupons(db))

			// 消息
			auth.GET("/messages", handlers.ListMessages(db))
			auth.PUT("/messages/:id/read", handlers.ReadMessage(db))

			// 团购提醒
			auth.POST("/groupon/subscribe", handlers.WxSubscribeGroupon(db))
			auth.GET("/groupon/reminds", handlers.WxListGrouponReminds(db))
			auth.DELETE("/groupon/reminds/:id", handlers.WxDeleteGrouponRemind(db))
		}
	}

	log.Printf("服务启动于 http://localhost%s", config.ServerPort)
	r.Run(config.ServerPort)
}

// seedData 初始化种子数据
func seedData(db *gorm.DB) {
	// ── 建立预设角色 ──────────────────────────
	roles := []models.Role{
		{Name: "超级管理员", Desc: "拥有所有权限", Menus: "1,2,3,4,5,6,7,8,9"},
		{Name: "团长", Desc: "查看销量与站点订单"},
		{Name: "业务管理员", Desc: "发布与查询商品、发布通知"},
		{Name: "人事管理员", Desc: "审核供应商和团长申请、角色管理"},
	}
	for _, r := range roles {
		db.Where("name = ?", r.Name).FirstOrCreate(&r)
	}

	// ── 建立预设选单 ──────────────────────────
	menus := []models.Menu{
		{Name: "仪表板", Path: "/admin/dashboard", Icon: "layui-icon-home", Sort: 1},
		{Name: "商品管理", Path: "/admin/commodity", Icon: "layui-icon-cart-simple", Sort: 2},
		{Name: "订单管理", Path: "/admin/order", Icon: "layui-icon-form", Sort: 3},
		{Name: "用户管理", Path: "/admin/user", Icon: "layui-icon-user", Sort: 4},
		{Name: "角色管理", Path: "/admin/role", Icon: "layui-icon-username", Sort: 5},
		{Name: "门店管理", Path: "/admin/store", Icon: "layui-icon-location", Sort: 6},
		{Name: "广告管理", Path: "/admin/ads", Icon: "layui-icon-picture", Sort: 7},
		{Name: "公告管理", Path: "/admin/announcement", Icon: "layui-icon-notice", Sort: 8},
		{Name: "团长销量", Path: "/admin/leader-sales", Icon: "layui-icon-chart", Sort: 9},
	}
	for _, m := range menus {
		db.Where("name = ?", m.Name).FirstOrCreate(&m)
	}

	// ── 建立预设管理员 ──────────────────────────
	var adminCount int64
	db.Model(&models.Admin{}).Count(&adminCount)
	if adminCount == 0 {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		db.Create(&models.Admin{
			Username: "admin",
			Password: string(hashed),
			RoleName: "超级管理员",
			Status:   1,
		})

	}

	// ── 建立预设分类 ──────────────────────────
	categories := []models.CommodityCategory{
		{Name: "新鲜水果", Sort: 1},
		{Name: "时令蔬菜", Sort: 2},
		{Name: "肉禽蛋奶", Sort: 3},
		{Name: "海鲜水产", Sort: 4},
		{Name: "粮油调味", Sort: 5},
		{Name: "休闲零食", Sort: 6},
	}
	for _, c := range categories {
		db.Where("name = ?", c.Name).FirstOrCreate(&c)
	}

	// ── 建立预设门店 ──────────────────────────
	var storeCount int64
	db.Model(&models.Store{}).Count(&storeCount)
	if storeCount == 0 {
		stores := []models.Store{
			{Name: "阳光花园自提点", Address: "阳光花园小区1号楼底商", Phone: "13800001001", Status: 1},
			{Name: "翠苑新邨自提点", Address: "翠苑新邨东门旁", Phone: "13800001002", Status: 1},
			{Name: "万达广场自提点", Address: "万达广场B1层超市入口", Phone: "13800001003", Status: 1},
		}
		for _, s := range stores {
			db.Create(&s)
		}
	}

	// ── 建立预设商品 ──────────────────────────
	var commodityCount int64
	db.Model(&models.Commodity{}).Count(&commodityCount)
	if commodityCount == 0 {
		var cats []models.CommodityCategory
		db.Find(&cats)
		catMap := map[string]uint{}
		for _, c := range cats {
			catMap[c.Name] = c.ID
		}

		now := time.Now()
		commodities := []models.Commodity{
			{
				Name:           "福建平和蜜柚",
				CategoryID:     catMap["新鲜水果"],
				CategoryName:   "新鲜水果",
				Price:          19.90,
				GroupPrice:     16.90,
				Stock:          200,
				Image:          "https://picsum.photos/seed/pomelo/400/400",
				Description:    "清甜多汁，适合家庭分享。",
				IsGroupon:      1,
				GroupStartTime: &now,
			},
			{
				Name:          "阳光草莓 500g",
				CategoryID:    catMap["新鲜水果"],
				CategoryName:  "新鲜水果",
				Price:         29.90,
				Stock:         120,
				Image:         "https://picsum.photos/seed/strawberry/400/400",
				Description:   "香甜多汁，冷藏口感更佳。",
				SaleStartTime: &now,
			},
			{
				Name:         "有机青菜 300g",
				CategoryID:   catMap["时令蔬菜"],
				CategoryName: "时令蔬菜",
				Price:        6.80,
				Stock:        300,
				Image:        "https://picsum.photos/seed/greens/400/400",
				Description:  "当日采摘，新鲜直达。",
			},
			{
				Name:         "散养土鸡蛋 10枚",
				CategoryID:   catMap["肉禽蛋奶"],
				CategoryName: "肉禽蛋奶",
				Price:        19.80,
				Stock:        160,
				Image:        "https://picsum.photos/seed/eggs/400/400",
				Description:  "蛋黄饱满，口感更香。",
			},
			{
				Name:         "冷冻对虾 500g",
				CategoryID:   catMap["海鲜水产"],
				CategoryName: "海鲜水产",
				Price:        36.90,
				Stock:        80,
				Image:        "https://picsum.photos/seed/shrimp/400/400",
				Description:  "肉质紧实，适合家常烹饪。",
			},
			{
				Name:         "橄榄油 500ml",
				CategoryID:   catMap["粮油调味"],
				CategoryName: "粮油调味",
				Price:        49.90,
				Stock:        60,
				Image:        "https://picsum.photos/seed/oil/400/400",
				Description:  "清香不腻，适合凉拌与轻炒。",
			},
			{
				Name:         "坚果礼包 350g",
				CategoryID:   catMap["休闲零食"],
				CategoryName: "休闲零食",
				Price:        25.90,
				Stock:        140,
				Image:        "https://picsum.photos/seed/nuts/400/400",
				Description:  "多种坚果搭配，营养均衡。",
			},
		}
		for _, c := range commodities {
			if c.CategoryID > 0 {
				db.Create(&c)
			}
		}
	}

	// ── 建立预设广告 ──────────────────────────
	var adsCount int64
	db.Model(&models.Ads{}).Count(&adsCount)
	if adsCount == 0 {
		ads := []models.Ads{
			{Title: "本周爆品", Image: "https://picsum.photos/seed/banner1/800/400", Link: "/mock-wx", Sort: 1, Status: 1},
			{Title: "新人专享", Image: "https://picsum.photos/seed/banner2/800/400", Link: "/mock-wx", Sort: 2, Status: 1},
		}
		for _, a := range ads {
			db.Create(&a)
		}
	}

	// ── 建立预设公告 ──────────────────────────
	var annCount int64
	db.Model(&models.Announcement{}).Count(&annCount)
	if annCount == 0 {
		anns := []models.Announcement{
			{Title: "今日下单，次日自提", Content: "请在营业时间内到门店自提。", Status: 1},
			{Title: "团购开启", Content: "团购价商品数量有限，先到先得。", Status: 1},
		}
		for _, a := range anns {
			db.Create(&a)
		}
	}

	var couponCount int64
	db.Model(&models.Coupon{}).Count(&couponCount)
	if couponCount == 0 {
		now := time.Now()
		end := now.AddDate(0, 1, 0)
		coupons := []models.Coupon{
			{Name: "新人满30减5", Threshold: 30, Amount: 5, Total: 500, Status: 1, StartTime: &now, EndTime: &end},
			{Name: "社区团购满80减12", Threshold: 80, Amount: 12, Total: 300, Status: 1, StartTime: &now, EndTime: &end},
			{Name: "生鲜专享满120减20", Threshold: 120, Amount: 20, Total: 200, Status: 1, StartTime: &now, EndTime: &end},
		}
		for _, coupon := range coupons {
			db.Create(&coupon)
		}
	}
}
