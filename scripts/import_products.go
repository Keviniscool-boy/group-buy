package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"zhixiang-group-buying/config"
	"zhixiang-group-buying/models"
)

type productSeed struct {
	Name         string
	CategoryName string
	Price        float64
	GroupPrice   float64
	Stock        int
	ImageFile    string
	Description  string
	IsGroupon    int
}

func main() {
	db, err := gorm.Open(mysql.Open(config.GetDSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}

	if err := db.AutoMigrate(&models.CommodityCategory{}, &models.Commodity{}); err != nil {
		log.Fatalf("migrate tables: %v", err)
	}

	assetDir := filepath.Join("static", "products")
	if err := os.MkdirAll(assetDir, 0755); err != nil {
		log.Fatalf("create asset dir: %v", err)
	}

	for _, p := range products {
		dstName := stableImageName(p.ImageFile)
		dst := filepath.Join(assetDir, dstName)
		if _, err := os.Stat(dst); err != nil {
			log.Fatalf("missing image %s: %v", dst, err)
		}

		catID, err := ensureCategory(db, p.CategoryName)
		if err != nil {
			log.Fatalf("ensure category %s: %v", p.CategoryName, err)
		}

		imageURL := "/static/products/" + dstName
		commodity := models.Commodity{
			Name:         p.Name,
			CategoryID:   catID,
			CategoryName: p.CategoryName,
			Price:        p.Price,
			GroupPrice:   p.GroupPrice,
			Stock:        p.Stock,
			Image:        imageURL,
			Images:       imageURL,
			Description:  p.Description,
			Status:       1,
			IsGroupon:    p.IsGroupon,
		}
		if p.IsGroupon == 1 {
			now := time.Now()
			commodity.GroupStartTime = &now
		}

		var existing models.Commodity
		err = db.Where("name = ?", p.Name).First(&existing).Error
		if err == nil {
			commodity.ID = existing.ID
			commodity.CreatedAt = existing.CreatedAt
			if err := db.Model(&existing).Updates(commodity).Error; err != nil {
				log.Fatalf("update product %s: %v", p.Name, err)
			}
			fmt.Printf("updated: %s\n", p.Name)
			continue
		}
		if err != gorm.ErrRecordNotFound {
			log.Fatalf("find product %s: %v", p.Name, err)
		}
		if err := db.Create(&commodity).Error; err != nil {
			log.Fatalf("create product %s: %v", p.Name, err)
		}
		fmt.Printf("created: %s\n", p.Name)
	}

	config.InitRedis()
	config.CacheDelPattern(context.Background(), config.CacheKey("commodities")+"*")
	config.CacheDel(context.Background(), config.CacheKey("categories"))
}

func ensureCategory(db *gorm.DB, name string) (uint, error) {
	var cat models.CommodityCategory
	err := db.Where("name = ?", name).First(&cat).Error
	if err == nil {
		return cat.ID, nil
	}
	if err != gorm.ErrRecordNotFound {
		return 0, err
	}
	cat = models.CommodityCategory{Name: name, Sort: 99}
	if err := db.Create(&cat).Error; err != nil {
		return 0, err
	}
	return cat.ID, nil
}

func stableImageName(name string) string {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	replacer := strings.NewReplacer(" ", "-", "（", "-", "）", "-", "(", "-", ")", "-")
	base = replacer.Replace(base)
	return base + ext
}

var products = []productSeed{
	{Name: "士力架花生夹心巧克力", CategoryName: "休闲零食", Price: 4.90, GroupPrice: 3.90, Stock: 300, ImageFile: "product1.jpg", Description: "花生夹心巧克力，补充能量，适合随身携带。", IsGroupon: 1},
	{Name: "手撕烤筋辣条", CategoryName: "休闲零食", Price: 2.90, GroupPrice: 2.50, Stock: 500, ImageFile: "product2.jpg", Description: "多口味手撕素肉零食，香辣有嚼劲。", IsGroupon: 1},
	{Name: "夹心铜锣烧", CategoryName: "休闲零食", Price: 3.50, GroupPrice: 2.99, Stock: 260, ImageFile: "product3.jpg", Description: "松软饼皮搭配香甜夹心，早餐加餐都合适。", IsGroupon: 1},
	{Name: "夏威夷果小包装", CategoryName: "休闲零食", Price: 12.90, GroupPrice: 10.90, Stock: 120, ImageFile: "product4.jpg", Description: "精选坚果，口感酥香，独立包装更方便。", IsGroupon: 1},
	{Name: "绿豆糕礼盒", CategoryName: "休闲零食", Price: 18.80, GroupPrice: 16.80, Stock: 90, ImageFile: "product5.jpg", Description: "越南风味绿豆糕，清甜细腻。", IsGroupon: 1},
	{Name: "德芙巧克力小罐装", CategoryName: "休闲零食", Price: 29.90, GroupPrice: 26.90, Stock: 80, ImageFile: "product6.jpg", Description: "丝滑巧克力，小罐包装，分享方便。", IsGroupon: 1},
	{Name: "星球杯巧克力饼干", CategoryName: "休闲零食", Price: 15.90, GroupPrice: 13.90, Stock: 150, ImageFile: "product7.jpg", Description: "经典星球杯，巧克力酱搭配脆饼粒。", IsGroupon: 1},
	{Name: "糖果组合礼盒", CategoryName: "休闲零食", Price: 19.90, GroupPrice: 17.90, Stock: 100, ImageFile: "product8.jpg", Description: "棒棒糖、巧克力和软糖组合，适合节日分享。", IsGroupon: 1},
	{Name: "炭烧夹心饼干", CategoryName: "休闲零食", Price: 5.90, GroupPrice: 4.90, Stock: 220, ImageFile: "product9.jpg", Description: "优酸乳味夹心饼干，香脆可口。", IsGroupon: 1},
	{Name: "杏仁曲奇饼干", CategoryName: "休闲零食", Price: 16.90, GroupPrice: 14.90, Stock: 140, ImageFile: "product10.jpg", Description: "酥松曲奇配整颗杏仁，下午茶优选。", IsGroupon: 1},
	{Name: "旺仔QQ糖", CategoryName: "休闲零食", Price: 6.90, GroupPrice: 5.90, Stock: 260, ImageFile: "product11.jpg", Description: "果味软糖，Q弹香甜。", IsGroupon: 1},
	{Name: "0.5mm中性笔笔芯", CategoryName: "文具用品", Price: 6.90, GroupPrice: 5.90, Stock: 300, ImageFile: "笔芯.jpg", Description: "黑色中性笔替芯，书写顺滑，学习办公常备。"},
	{Name: "四色荧光彩笔", CategoryName: "文具用品", Price: 8.90, GroupPrice: 7.90, Stock: 180, ImageFile: "彩笔.jpg", Description: "四色荧光标记笔，重点标注醒目。"},
	{Name: "双歧杆菌活菌胶囊", CategoryName: "健康药品", Price: 39.90, GroupPrice: 35.90, Stock: 60, ImageFile: "活菌胶囊.jpg", Description: "OTC肠道调理类药品，请按说明书或医嘱使用。"},
	{Name: "新鲜草莓", CategoryName: "新鲜水果", Price: 29.90, GroupPrice: 25.90, Stock: 120, ImageFile: "基础语法.jpg", Description: "鲜红饱满，酸甜多汁，冷藏口感更佳。", IsGroupon: 1},
	{Name: "颈舒颗粒", CategoryName: "健康药品", Price: 45.90, GroupPrice: 39.90, Stock: 50, ImageFile: "颈疏颗粒.jpg", Description: "OTC颗粒剂，请仔细阅读说明书并按需购买。"},
	{Name: "昆仑山雪山矿泉水", CategoryName: "酒水饮料", Price: 5.90, GroupPrice: 4.90, Stock: 240, ImageFile: "昆仑山.jpg", Description: "雪山矿泉水，清冽甘爽。", IsGroupon: 1},
	{Name: "牛黄解毒丸", CategoryName: "健康药品", Price: 12.90, GroupPrice: 10.90, Stock: 80, ImageFile: "牛黄解毒丸.jpg", Description: "OTC中成药，请按说明书或医嘱使用。"},
	{Name: "农夫山泉饮用天然水550ml", CategoryName: "酒水饮料", Price: 2.00, GroupPrice: 1.80, Stock: 500, ImageFile: "农夫山泉.jpg", Description: "550ml瓶装饮用天然水，日常补水。", IsGroupon: 1},
	{Name: "简约挂钟", CategoryName: "日用百货", Price: 29.90, GroupPrice: 26.90, Stock: 70, ImageFile: "时钟.jpg", Description: "简约数字挂钟，客厅卧室均可使用。"},
	{Name: "长柄梳子", CategoryName: "日用百货", Price: 3.90, GroupPrice: 3.50, Stock: 200, ImageFile: "梳子.jpg", Description: "长柄细齿梳，轻巧便携。"},
	{Name: "闪粉透明笔袋", CategoryName: "文具用品", Price: 12.90, GroupPrice: 10.90, Stock: 160, ImageFile: "透明笔袋.jpg", Description: "透明大容量笔袋，多款颜色可选。"},
	{Name: "彩色透明胶带", CategoryName: "文具用品", Price: 4.90, GroupPrice: 3.90, Stock: 260, ImageFile: "透明胶.jpg", Description: "彩色透明胶带，手账、封口、学习都适用。"},
	{Name: "居家编织拖鞋", CategoryName: "日用百货", Price: 19.90, GroupPrice: 16.90, Stock: 100, ImageFile: "拖鞋.jpg", Description: "柔软舒适，居家浴室均可穿。"},
	{Name: "维生素AD滴剂", CategoryName: "健康药品", Price: 32.90, GroupPrice: 29.90, Stock: 70, ImageFile: "维生素AD.jpg", Description: "OTC维生素补充剂，请按说明书使用。"},
	{Name: "防滑衣架", CategoryName: "日用百货", Price: 9.90, GroupPrice: 8.90, Stock: 180, ImageFile: "衣架.jpg", Description: "多色防滑衣架，承重稳固，不易变形。"},
	{Name: "依云天然矿泉水330ml", CategoryName: "酒水饮料", Price: 8.90, GroupPrice: 7.90, Stock: 180, ImageFile: "依云.jpg", Description: "330ml天然矿泉水，小瓶便携。"},
	{Name: "怡宝饮用纯净水", CategoryName: "酒水饮料", Price: 2.00, GroupPrice: 1.80, Stock: 500, ImageFile: "怡宝.jpg", Description: "怡宝瓶装饮用纯净水，清爽解渴。", IsGroupon: 1},
	{Name: "银黄颗粒", CategoryName: "健康药品", Price: 18.90, GroupPrice: 16.90, Stock: 90, ImageFile: "银黄颗粒.jpg", Description: "OTC颗粒剂，请阅读说明书后使用。"},
	{Name: "长柄雨伞", CategoryName: "日用百货", Price: 24.90, GroupPrice: 21.90, Stock: 110, ImageFile: "雨伞.jpg", Description: "长柄防雨伞，伞面宽大，通勤备用。"},
	{Name: "新鲜甜玉米", CategoryName: "时令蔬菜", Price: 4.90, GroupPrice: 3.90, Stock: 240, ImageFile: "玉米.jpg", Description: "颗粒饱满，香甜软糯，适合蒸煮。", IsGroupon: 1},
}
