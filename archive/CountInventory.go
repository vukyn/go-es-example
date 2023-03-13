package archive


// This file for debugging purposes only.
import (
	"context"
	"fmt"
	"go_es_example/application/dto"
	"go_es_example/application/query"
	"go_es_example/utils"
	"sync"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
)

// No concurrency, many loop and not optimized
func CountInventoryRepoV1(es *elasticsearch.Client, index string) ([]*dto.ReportStock, error) {
	start := time.Now()
	defer func() {
		fmt.Println("Execution Time: ", time.Since(start))
	}()

	keyAllSku := "sku_count_all"
	keyInStock := "sku_count_in_stock"
	keyCommitted := "sku_count_committed"
	reportStock := make([]*dto.ReportStock, 0)
	var countSkuInStock []*dto.SkuCount
	var countSkuCommitted []*dto.SkuCount

	// Query all SKU
	QueryAggAllSku, err := query.QueryAggAllSku(keyAllSku)
	if err != nil {
		fmt.Printf("Error when query all sku: %s", err)
	}
	resAllSku, err := QueryAggAllSku.Run(es,
		es.Search.WithContext(context.TODO()),
		es.Search.WithIndex(index),
	)
	if err != nil {
		fmt.Printf("Error getting response: %s", err)
	}
	defer resAllSku.Body.Close()
	listSku := dto.FromElasticSearchResponseToReportStock(utils.GetAggregationResponse(resAllSku, keyAllSku))
	for _, v := range listSku {
		reportStock = append(reportStock, &dto.ReportStock{
			Sku: v.Sku,
		})
	}

	// Query In-stock
	QueryAggInStock, err := query.QueryAggInStock(keyInStock)
	if err != nil {
		fmt.Printf("Error when query in-stock: %s", err)
	}
	resInStock, err := QueryAggInStock.Run(es,
		es.Search.WithContext(context.TODO()),
		es.Search.WithIndex(index),
	)
	if err != nil {
		fmt.Printf("Error getting response: %s", err)
	}
	defer resInStock.Body.Close()
	countSkuInStock = dto.FromElasticSearchResponseToReportStock(utils.GetAggregationResponse(resInStock, keyInStock))
	for _, v := range countSkuInStock {
		for i, y := range reportStock {
			if y.Sku == v.Sku {
				reportStock[i].InStock = int64(v.Count)
				break
			}
		}
	}

	// Query Committed
	QueryAggCommitted, err := query.QueryAggCommitted(keyCommitted)
	if err != nil {
		fmt.Printf("Error when query committed: %s", err)
	}
	resCommitted, err := QueryAggCommitted.Run(es,
		es.Search.WithContext(context.TODO()),
		es.Search.WithIndex(index),
	)
	if err != nil {
		fmt.Printf("Error getting response: %s", err)
	}
	defer resCommitted.Body.Close()
	countSkuCommitted = dto.FromElasticSearchResponseToReportStock(utils.GetAggregationResponse(resCommitted, keyCommitted))
	for _, v := range countSkuCommitted {
		for i, y := range reportStock {
			if y.Sku == v.Sku {
				reportStock[i].Committed = int64(v.Count)
				break
			}
		}
	}

	return reportStock, nil
}

// Add concurrency, many loop and not optimized
func CountInventoryRepoV2(es *elasticsearch.Client, index string) ([]*dto.ReportStock, error) {
	start := time.Now()
	defer func() {
		fmt.Println("Execution Time: ", time.Since(start))
	}()

	keyAllSku := "sku_count_all"
	keyInStock := "sku_count_in_stock"
	keyCommitted := "sku_count_committed"
	wg := sync.WaitGroup{}
	reportStock := make([]*dto.ReportStock, 0)
	var countSkuInStock []*dto.SkuCount
	var countSkuCommitted []*dto.SkuCount

	wg.Add(1)
	// Query all SKU
	go func() {
		query, err := query.QueryAggAllSku(keyAllSku)
		if err != nil {
			fmt.Printf("Error when query all sku: %s", err)
		}
		resAllSku, err := query.Run(es,
			es.Search.WithContext(context.TODO()),
			es.Search.WithIndex(index),
		)
		if err != nil {
			fmt.Printf("Error getting response: %s", err)
		}
		defer resAllSku.Body.Close()
		listSku := dto.FromElasticSearchResponseToReportStock(utils.GetAggregationResponse(resAllSku, keyAllSku))
		for _, v := range listSku {
			reportStock = append(reportStock, &dto.ReportStock{
				Sku: v.Sku,
			})
		}
		wg.Done()
	}()

	wg.Wait()

	wg.Add(2)
	// Query In-stock
	go func() {
		query, err := query.QueryAggInStock(keyInStock)
		if err != nil {
			fmt.Printf("Error when query in-stock: %s", err)
		}
		resInStock, err := query.Run(es,
			es.Search.WithContext(context.TODO()),
			es.Search.WithIndex(index),
		)
		if err != nil {
			fmt.Printf("Error getting response: %s", err)
		}
		defer resInStock.Body.Close()
		countSkuInStock = dto.FromElasticSearchResponseToReportStock(utils.GetAggregationResponse(resInStock, keyInStock))
		for _, v := range countSkuInStock {
			for i, y := range reportStock {
				if y.Sku == v.Sku {
					reportStock[i].InStock = int64(v.Count)
					break
				}
			}
		}
		wg.Done()
	}()

	// Query Committed
	go func() {
		query, err := query.QueryAggCommitted(keyCommitted)
		if err != nil {
			fmt.Printf("Error when query committed: %s", err)
		}
		resCommitted, err := query.Run(es,
			es.Search.WithContext(context.TODO()),
			es.Search.WithIndex(index),
		)
		if err != nil {
			fmt.Printf("Error getting response: %s", err)
		}
		defer resCommitted.Body.Close()
		countSkuCommitted = dto.FromElasticSearchResponseToReportStock(utils.GetAggregationResponse(resCommitted, keyCommitted))
		for _, v := range countSkuCommitted {
			for i, y := range reportStock {
				if y.Sku == v.Sku {
					reportStock[i].Committed = int64(v.Count)
					break
				}
			}
		}
		wg.Done()
	}()

	wg.Wait()

	return reportStock, nil
}

// Add concurrency, reduce some loop and optimized
func CountInventoryRepoV3(es *elasticsearch.Client, index string) ([]*dto.ReportStock, error) {
	start := time.Now()
	defer func() {
		fmt.Println("Execution Time: ", time.Since(start))
	}()

	keyAllSku := "sku_count_all"
	keyInStock := "sku_count_in_stock"
	keyCommitted := "sku_count_committed"
	wg := sync.WaitGroup{}
	reportStock := make([]*dto.ReportStock, 0)
	var countSkuInStock []*dto.SkuCount
	var countSkuCommitted []*dto.SkuCount

	wg.Add(1)
	// Query all SKU
	go func() {
		query, err := query.QueryAggAllSku(keyAllSku)
		if err != nil {
			fmt.Printf("Error when query all sku: %s", err)
		}
		resAllSku, err := query.Run(es,
			es.Search.WithContext(context.TODO()),
			es.Search.WithIndex(index),
		)
		if err != nil {
			fmt.Printf("Error getting response: %s", err)
		}
		defer resAllSku.Body.Close()
		listSku := dto.FromElasticSearchResponseToReportStock(utils.GetAggregationResponse(resAllSku, keyAllSku))
		for _, v := range listSku {
			reportStock = append(reportStock, &dto.ReportStock{
				Sku: v.Sku,
			})
		}
		wg.Done()
	}()

	wg.Wait()

	wg.Add(2)
	// Query In-stock
	go func() {
		query, err := query.QueryAggInStock(keyInStock)
		if err != nil {
			fmt.Printf("Error when query in-stock: %s", err)
		}
		resInStock, err := query.Run(es,
			es.Search.WithContext(context.TODO()),
			es.Search.WithIndex(index),
		)
		if err != nil {
			fmt.Printf("Error getting response: %s", err)
		}
		defer resInStock.Body.Close()
		countSkuInStock = dto.FromElasticSearchResponseToReportStock(utils.GetAggregationResponse(resInStock, keyInStock))
		wg.Done()
	}()

	// Query Committed
	go func() {
		query, err := query.QueryAggCommitted(keyCommitted)
		if err != nil {
			fmt.Printf("Error when query committed: %s", err)
		}
		resCommitted, err := query.Run(es,
			es.Search.WithContext(context.TODO()),
			es.Search.WithIndex(index),
		)
		if err != nil {
			fmt.Printf("Error getting response: %s", err)
		}
		defer resCommitted.Body.Close()
		countSkuCommitted = dto.FromElasticSearchResponseToReportStock(utils.GetAggregationResponse(resCommitted, keyCommitted))
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
	}

	return reportStock, nil
}
