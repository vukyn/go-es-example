package dto

import (
	_ "go_es_example/utils"
	"strconv"
)

type EsIndice struct {
	Health string `json:"health"`
	Status string `json:"status"`
	Index  string `json:"index"`
}

type SkuCount struct {
	Sku           float64 `json:"sku"`
	Count         float64 `json:"count"`
	ProductId     float64 `json:"product_id"`
	WarehouseId   float64 `json:"warehouse_id"`
	BrandId       float64 `json:"brand_id"`
	InboundShmtId float64 `json:"inbound_shmt_id"`
}

type Product struct {
	Sku         float64 `json:"sku"`
	Barcode     string  `json:"barcode"`
	Type        float64 `json:"type"`
	ProductName string  `json:"product_name"`
	BrandId     float64 `json:"brand_id"`
}

type ReportStockRequest struct {
	WarehouseId int64  `json:"warehouse_id"`
	Sku         string `json:"sku"`
	Type        int32  `json:"type"`
	BrandIds    string `json:"brand_ids"`
	PageSize    int64  `json:"page_size"`
	PageNumber  int64  `json:"page_number"`
}

type ReportStock struct {
	Sku     float64 `json:"sku"`
	Barcode string  `json:"barcode"`
	// Category          string `json:"category"`
	BrandId int64 `json:"brand_id"`
	// BrandName string `json:"brand_name"`
	WarehouseId int64 `json:"warehouse_id"`
	// WarehouseName     string `json:"warehouse_name"`
	Type              int32  `json:"type"`
	ProductName       string `json:"product_name"`
	InStock           int64  `json:"in_stock"`
	Committed         int64  `json:"committed"`
	Available         int64  `json:"available"`
	InComming         int64  `json:"in_comming"`
	Receving          int64  `json:"receving"`
	PickOrder         int64  `json:"pick_order"`
	PackOrder         int64  `json:"pack_order"`
	PickIT            int64  `json:"pick_it"`
	PackIT            int64  `json:"pack_it"`
	Test              int64  `json:"test"`
	UnsuitableProduct int64  `json:"unsuitable_product"`
}

type ReportStockResponse struct {
	Count  int64          `json:"count"`
	Page   int64          `json:"page"`
	Size   int64          `json:"size"`
	Record []*ReportStock `json:"record"`
}

func FromElasticSearchResponseToSkuCount(in []interface{}) []*SkuCount {
	res := make([]*SkuCount, 0)
	for _, v := range in {
		sku := v.(map[string]interface{})
		warehouse := v.(map[string]interface{})["warehouse_id"].(map[string]interface{})["buckets"].([]interface{})
		for _, w := range warehouse {
			skuF, _ := strconv.ParseFloat(sku["key"].(string), 64)
			res = append(res, &SkuCount{
				Sku:         skuF,
				Count:       w.(map[string]interface{})["doc_count"].(float64),
				WarehouseId: w.(map[string]interface{})["key"].(float64),
			})
		}
	}
	return res
}

func FromElasticSearchResponseToListProduct(in []interface{}) []*Product {
	res := make([]*Product, 0)
	for _, v := range in {
		product := v.(map[string]interface{})["_source"].(map[string]interface{})["product"].(map[string]interface{})
		res = append(res, &Product{
			Sku:         product["product_sku"].(float64),
			Barcode:     product["product_barcode"].(string),
			Type:        product["product_type"].(float64),
			ProductName: product["product_name"].(string),
			BrandId:     product["product_brand_id"].(float64),
		})
	}
	return res
}
