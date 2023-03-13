package dto

type EsIndice struct {
	Health string `json:"health"`
	Status string `json:"status"`
	Index  string `json:"index"`
}

type SkuCount struct {
	Sku       string  `json:"sku"`
	Count     float64 `json:"count"`
	ProductId float64 `json:"product_id"`
}

type ReportStock struct {
	Sku string `json:"sku"`
	// Barcode           string `json:"barcode"`
	// Category          string `json:"category"`
	// BrandName         string `json:"brand_name"`
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
		skuCount := v.(map[string]interface{})
		res = append(res, &SkuCount{
			Sku:   skuCount["key"].(string),
			Count: skuCount["doc_count"].(float64),
		})
	}
	return res
}

func FromElasticSearchResponseToSkuGetAll(in []interface{}) []*SkuCount {
	res := make([]*SkuCount, 0)
	for _, v := range in {
		skuCount := v.(map[string]interface{})
		productId := skuCount["product_id"].(map[string]interface{})["buckets"].([]interface{})[0].(map[string]interface{})["key"].(float64)
		res = append(res, &SkuCount{
			Sku:       skuCount["key"].(string),
			Count:     skuCount["doc_count"].(float64),
			ProductId: productId,
		})
	}
	return res
}
