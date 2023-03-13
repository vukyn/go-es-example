package common

import (
	"context"
	"encoding/json"
	"fmt"
	"go_es_example/application/dto"
	"go_es_example/utils"
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

func SearchInventoryUseCase(es *elasticsearch.Client, index string) ([]interface{}, error) {
	checkExistIndex, err := es.Indices.Exists([]string{index})
	if err != nil {
		fmt.Printf("cannot check index existence: %s", err)
		return nil, fmt.Errorf("cannot check index existence: %s", err)
	}
	if checkExistIndex.StatusCode == 404 {
		fmt.Printf("not found index: %s", index)
		return nil, fmt.Errorf("not found index: %s", index)
	}

	countSkuRepo, err := CountInventoryRepo(es, index)
	if err != nil {
		fmt.Printf("CountInventoryRepo.Err %s", err)
		return nil, fmt.Errorf("CountInventoryRepo.Err %s", err)
	}

	byteCountSku, err := json.Marshal(&dto.ReportStockResponse{
		Size:   int64(len(countSkuRepo)),
		Record: countSkuRepo,
	})
	if err != nil {
		fmt.Printf("Json.Marshal.Err %s", err)
		return nil, fmt.Errorf("Json.Marshal.Err %s", err)
	}
	if err := utils.WriteFile(string(byteCountSku)); err != nil {
		fmt.Printf("WriteFile.Err %s", err)
		return nil, fmt.Errorf("WriteFile.Err %s", err)
	}
	return nil, nil
}

func CountInventoryRepo(es *elasticsearch.Client, index string) ([]*dto.ReportStock, error) {
	start := time.Now()
	defer func() {
		fmt.Println("Execution Time: ", time.Since(start))
	}()

	keyAllSku := "sku_count_all"
	keyInStock := "sku_count_in_stock"
	keyCommitted := "sku_count_committed"
	wg := sync.WaitGroup{}
	reportStock := make([]*dto.ReportStock, 0)

	queryAllSku, err := QueryAllSku(keyAllSku)
	if err != nil {
		fmt.Printf("Error when query all sku: %s", err)
	}

	wg.Add(1)
	go func(query *esquery.SearchRequest) {
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
				SKU: v.Sku,
			})
		}
		wg.Done()
	}(queryAllSku)

	queryInStock, err := QueryInStock(keyInStock)
	if err != nil {
		fmt.Printf("Error when query in-stock: %s", err)
	}
	queryCommitted, err := QueryCommitted(keyCommitted)
	if err != nil {
		fmt.Printf("Error when query committed: %s", err)
	}

	wg.Wait()

	wg.Add(2)
	// Query In-stock
	go func(query *esquery.SearchRequest) {
		resInStock, err := query.Run(es,
			es.Search.WithContext(context.TODO()),
			es.Search.WithIndex(index),
		)
		if err != nil {
			fmt.Printf("Error getting response: %s", err)
		}
		defer resInStock.Body.Close()
		countSkuInStock := dto.FromElasticSearchResponseToReportStock(utils.GetAggregationResponse(resInStock, keyInStock))
		for _, v := range countSkuInStock {
			for i, y := range reportStock {
				if y.SKU == v.Sku {
					reportStock[i].InStock = int64(v.Count)
					break
				}
			}
		}
		wg.Done()
	}(queryInStock)

	// Query Committed
	go func(query *esquery.SearchRequest) {
		resCommitted, err := queryCommitted.Run(es,
			es.Search.WithContext(context.TODO()),
			es.Search.WithIndex(index),
		)
		if err != nil {
			fmt.Printf("Error getting response: %s", err)
		}
		defer resCommitted.Body.Close()
		countSkuCommitted := dto.FromElasticSearchResponseToReportStock(utils.GetAggregationResponse(resCommitted, keyCommitted))
		for _, v := range countSkuCommitted {
			for i, y := range reportStock {
				if y.SKU == v.Sku {
					reportStock[i].Committed = int64(v.Count)
					break
				}
			}
		}
		wg.Done()
	}(queryCommitted)

	wg.Wait()
	return reportStock, nil
}

func QueryAllSku(key string) (*esquery.SearchRequest, error) {
	agg := esquery.TermsAgg(key, "sku.keyword").Size(1000)

	query := esquery.Search().Aggs(agg).Size(0)
	result, err := query.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("Error.Parsing: %s", err)
	}
	fmt.Printf("Query: %s\n", string(result))
	return query, nil
}

func QueryInStock(key string) (*esquery.SearchRequest, error) {
	agg := esquery.TermsAgg(key, "sku.keyword").Size(1000)

	query := esquery.Search().
		Query(
			esquery.Bool().Should(
				esquery.Match("status_id", 2),
				esquery.Match("status_id", 3),
				esquery.Match("status_id", 4),
				esquery.Match("status_id", 5),
				esquery.Match("status_id", 6),
				esquery.Match("status_id", 7),
				esquery.Match("status_id", 8),
				esquery.Match("status_id", 9),
				esquery.Match("status_id", 10),
				esquery.Match("status_id", 11),
				esquery.Match("status_id", 14),
				esquery.Match("status_id", 15),
				esquery.Match("status_id", 16),
				esquery.Match("status_id", 17),
				esquery.Match("status_id", 20),
				esquery.Match("status_id", 21),
				esquery.Match("status_id", 22),
				esquery.Match("status_id", 23),
				esquery.Match("status_id", 26),
				esquery.Match("status_id", 27),
				esquery.Match("status_id", 28),
				esquery.Match("status_id", 29),
				esquery.Match("status_id", 30),
				esquery.Match("status_id", 33),
			),
		).
		Aggs(agg).
		Size(0)

	result, err := query.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("Error.Parsing: %s", err)
	}
	fmt.Printf("Query: %s\n", string(result))

	return query, nil
}

func QueryCommitted(key string) (*esquery.SearchRequest, error) {
	agg := esquery.TermsAgg(key, "sku.keyword").Size(1000)

	query := esquery.Search().
		Query(
			esquery.Bool().Should(
				esquery.Match("status_id", 7),
				esquery.Match("status_id", 9),
				esquery.Match("status_id", 11),
				esquery.Match("status_id", 12),
			),
		).
		Aggs(agg).
		Size(0)

	result, err := query.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("Error.Parsing: %s", err)
	}
	fmt.Printf("Query: %s\n", string(result))

	return query, nil
}

// func BuildQueryCountSku([]*int status_id) *esquery.SearchRequest {

// }
