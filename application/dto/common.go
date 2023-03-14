package dto

import (
	_ "go_es_example/utils"
)

type EsIndice struct {
	Health string `json:"health"`
	Status string `json:"status"`
	Index  string `json:"index"`
}

type SkuCount struct {
	Sku         string  `json:"sku"`
	Count       float64 `json:"count"`
	ProductId   float64 `json:"product_id"`
	WarehouseId float64 `json:"warehouse_id"`
	BrandId     float64 `json:"brand_id"`
}

type ReportStock struct {
	Sku string `json:"sku"`
	// Barcode           string `json:"barcode"`
	// Category          string `json:"category"`
	BrandId int64 `json:"brand_id"`
	// BrandName string `json:"brand_name"`
	WarehouseId int64 `json:"warehouse_id"`
	// WarehouseName     string `json:"warehouse_name"`
	// Type              string `json:"type"`
	ProductId int64 `json:"product_id"`
	// ProductName       string `json:"product_name"`
	InStock   int64 `json:"in_stock"`
	Committed int64 `json:"committed"`
	// Available         int64  `json:"available"`
	// InComming         int64  `json:"in_comming"`
	Receving int64 `json:"receving"`
	// PickOrder         int64  `json:"pick_order"`
	// PackOrder         int64  `json:"pack_order"`
	// PickIT            int64  `json:"pick_it"`
	// PackIT            int64  `json:"pack_it"`
	Test int64 `json:"test"`
	// UnsuitableProduct int64  `json:"unsuitable_product"`
}

type ReportStockResponse struct {
	Size   int64          `json:"size"`
	Record []*ReportStock `json:"record"`
}

func FromElasticSearchResponseToSkuCount(in []interface{}) []*SkuCount {
	res := make([]*SkuCount, 0)
	for _, v := range in {
		sku := v.(map[string]interface{})
		skuCount := v.(map[string]interface{})["warehouse_id"].(map[string]interface{})["buckets"].([]interface{})
		for _, k := range skuCount {
			res = append(res, &SkuCount{
				Sku:         sku["key"].(string),
				Count:       k.(map[string]interface{})["doc_count"].(float64),
				WarehouseId: k.(map[string]interface{})["key"].(float64),
			})
		}
	}
	return res
}

func FromElasticSearchResponseToSkuGetAll(in []interface{}) []*SkuCount {
	res := make([]*SkuCount, 0)
	for _, v := range in {

		var skuCount SkuCount
		sku := v.(map[string]interface{})

		// Get product_id from buckets
		product := sku["product_id"].(map[string]interface{})["buckets"].([]interface{})
		if len(product) > 0 {
			skuCount.ProductId = product[0].(map[string]interface{})["key"].(float64)
		}

		// Get brand_id from buckets
		brand := product[0].(map[string]interface{})["brand_id"].(map[string]interface{})["buckets"].([]interface{})
		if len(brand) > 0 {
			skuCount.BrandId = brand[0].(map[string]interface{})["key"].(float64)

			// Get warehouse_id from buckets
			warehouse := brand[0].(map[string]interface{})["warehouse_id"].(map[string]interface{})["buckets"].([]interface{})
			for _, w := range warehouse {
				res = append(res, &SkuCount{
					Sku:         sku["key"].(string),
					ProductId:   skuCount.ProductId,
					BrandId:     skuCount.BrandId,
					WarehouseId: w.(map[string]interface{})["key"].(float64),
				})
			}
		} else {
			res = append(res, &SkuCount{
				Sku:       sku["key"].(string),
				ProductId: skuCount.ProductId,
			})
		}
	}
	return res
}
