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
	Sku   string `json:"sku"`
	Count float64  `json:"count"`
}
