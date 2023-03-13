package common

import (
	"context"
	"encoding/json"
	"fmt"
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

func SearchInventoryUseCase(es *elasticsearch.Client, index string) {
	checkExistIndex, err := es.Indices.Exists([]string{index})
	if err != nil {
		fmt.Printf("cannot check index existence: %s", err)
	}
	if checkExistIndex.StatusCode == 404 {
		fmt.Printf("not found index: %s", index)
	}

	countSkuRepo, err := CountInventoryRepo(es, index)
	if err != nil {
		fmt.Printf("CountInventoryRepo.Err %s", err)
	}

	byteCountSku, err := json.Marshal(&dto.ReportStockResponse{
		Size:   int64(len(countSkuRepo)),
		Record: countSkuRepo,
	})
	if err != nil {
		fmt.Printf("Json.Marshal.Err %s", err)
	}
	if err := utils.WriteFile(string(byteCountSku)); err != nil {
		fmt.Printf("WriteFile.Err %s", err)
	}
}

func CountInventoryRepo(es *elasticsearch.Client, index string) ([]*dto.ReportStock, error) {
	start := time.Now()
	defer func() {
		fmt.Println("Execution Time: ", time.Since(start))
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
		query, err := query.QueryAggAllSku(keyAllSku)
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
		listSku := dto.FromElasticSearchResponseToSkuGetAll(utils.GetAggregationResponse(res, keyAllSku))
		for _, v := range listSku {
			reportStock = append(reportStock, &dto.ReportStock{
				Sku: v.Sku,
				ProductId: int64(v.ProductId),
			})
		}
		wg.Done()
	}()

	wg.Wait()

	wg.Add(3)
	// Query In-stock
	go func() {
		query, err := query.QueryAggInStock(keyInStock)
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
		query, err := query.QueryAggCommitted(keyCommitted)
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
		query, err := query.QueryAggReceiving(keyReceiving)
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
			return v.Sku == i.Sku
		})
		if inStockItem != nil {
			reportStock[i].InStock = int64(inStockItem.Count)
		}

		committedItem := utils.Find(countSkuCommitted, func(i *dto.SkuCount) bool {
			return v.Sku == i.Sku
		})
		if committedItem != nil {
			reportStock[i].Committed = int64(committedItem.Count)
		}

		receivingItem := utils.Find(countSkuReceiving, func(i *dto.SkuCount) bool {
			return v.Sku == i.Sku
		})
		if receivingItem != nil {
			reportStock[i].Receving = int64(receivingItem.Count)
		}
	}

	return reportStock, nil
}
