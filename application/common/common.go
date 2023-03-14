package common

import (
	"context"
	"encoding/json"
	"fmt"
	"go_es_example/application/constant"
	"go_es_example/application/dto"
	"go_es_example/application/query"
	"go_es_example/utils"
	"sync"
	"time"

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

	// Query report get count
	reportStock, err := ReportStockRepo(es, indexInventory)
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
	if err := utils.WriteFile(string(byteListProduct), "data_list_product"); err != nil {
		fmt.Printf("WriteFile.Err %s", err)
	}

	for i, v := range reportStock {
		product := utils.Find(listProduct, func(i *dto.Product) bool {
			return v.Sku == i.Sku
		})
		if product != nil {
			reportStock[i].ProductName = product.ProductName
			reportStock[i].Barcode = product.Barcode
			reportStock[i].Type = int32(product.Type)
		}
	}

	byteCountSku, err := json.Marshal(&dto.ReportStockResponse{
		Size:   int64(len(reportStock)),
		Record: reportStock,
	})
	if err != nil {
		fmt.Printf("Json.Marshal.Err %s", err)
	}
	if err := utils.WriteFile(string(byteCountSku), "data_count_sku"); err != nil {
		fmt.Printf("WriteFile.Err %s", err)
	}
}

func ReportStockRepo(es *elasticsearch.Client, index string) ([]*dto.ReportStock, error) {
	start := time.Now()
	defer func() {
		fmt.Println("Execution Time CountSku: ", time.Since(start))
	}()

	keyAllSku := "sku_count_all"
	keyInStock := "sku_count_in_stock"
	keyCommitted := "sku_count_committed"
	keyReceiving := "sku_count_receiving"
	wg := sync.WaitGroup{}
	reportStock := make([]*dto.ReportStock, 0)
	var countSkuInStock []*dto.SkuCount
	var countSkuCommitted []*dto.SkuCount
	var countSkuReceiving []*dto.SkuCount

	wg.Add(1)
	// Query all SKU
	go func() {
		query, err := query.QueryAggSku(keyAllSku, false, nil, nil)
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

	wg.Add(3)
	// Query In-stock
	go func() {
		query, err := query.QueryAggSku(keyInStock, false, constant.SKU_STATUS_IN_STOCK, nil)
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
		query, err := query.QueryAggSku(keyCommitted, false, constant.SKU_STATUS_COMMITTED, nil)
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
		query, err := query.QueryAggSku(keyReceiving, false, constant.SKU_STATUS_RECEIVING, nil)
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
