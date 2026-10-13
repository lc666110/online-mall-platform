// Explicit opt-in demo catalog. Never run automatically in production.
package main

import (
	"log"
	"online-mall-platform/server/config"
	"online-mall-platform/server/database"
	"online-mall-platform/server/model"
)

func main() {
	c, e := config.Load()
	if e != nil {
		log.Fatal(e)
	}
	db, r, e := database.Open(c)
	if e != nil {
		log.Fatal(e)
	}
	defer r.Close()
	cats := []string{"居家生活", "数码配件", "出行好物", "日常穿搭"}
	ids := []uint{}
	for _, name := range cats {
		cat := model.Category{Name: name}
		if e := db.FirstOrCreate(&cat, model.Category{Name: name}).Error; e != nil {
			log.Fatal(e)
		}
		ids = append(ids, cat.ID)
	}
	names := []string{"静音无线耳机", "原木弧线休闲椅", "轻量日常帆布包", "陶瓷手冲咖啡壶", "柔光阅读台灯", "随行保温水杯", "极简桌面时钟", "云感针织围巾"}
	catIdx := []int{1, 0, 2, 0, 0, 2, 1, 3}
	prices := []int64{29900, 68900, 12900, 15900, 23900, 9900, 8900, 16900}
	for i, name := range names {
		var n int64
		db.Model(&model.Product{}).Where("name = ?", name).Count(&n)
		if n > 0 {
			continue
		}
		p := model.Product{Name: name, CategoryID: ids[catIdx[i]], Description: "用简洁设计，留住生活的舒适。\n精选材质，细致打磨，适合每一天的使用。\n此为可删除的演示商品，请根据真实商品补充材质、尺寸和保障说明。", Images: "[\"/images/product-" + string(rune('1'+i)) + ".svg\"]", Active: true, Variants: []model.Variant{{SKU: "MORI-" + string(rune('1'+i)) + "-A", Specification: "自然色 / 标准款", Price: prices[i], Stock: 50}, {SKU: "MORI-" + string(rune('1'+i)) + "-B", Specification: "深色 / 经典款", Price: prices[i] + 2000, Stock: 30}}}
		if e := db.Create(&p).Error; e != nil {
			log.Fatal(e)
		}
	}
	log.Print("Demo catalog ready")
}
