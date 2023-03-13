package dto

type EsIndice struct {
	Health string `json:"health"`
	Status string `json:"status"`
	Index  string `json:"index"`
}

// type EsReponse struct {
// 	Took int64                  `json:"took"`
// 	Hits map[string]interface{} `json:"hits`
// }

type SkuCount struct {
	Sku   string  `json:"sku"`
	Count float64 `json:"count"`
}

type ReportStock struct {
	SKU               string `json:"sku"`
	Barcode           string `json:"barcode"`
	Category          string `json:"category"`
	BrandName         string `json:"brand_name"`
	WarehouseName     string `json:"warehouse_name"`
	Type              string `json:"type"`
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

func FromElasticSearchResponseToReportStock(in []interface{}) []SkuCount {
	res := make([]SkuCount, 0)
	for _, v := range in {
		skuCount := v.(map[string]interface{})
		res = append(res, SkuCount{
			Sku:   skuCount["key"].(string),
			Count: skuCount["doc_count"].(float64),
		})
	}
	return res
}
