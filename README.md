# 知香团购（zhixiang-group-buying）

基于 Go 的团购电商后端：一套代码提供**管理后台**（服务端渲染页面 + REST API）和**微信小程序 API** 两个端，覆盖商品、订单、优惠券、库存、门店、RBAC 权限和操作审计的完整链路。

## 技术栈

- Web 框架：Gin
- ORM：GORM（MySQL，自动迁移建表）
- 缓存 / 队列：Redis
- 认证：JWT + bcrypt 密码哈希 + 图形验证码
- 前端模板：Go html/template（管理后台页面由后端直接渲染）

## 功能总览

### 管理后台（/admin）

- 登录 / 注册 / 图形验证码，JWT 鉴权的完整 CRUD：
  - 管理员、用户（封禁开关）、角色、菜单、权限
  - 商品、商品分类、自提门店
  - 订单（状态流转 + 明细）、提货码核验
  - 优惠券、库存流水、广告轮播、公告、消息
  - 团长销量统计、操作审计日志
- 仪表盘页 + 14 个后台页面（templates/ 下服务端渲染）
- `/mock-wx`：模拟微信端调试页

### 微信小程序 API（/wx）

小程序端消费的只读接口：登录、商品列表/详情、分类、门店、广告、公告、优惠券领取等。

## 快速开始

### 1. 准备环境

- Go 1.20+
- MySQL 8（建库 `zhixiang_group_buying`）
- Redis

### 2. 修改连接配置

编辑 `config/config.go`，改成你本地的 MySQL / Redis 参数和 JWT 密钥（默认值见文件内注释，**部署时务必修改默认密码和密钥**）。

### 3. 启动

```bash
go run main.go
```

启动时自动迁移建表（Admin/User/Commodity/Order 等 20+ 张表），服务监听 `:8080`。

### 4. 初始化数据

```bash
# 导入内置商品数据（含 static/products/ 下的商品图）
go run scripts/import_products.go
```

浏览器打开 `http://localhost:8080/admin/login`，先注册管理员，登录后即可使用后台。

## 项目结构

```
├── main.go               # 入口：连接 MySQL/Redis、自动迁移、注册路由
├── config/               # 数据库 / Redis / JWT 配置
├── handlers/             # 全部业务处理器（管理端 + 小程序端）
├── middleware/           # JWT 鉴权中间件
├── models/               # 20+ 张表的模型定义
├── templates/            # 管理后台 HTML 模板
├── static/products/      # 内置商品图片
└── scripts/
    ├── import_products.go  # 商品数据导入
    └── full_check.ps1      # 本地全量检查脚本
```

## 说明

- 本项目用于学习 Go Web 开发全流程（路由 / ORM / 鉴权 / RBAC / 审计 / 缓存），不是生产可用系统。
- `config/config.go` 中的默认数据库密码、JWT 密钥仅为本地开发方便，任何对外部署前必须全部更换。
