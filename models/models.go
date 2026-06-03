package models

import "time"

// ──────────────────────────────────────────────
// 1. cs_admin – 后台管理者表
// ──────────────────────────────────────────────
type Admin struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	RoleID    uint      `gorm:"index;default:0" json:"role_id"`
	RoleName  string    `gorm:"size:50" json:"role_name"`
	Status    int       `gorm:"default:1" json:"status"` // 1启用 0停用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Admin) TableName() string { return "cs_admin" }

// ──────────────────────────────────────────────
// 2. cs_user – 前台用户表（微信小程序端）
// ──────────────────────────────────────────────
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OpenID    string    `gorm:"uniqueIndex;size:100" json:"openid"`
	Nickname  string    `gorm:"size:50" json:"nickname"`
	Phone     string    `gorm:"size:20" json:"phone"`
	Avatar    string    `gorm:"size:255" json:"avatar"`
	Status    int       `gorm:"default:1" json:"status"` // 1正常 0禁用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (User) TableName() string { return "cs_user" }

// ──────────────────────────────────────────────
// 3. cs_role – 角色表
// ──────────────────────────────────────────────
type Role struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;size:50;not null" json:"name"`
	Desc      string    `gorm:"size:255" json:"desc"`
	Menus     string    `gorm:"type:text" json:"menus"` // 以逗号分隔的菜单ID列表
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Role) TableName() string { return "cs_role" }

// ──────────────────────────────────────────────
// 4. cs_menu – 选单功能表
// ──────────────────────────────────────────────
type Menu struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	ParentID  uint      `gorm:"default:0" json:"parent_id"`
	Path      string    `gorm:"size:100" json:"path"`
	Icon      string    `gorm:"size:50" json:"icon"`
	Sort      int       `gorm:"default:0" json:"sort"`
	CreatedAt time.Time `json:"created_at"`
}

func (Menu) TableName() string { return "cs_menu" }

// ──────────────────────────────────────────────
// 5. cs_authority – 角色功能关联表（权限表）
// ──────────────────────────────────────────────
type Authority struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	RoleID uint `gorm:"index;not null" json:"role_id"`
	MenuID uint `gorm:"index;not null" json:"menu_id"`
}

func (Authority) TableName() string { return "cs_authority" }

// ──────────────────────────────────────────────
// 6. cs_commodity – 商品表
// ──────────────────────────────────────────────
type Commodity struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	Name           string     `gorm:"size:100;not null" json:"name"`
	CategoryID     uint       `gorm:"index" json:"category_id"`
	CategoryName   string     `gorm:"size:50" json:"category_name"`
	Price          float64    `gorm:"type:decimal(10,2);not null" json:"price"`
	GroupPrice     float64    `gorm:"type:decimal(10,2);default:0" json:"group_price"`
	Stock          int        `gorm:"default:0" json:"stock"`
	Sales          int        `gorm:"default:0" json:"sales"`
	Image          string     `gorm:"size:255" json:"image"`
	Images         string     `gorm:"type:text" json:"images"`
	Description    string     `gorm:"type:text" json:"description"`
	Status         int        `gorm:"default:1" json:"status"`     // 1上架 0下架
	SaleStartTime  *time.Time `json:"sale_start_time"`             // 预售开始时间
	GroupStartTime *time.Time `json:"group_start_time"`            // 团购开始时间
	IsGroupon      int        `gorm:"default:0" json:"is_groupon"` // 0普通 1团购
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (Commodity) TableName() string { return "cs_commodity" }

// ──────────────────────────────────────────────
// 7. cs_commodity_category – 商品分类表
// ──────────────────────────────────────────────
type CommodityCategory struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	Sort      int       `gorm:"default:0" json:"sort"`
	Icon      string    `gorm:"size:255" json:"icon"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (CommodityCategory) TableName() string { return "cs_commodity_category" }

// ──────────────────────────────────────────────
// 8. cs_shoppingcart – 购物车表
// ──────────────────────────────────────────────
type ShoppingCart struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ShoppingCart) TableName() string { return "cs_shoppingcart" }

// ──────────────────────────────────────────────
// 9. cs_cartitem – 购物车项表
// ──────────────────────────────────────────────
type CartItem struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	CartID        uint      `gorm:"index;not null" json:"cart_id"`
	CommodityID   uint      `gorm:"index;not null" json:"commodity_id"`
	CommodityName string    `gorm:"size:100" json:"commodity_name"`
	Price         float64   `gorm:"type:decimal(10,2)" json:"price"`
	Quantity      int       `gorm:"default:1" json:"quantity"`
	Image         string    `gorm:"size:255" json:"image"`
	Checked       int       `gorm:"default:1" json:"checked"` // 1勾选 0未勾选
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (CartItem) TableName() string { return "cs_cartitem" }

// ──────────────────────────────────────────────
// 10. cs_order – 订单表
// ──────────────────────────────────────────────
type Order struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	OrderNo        string     `gorm:"uniqueIndex;size:32;not null" json:"order_no"`
	UserID         uint       `gorm:"index;not null" json:"user_id"`
	UserName       string     `gorm:"size:50" json:"user_name"`
	TotalAmount    float64    `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	DiscountAmount float64    `gorm:"type:decimal(10,2);default:0" json:"discount_amount"`
	PayAmount      float64    `gorm:"type:decimal(10,2);default:0" json:"pay_amount"`
	CouponID       uint       `gorm:"default:0" json:"coupon_id"`
	CouponName     string     `gorm:"size:100" json:"coupon_name"`
	Status         int        `gorm:"default:0" json:"status"` // 0待付款 1已付款 2待取货 3已取货 4已取消
	StoreID        uint       `gorm:"default:0" json:"store_id"`
	StoreName      string     `gorm:"size:100" json:"store_name"`
	PickupCode     string     `gorm:"size:10" json:"pickup_code"`
	VerifyBy       string     `gorm:"size:50" json:"verify_by"`
	VerifyTime     *time.Time `json:"verify_time"`
	Remark         string     `gorm:"size:255" json:"remark"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func (Order) TableName() string { return "cs_order" }

// ──────────────────────────────────────────────
// 11. cs_order_item – 订单项表
// ──────────────────────────────────────────────
type OrderItem struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	OrderID       uint      `gorm:"index;not null" json:"order_id"`
	CommodityID   uint      `json:"commodity_id"`
	CommodityName string    `gorm:"size:100" json:"commodity_name"`
	Price         float64   `gorm:"type:decimal(10,2)" json:"price"`
	Quantity      int       `json:"quantity"`
	Image         string    `gorm:"size:255" json:"image"`
	CreatedAt     time.Time `json:"created_at"`
}

func (OrderItem) TableName() string { return "cs_order_item" }

// ──────────────────────────────────────────────
// 12. cs_store – 自提门店表
// ──────────────────────────────────────────────
type Store struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	Name       string    `gorm:"size:100;not null" json:"name"`
	Address    string    `gorm:"size:255" json:"address"`
	Phone      string    `gorm:"size:20" json:"phone"`
	LeaderID   uint      `gorm:"default:0" json:"leader_id"`
	LeaderName string    `gorm:"size:50" json:"leader_name"`
	Status     int       `gorm:"default:1" json:"status"` // 1营业 0休息
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (Store) TableName() string { return "cs_store" }

// ──────────────────────────────────────────────
// 13. cm_groupon_remind – 开团提醒表
// ──────────────────────────────────────────────
type GrouponRemind struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"index;not null" json:"user_id"`
	CommodityID   uint      `gorm:"index;not null" json:"commodity_id"`
	CommodityName string    `gorm:"size:100" json:"commodity_name"`
	Notified      int       `gorm:"default:0" json:"notified"` // 0未通知 1已通知
	CreatedAt     time.Time `json:"created_at"`
}

func (GrouponRemind) TableName() string { return "cm_groupon_remind" }

// ──────────────────────────────────────────────
// 14. cs_message – 站内消息表
// ──────────────────────────────────────────────
type Message struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index" json:"user_id"` // 0表示全体用户
	Title     string    `gorm:"size:100;not null" json:"title"`
	Content   string    `gorm:"type:text" json:"content"`
	IsRead    int       `gorm:"default:0" json:"is_read"`
	Type      int       `gorm:"default:0" json:"type"` // 0系统 1订单 2团购
	CreatedAt time.Time `json:"created_at"`
}

func (Message) TableName() string { return "cs_message" }

// ──────────────────────────────────────────────
// 15. cs_ads – 广告轮播图表
// ──────────────────────────────────────────────
type Ads struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"size:100" json:"title"`
	Image     string    `gorm:"size:255;not null" json:"image"`
	Link      string    `gorm:"size:255" json:"link"`
	Sort      int       `gorm:"default:0" json:"sort"`
	Status    int       `gorm:"default:1" json:"status"` // 1显示 0隐藏
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Ads) TableName() string { return "cs_ads" }

// ──────────────────────────────────────────────
// 16. cs_announcement – 公告表
// ──────────────────────────────────────────────
type Announcement struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"size:100;not null" json:"title"`
	Content   string    `gorm:"type:text" json:"content"`
	Status    int       `gorm:"default:1" json:"status"` // 1发布 0隐藏
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Announcement) TableName() string { return "cs_announcement" }

// ──────────────────────────────────────────────
// 17. cs_operation_log – 后台操作审计日志
// ──────────────────────────────────────────────
type OperationLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	AdminID   uint      `gorm:"index" json:"admin_id"`
	Username  string    `gorm:"size:50" json:"username"`
	RoleName  string    `gorm:"size:50" json:"role_name"`
	Method    string    `gorm:"size:10" json:"method"`
	Path      string    `gorm:"size:255" json:"path"`
	Action    string    `gorm:"size:50" json:"action"`
	IP        string    `gorm:"size:64" json:"ip"`
	UserAgent string    `gorm:"size:255" json:"user_agent"`
	Status    int       `json:"status"`
	Duration  int64     `json:"duration"`
	CreatedAt time.Time `json:"created_at"`
}

func (OperationLog) TableName() string { return "cs_operation_log" }

// Coupon stores platform promotion rules.
type Coupon struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	Name      string     `gorm:"size:100;not null" json:"name"`
	Threshold float64    `gorm:"type:decimal(10,2);default:0" json:"threshold"`
	Amount    float64    `gorm:"type:decimal(10,2);not null" json:"amount"`
	Total     int        `gorm:"default:0" json:"total"`
	Received  int        `gorm:"default:0" json:"received"`
	Used      int        `gorm:"default:0" json:"used"`
	Status    int        `gorm:"default:1" json:"status"`
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (Coupon) TableName() string { return "cs_coupon" }

// UserCoupon is one claimed coupon owned by a user.
type UserCoupon struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `gorm:"index;not null" json:"user_id"`
	CouponID   uint       `gorm:"index;not null" json:"coupon_id"`
	CouponName string     `gorm:"size:100" json:"coupon_name"`
	Threshold  float64    `gorm:"type:decimal(10,2)" json:"threshold"`
	Amount     float64    `gorm:"type:decimal(10,2)" json:"amount"`
	Status     int        `gorm:"default:0" json:"status"`
	OrderID    uint       `gorm:"default:0" json:"order_id"`
	CreatedAt  time.Time  `json:"created_at"`
	UsedAt     *time.Time `json:"used_at"`
}

func (UserCoupon) TableName() string { return "cs_user_coupon" }

// StockLog records every inventory movement.
type StockLog struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	CommodityID   uint      `gorm:"index;not null" json:"commodity_id"`
	CommodityName string    `gorm:"size:100" json:"commodity_name"`
	ChangeQty     int       `json:"change_qty"`
	BeforeQty     int       `json:"before_qty"`
	AfterQty      int       `json:"after_qty"`
	Type          string    `gorm:"size:30" json:"type"`
	RefID         uint      `gorm:"default:0" json:"ref_id"`
	Remark        string    `gorm:"size:255" json:"remark"`
	Operator      string    `gorm:"size:50" json:"operator"`
	CreatedAt     time.Time `json:"created_at"`
}

func (StockLog) TableName() string { return "cs_stock_log" }

// ──────────────────────────────────────────────
// 辅助结构
// ──────────────────────────────────────────────

// LoginRequest 登入请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Captcha  string `json:"captcha"`
}

// WxLoginRequest 微信登入请求
type WxLoginRequest struct {
	Code     string `json:"code" binding:"required"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// PageResult 分页结果
type PageResult struct {
	Code  int         `json:"code"`
	Msg   string      `json:"msg"`
	Count int64       `json:"count"`
	Data  interface{} `json:"data"`
}

// Result 通用回传
type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// LoginResult 登入回传
type LoginResult struct {
	Token string `json:"token"`
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}
