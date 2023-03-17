package common

import (
	"context"
	"encoding/json"
	"fmt"
	"go_es_example/application/constant"
	"go_es_example/application/dto"
	"go_es_example/application/query"
	"go_es_example/utils"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aquasecurity/esquery"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

var _ dto.SkuCount // For debugging; delete when done
var _ json.Number  // For debugging; delete when done

func GetAllIndexes(es *elasticsearch.Client) (string, error) {
	res, err := esapi.CatIndicesRequest{Format: "json"}.Do(context.Background(), es)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	return res.String(), nil
}

func ReportStockUseCase(es *elasticsearch.Client) {
	indexInventory := "qc_fulfillment.qc_fulfillment_inventory.wms_inventory"
	indexProduct := "fulfillment.qc.cdc.final_product_v2"

	// Check exist index
	checkExistIndex, err := es.Indices.Exists([]string{indexInventory, indexProduct})
	if err != nil {
		fmt.Printf("cannot check index existence: %s", err)
	}
	if checkExistIndex.StatusCode == 404 {
		fmt.Printf("not found index: %s", checkExistIndex.Body)
	}

	// Filter
	params := &dto.ReportStockRequest{
		// WarehouseId: 19,
		// Sku:         "100240042,100170031",
		// BrandIds:    "1015",
		// Type: 2,
		// PageSize:   10,
		// PageNumber: 1,
	}

	//Query report
	reportStock, err := ReportStockRepo(es, indexInventory, params)
	if err != nil {
		fmt.Printf("CountInventoryRepo.Err %s", err)
	}

	// Mapping product information
	listSku := make([]float64, 0)
	for _, v := range reportStock {
		listSku = append(listSku, v.Sku)
	}
	listProduct, err := GetListProductRepo(es, indexProduct, utils.Distinct(listSku))
	if err != nil {
		fmt.Printf("GetListProductRepo.Err %s", err)
	}
	byteListProduct, err := json.Marshal(listProduct)
	if err != nil {
		fmt.Printf("Json.Marshal.Err %s", err)
	}
	if err := utils.WriteFile(string(byteListProduct), "data/data_list_product.txt"); err != nil {
		fmt.Printf("WriteFile.Err %s", err)
	}
	fmt.Println("write file product done!")
	for i, v := range reportStock {
		product := utils.Find(listProduct, func(i *dto.Product) bool {
			return v.Sku == i.Sku
		})
		if product != nil {
			reportStock[i].ProductName = product.ProductName
			reportStock[i].Barcode = product.Barcode
			reportStock[i].Type = int32(product.Type)
			reportStock[i].BrandId = int64(product.BrandId)
		}
	}
	// End Mapping

	// Additional filters
	if params.Type != 0 {
		reportStock = utils.Where(reportStock, func(i *dto.ReportStock) bool {
			return i.Type == params.Type
		})
	}
	// End additional filters
	count := int64(len(reportStock))

	// Paging
	if len(reportStock) > 0 && params.PageNumber != 0 && params.PageSize != 0 {
		from := ((params.PageNumber - 1) * params.PageSize)
		reportStock = reportStock[from : from+params.PageSize]
	}
	// End paging

	byteCountSku, err := json.Marshal(&dto.ReportStockResponse{
		Size:   int64(len(reportStock)),
		Page:   params.PageNumber,
		Count:  count,
		Record: reportStock,
	})
	if err != nil {
		fmt.Printf("Json.Marshal.Err %s", err)
	}
	if err := utils.WriteFile(string(byteCountSku), "data/data_count_sku.txt"); err != nil {
		fmt.Printf("WriteFile.Err %s", err)
	}
	fmt.Println("write file count sku done!")
}

func ReportStockRepo(es *elasticsearch.Client, index string, params *dto.ReportStockRequest) ([]*dto.ReportStock, error) {
	start := time.Now()
	defer func() {
		fmt.Println("Execution Time CountSku: ", time.Since(start))
	}()

	wg := sync.WaitGroup{}
	keyAllSku := "sku_count_all"
	keyInStock := "sku_count_in_stock"
	keyCommitted := "sku_count_committed"
	keyReceiving := "sku_count_receiving"
	keyInComming := "sku_count_in_comming"
	keyAvailable := "sku_count_available"
	keyTest := "sku_count_test"
	keyPickOrder := "sku_count_pick_order"
	keyPackOrder := "sku_count_pack_order"
	keyPickIT := "sku_count_pick_IT"
	keyPackIT := "sku_count_pack_IT"
	keyUP := "sku_count_UP"
	var countSkuInStock []*dto.SkuCount
	var countSkuCommitted []*dto.SkuCount
	var countSkuReceiving []*dto.SkuCount
	var countSkuInComming []*dto.SkuCount
	var countSkuAvailable []*dto.SkuCount
	var countSkuTest []*dto.SkuCount
	var countSkuPickOrder []*dto.SkuCount
	var countSkuPackOrder []*dto.SkuCount
	var countSkuPickIT []*dto.SkuCount
	var countSkuPackIT []*dto.SkuCount
	var countSkuUP []*dto.SkuCount
	reportStock := make([]*dto.ReportStock, 0)

	req := make([]esquery.Mappable, 0)
	must := make([]esquery.Mappable, 0)
	if params.Sku != "" {
		should_sku := make([]esquery.Mappable, 0)
		for _, sku := range strings.Split(params.Sku, ",") {
			should_sku = append(should_sku, esquery.Term("sku", sku))
		}
		must = append(must, esquery.Bool().Should(should_sku...))
	}
	if params.BrandIds != "" {
		should_brand := make([]esquery.Mappable, 0)
		for _, brand := range strings.Split(params.BrandIds, ",") {
			id, err := strconv.Atoi(brand)
			if err != nil {
				return nil, err
			}
			should_brand = append(should_brand, esquery.Term("brand_id", id))
		}
		must = append(must, esquery.Bool().Should(should_brand...))
	}
	if params.WarehouseId != 0 {
		must = append(must, esquery.Term("warehouse_id", params.WarehouseId))
	}
	req = append(req, esquery.Bool().Must(must...))

	wg.Add(1)
	// Query all SKU
	go func() {
		query, err := query.QueryAggSku(keyAllSku, false, nil, nil, req)
		if err != nil {
			fmt.Printf("Error when query all sku: %s", err)
		}
		res, err := query.Run(es,
			es.Search.WithContext(context.TODO()),
			es.Search.WithIndex(index),
		)
		if err != nil {
			fmt.Printf("Error getting response: %s", err)
		}
		defer res.Body.Close()
		listSku := dto.FromElasticSearchResponseToSkuCount(utils.GetAggregationResponse(res, keyAllSku))
		for _, v := range listSku {
			reportStock = append(reportStock, &dto.ReportStock{
				Sku:         v.Sku,
				WarehouseId: int64(v.WarehouseId),
			})
		}
		wg.Done()
	}()

	wg.Wait()

	wg.Add(11)
	// Query In-stock
	go func() {
		query, err := query.QueryAggSku(keyInStock, false, constant.SKU_STATUS_IN_STOCK, nil, req)
		if err != nil {
			fmt.Printf("Error when query in-stock: %s", err)
		}
		res, err := query.Run(es,
			es.Search.WithContext(context.TODO()),
			es.Search.WithIndex(index),
		)
		if err != nil {
			fmt.Printf("Error getting response: %s", err)
		}
		defer res.Body.Close()
		countSkuInStock = dto.FromElasticSearchResponseToSkuCount(utils.GetAggregationResponse(res, keyInStock))
		wg.Done()
	}()

	// Query Committed
	go func() {
		query, err := query.QueryAggSku(keyCommitted, false, constant.SKU_STATUS_COMMITTED, nil, req)
		if err != nil {
			fmt.Printf("Error when query committed: %s", err)
		}
		res, err := query.Run(es,
			es.Search.WithContext(context.TODO()),
			es.Search.WithIndex(index),
		)
		if err != nil {
			fmt.Printf("Error getting response: %s", err)
		}
		defer res.Body.Close()
		countSkuCommitted = dto.FromElasticSearchResponseToSkuCount(utils.GetAggregationResponse(res, keyCommitted))
		wg.Done()
	}()

	// Query Receiving
	go func() {
		query, err := query.QueryAggSku(keyReceiving, false, constant.SKU_STATUS_RECEIVING, nil, req)
		if err != nil {
			fmt.Printf("Error when query receving: %s", err)
		}
		res, err := query.Run(es,
			es.Search.WithContext(context.TODO()),
			es.Search.WithIndex(index),
		)
		if err != nil {
			fmt.Printf("Error getting response: %s", err)
		}
		defer res.Body.Close()
		countSkuReceiving = dto.FromElasticSearchResponseToSkuCount(utils.GetAggregationResponse(res, keyReceiving))
		wg.Done()
	}()

	// Query In-comming
	go func() {
		const KEY_INBOUND_SHMT_ID = "inbound_shmt_id"
		listInboundShmtId := make([]float64, 0)
		indexInboundShmt := "qc_fulfillment.qc_fulfillment_inventory.wms_inbound_shmt"
		indexInboundShmtItem := "qc_fulfillment.qc_fulfillment_inventory.wms_inbound_shmt_item"

		// Build query count
		matchStatusId := make([]esquery.Mappable, 0)
		aggInComming := esquery.TermsAgg(keyInComming, "sku.keyword").
			Aggs(esquery.TermsAgg(KEY_INBOUND_SHMT_ID, KEY_INBOUND_SHMT_ID)).
			Size(1000)

		for _, v := range constant.SKU_STATUS_IN_COMMING {
			matchStatusId = append(matchStatusId, esquery.Term("status_id", v))
		}
		// End build query count

		// Call query count
		reqSkuCount := esquery.Search().
			Query(
				esquery.Bool().
					Should(matchStatusId...),
			).
			Aggs(aggInComming).
			Size(0)

		resSkuCount, err := reqSkuCount.Run(es,
			es.Search.WithContext(context.TODO()),
			es.Search.WithIndex(indexInboundShmtItem),
		)
		if err != nil {
			fmt.Printf("Error getting response: %s", err)
		}
		defer resSkuCount.Body.Close()
		// End call query call

		for _, v := range utils.GetAggregationResponse(resSkuCount, keyInComming) {
			sku := v.(map[string]interface{})
			inboundShmt := v.(map[string]interface{})[KEY_INBOUND_SHMT_ID].(map[string]interface{})["buckets"].([]interface{})
			for _, i := range inboundShmt {
				skuF, _ := strconv.ParseFloat(sku["key"].(string), 64)
				inboundShmtId := i.(map[string]interface{})["key"].(float64)
				listInboundShmtId = append(listInboundShmtId, inboundShmtId)
				countSkuInComming = append(countSkuInComming, &dto.SkuCount{
					Sku:           skuF,
					Count:         i.(map[string]interface{})["doc_count"].(float64),
					InboundShmtId: inboundShmtId,
				})
			}
		}

		// Build query inbound shmt
		termsInboundId := make([]esquery.Mappable, 0)
		for _, v := range utils.Distinct(listInboundShmtId) {
			termsInboundId = append(termsInboundId, esquery.Term("inbound_shmt_id", v))
		}
		queryInboundShmt, _ := query.QueryList(false, []esquery.Mappable{esquery.Bool().Should(termsInboundId...)})
		// End build query count

		// Call query list inbound shmt
		resListInboundShmt, err := queryInboundShmt.Run(es,
			es.Search.WithContext(context.TODO()),
			es.Search.WithIndex(indexInboundShmt),
		)
		if err != nil {
			fmt.Printf("Error getting response: %s", err)
		}
		defer resListInboundShmt.Body.Close()
		// End call query list inbound shmt

		hitsListInboundShmt := utils.GetHitsResponse(resListInboundShmt)
		for _, v := range countSkuInComming {
			inboundShmt := utils.Find(hitsListInboundShmt, func(i interface{}) bool {
				return i.(map[string]interface{})["_source"].(map[string]interface{})["inbound_shmt_id"] == v.InboundShmtId
			})
			if inboundShmt != nil {
				v.WarehouseId = inboundShmt.(map[string]interface{})["_source"].(map[string]interface{})["warehouse_id"].(float64)
			}
		}

		wg.Done()
	}()

	// Query Available
	go func() {
		query, err := query.QueryAggSku(keyAvailable, false, constant.SKU_STATUS_AVAILABLE, constant.PRODUCT_STATUS_AVAILABLE, req)
		if err != nil {
			fmt.Printf("Error when query receving: %s", err)
		}
		res, err := query.Run(es,
			es.Search.WithContext(context.TODO()),
			es.Search.WithIndex(index),
		)
		if err != nil {
			fmt.Printf("Error getting response: %s", err)
		}
		defer res.Body.Close()
		countSkuAvailable = dto.FromElasticSearchResponseToSkuCount(utils.GetAggregationResponse(res, keyAvailable))
		wg.Done()
	}()

	// Query Test
	go func() {
		query, err := query.QueryAggSku(keyTest, false, constant.SKU_STATUS_TEST, constant.PRODUCT_STATUS_TEST, req)
		if err != nil {
			fmt.Printf("Error when query receving: %s", err)
		}
		res, err := query.Run(es,
			es.Search.WithContext(context.TODO()),
			es.Search.WithIndex(index),
		)
		if err != nil {
			fmt.Printf("Error getting response: %s", err)
		}
		defer res.Body.Close()
		countSkuTest = dto.FromElasticSearchResponseToSkuCount(utils.GetAggregationResponse(res, keyTest))
		wg.Done()
	}()

	// Query Pick Order
	go func() {
		pick_type := esquery.Match("sales_order_type", "ORDER")
		query, err := query.QueryAggSku(keyPickOrder, false, constant.SKU_STATUS_PICK_ORDER, nil, append(req, pick_type))
		if err != nil {
			fmt.Printf("Error when query pick order: %s", err)
		}

		res, err := query.Run(es,
			es.Search.WithContext(context.TODO()),
			es.Search.WithIndex(index),
		)
		if err != nil {
			fmt.Printf("Error getting response: %s", err)
		}
		defer res.Body.Close()
		countSkuPickOrder = dto.FromElasticSearchResponseToSkuCount(utils.GetAggregationResponse(res, keyPickOrder))
		wg.Done()
	}()

	// Query Pack Order
	go func() {
		pack_type := esquery.Match("sales_order_type", "ORDER")
		query, err := query.QueryAggSku(keyPackOrder, false, constant.SKU_STATUS_PACK_ORDER, nil, append(req, pack_type))
		if err != nil {
			fmt.Printf("Error when query pack order: %s", err)
		}

		res, err := query.Run(es,
			es.Search.WithContext(context.TODO()),
			es.Search.WithIndex(index),
		)
		if err != nil {
			fmt.Printf("Error getting response: %s", err)
		}
		defer res.Body.Close()
		countSkuPackOrder = dto.FromElasticSearchResponseToSkuCount(utils.GetAggregationResponse(res, keyPackOrder))
		wg.Done()
	}()

	// Query Pick IT
	go func() {
		pick_type := esquery.Match("sales_order_type", "INTERNAL_TRANSFER")
		query, err := query.QueryAggSku(keyPickIT, false, constant.SKU_STATUS_PICK_IT, nil, append(req, pick_type))
		if err != nil {
			fmt.Printf("Error when query pick IT: %s", err)
		}

		res, err := query.Run(es,
			es.Search.WithContext(context.TODO()),
			es.Search.WithIndex(index),
		)
		if err != nil {
			fmt.Printf("Error getting response: %s", err)
		}
		defer res.Body.Close()
		countSkuPickIT = dto.FromElasticSearchResponseToSkuCount(utils.GetAggregationResponse(res, keyPickIT))
		wg.Done()
	}()

	// Query Pack IT
	go func() {
		pack_type := esquery.Match("sales_order_type", "INTERNAL_TRANSFER")
		query, err := query.QueryAggSku(keyPackIT, false, constant.SKU_STATUS_PACK_IT, nil, append(req, pack_type))
		if err != nil {
			fmt.Printf("Error when query pack IT: %s", err)
		}

		res, err := query.Run(es,
			es.Search.WithContext(context.TODO()),
			es.Search.WithIndex(index),
		)
		if err != nil {
			fmt.Printf("Error getting response: %s", err)
		}
		defer res.Body.Close()
		countSkuPackIT = dto.FromElasticSearchResponseToSkuCount(utils.GetAggregationResponse(res, keyPackIT))
		wg.Done()
	}()

	// Query UP
	go func() {
		query, err := query.QueryAggSku(keyUP, false, nil, constant.PRODUCT_STATUS_UP, req)
		if err != nil {
			fmt.Printf("Error when query UP: %s", err)
		}

		res, err := query.Run(es,
			es.Search.WithContext(context.TODO()),
			es.Search.WithIndex(index),
		)
		if err != nil {
			fmt.Printf("Error getting response: %s", err)
		}
		defer res.Body.Close()
		countSkuUP = dto.FromElasticSearchResponseToSkuCount(utils.GetAggregationResponse(res, keyUP))
		wg.Done()
	}()

	wg.Wait()

	// Merge data from everywhere to a single list
	for i, v := range reportStock {
		inStockItem := utils.Find(countSkuInStock, func(i *dto.SkuCount) bool {
			return v.Sku == i.Sku && v.WarehouseId == int64(i.WarehouseId)
		})
		if inStockItem != nil {
			reportStock[i].InStock = int64(inStockItem.Count)
		}

		committedItem := utils.Find(countSkuCommitted, func(i *dto.SkuCount) bool {
			return v.Sku == i.Sku && v.WarehouseId == int64(i.WarehouseId)
		})
		if committedItem != nil {
			reportStock[i].Committed = int64(committedItem.Count)
		}

		receivingItem := utils.Find(countSkuReceiving, func(i *dto.SkuCount) bool {
			return v.Sku == i.Sku && v.WarehouseId == int64(i.WarehouseId)
		})
		if receivingItem != nil {
			reportStock[i].Receving = int64(receivingItem.Count)
		}

		availableItem := utils.Find(countSkuAvailable, func(i *dto.SkuCount) bool {
			return v.Sku == i.Sku && v.WarehouseId == int64(i.WarehouseId)
		})
		if availableItem != nil {
			reportStock[i].Available = int64(availableItem.Count)
		}

		testItem := utils.Find(countSkuTest, func(i *dto.SkuCount) bool {
			return v.Sku == i.Sku && v.WarehouseId == int64(i.WarehouseId)
		})
		if testItem != nil {
			reportStock[i].Test = int64(testItem.Count)
		}

		pickOrderItem := utils.Find(countSkuPickOrder, func(i *dto.SkuCount) bool {
			return v.Sku == i.Sku && v.WarehouseId == int64(i.WarehouseId)
		})
		if pickOrderItem != nil {
			reportStock[i].PickOrder = int64(pickOrderItem.Count)
		}

		packOrderItem := utils.Find(countSkuPackOrder, func(i *dto.SkuCount) bool {
			return v.Sku == i.Sku && v.WarehouseId == int64(i.WarehouseId)
		})
		if packOrderItem != nil {
			reportStock[i].PackOrder = int64(packOrderItem.Count)
		}

		pickITItem := utils.Find(countSkuPickIT, func(i *dto.SkuCount) bool {
			return v.Sku == i.Sku && v.WarehouseId == int64(i.WarehouseId)
		})
		if pickITItem != nil {
			reportStock[i].PickIT = int64(pickITItem.Count)
		}

		packITItem := utils.Find(countSkuPackIT, func(i *dto.SkuCount) bool {
			return v.Sku == i.Sku && v.WarehouseId == int64(i.WarehouseId)
		})
		if packITItem != nil {
			reportStock[i].PackIT = int64(packITItem.Count)
		}

		UPItem := utils.Find(countSkuUP, func(i *dto.SkuCount) bool {
			return v.Sku == i.Sku && v.WarehouseId == int64(i.WarehouseId)
		})
		if UPItem != nil {
			reportStock[i].UnsuitableProduct = int64(UPItem.Count)
		}
	}

	// Merge in comming to ReportStock
	for _, v := range countSkuInComming {
		isSkuExist := false
		for i, x := range reportStock {
			if v.Sku == x.Sku && int64(v.WarehouseId) == x.WarehouseId {
				reportStock[i].InComming = int64(v.Count)
				isSkuExist = true
				break
			}
		}
		if !isSkuExist {
			reportStock = append(reportStock, &dto.ReportStock{
				Sku:       v.Sku,
				InComming: int64(v.Count),
			})
		} else {
			isSkuExist = false
		}
	}

	return reportStock, nil
}

func GetListProductRepo(es *elasticsearch.Client, index string, skus []float64) ([]*dto.Product, error) {
	start := time.Now()
	defer func() {
		fmt.Println("Execution Time GetListProduct: ", time.Since(start))
	}()

	query, err := query.QueryListProduct(skus, false)
	if err != nil {
		fmt.Printf("Error when query list product: %s", err)
	}
	res, err := query.Run(es,
		es.Search.WithContext(context.TODO()),
		es.Search.WithIndex(index),
	)
	if err != nil {
		fmt.Printf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	return dto.FromElasticSearchResponseToListProduct(utils.GetHitsResponse(res)), nil
}
