package common

import (
	"context"
	"errors"
	"fmt"
	"go_es_example/application/dto"
	"go_es_example/utils"
	"sync"
	"time"

	// "go_es_example/application/dto"

	"github.com/aquasecurity/esquery"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

var _ dto.SkuCount // For debugging; delete when done

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
		return nil, fmt.Errorf("cannot check index existence: %w", err)
	}
	if checkExistIndex.StatusCode == 404 {
		return nil, fmt.Errorf("not found index: %s", index)
	}
	SearchInventoryTest(es, index)

	return nil, nil
}

func SearchInventoryTest(es *elasticsearch.Client, index string) {
	start := time.Now()
	defer func() {
		fmt.Println("Execution Time: ", time.Since(start))
	}()

	key := "sku_count"
	query, err := QueryInStock(key)
	if err != nil {
		fmt.Printf("Error getting query in stock: %s", err)
	}

	resInventory, err := query.Run(es,
		es.Search.WithContext(context.TODO()),
		es.Search.WithIndex(index),
	)
	if err != nil {
		fmt.Printf("Error getting response: %s", err)
	}
	defer resInventory.Body.Close()
	utils.PrettyPrint(resInventory)

	t := utils.GetAggregationResponse(resInventory, key)

	resCountSku := make([]*dto.SkuCount, 0)

	for _, v := range t {
		skuCount := v.(map[string]interface{})
		resCountSku = append(resCountSku, &dto.SkuCount{
			Sku:   skuCount["key"].(string),
			Count: skuCount["doc_count"].(float64),
		})
	}
	utils.PrettyPrint(resCountSku)
}

func SearchInventoryRepo(es *elasticsearch.Client, index string) ([]interface{}, error) {
	start := time.Now()
	defer func() {
		fmt.Println("Execution Time: ", time.Since(start))
	}()

	key := "sku_count"
	wg := sync.WaitGroup{}
	resCountSku := make([]*esapi.Response, 0)
	queryInStock, err := QueryInStock(key)
	if err != nil {
		fmt.Printf("Error when query in-stock: %s", err)
	}
	queryCommitted, err := QueryCommitted(key)
	if err != nil {
		fmt.Printf("Error when query committed: %s", err)
	}

	wg.Add(2)

	go func(query *esquery.SearchRequest) {
		resInStock, err := query.Run(es,
			es.Search.WithContext(context.TODO()),
			es.Search.WithIndex(index),
		)
		utils.PrettyPrint(resInStock)
		if err != nil {
			fmt.Printf("Error getting response: %s", err)
		}
		defer resInStock.Body.Close()
		resCountSku = append(resCountSku, resInStock)
		wg.Done()
	}(queryInStock)

	go func(query *esquery.SearchRequest) {
		resCommitted, err := queryCommitted.Run(es,
			es.Search.WithContext(context.TODO()),
			es.Search.WithIndex(index),
		)
		if err != nil {
			fmt.Printf("Error getting response: %s", err)
		}
		defer resCommitted.Body.Close()
		resCountSku = append(resCountSku, resCommitted)
		wg.Done()
	}(queryCommitted)

	wg.Wait()

	for _, res := range resCountSku {
		for _, v := range utils.GetAggregationResponse(res, key) {
			// skuCount := v.(map[string]interface{})
			// resCountSku = append(resCountSku, &dto.SkuCount{
			// 	Sku:   skuCount["key"].(string),
			// 	Count: skuCount["doc_count"].(float64),
			// })
			// utils.PrettyPrint(v)
			fmt.Println(v)
		}
	}
	// for _, v := range t {
	// 	skuCount := v.(map[string]interface{})
	// 	resCountSku = append(resCountSku, &dto.SkuCount{
	// 		Sku:   skuCount["key"].(string),
	// 		Count: skuCount["doc_count"].(float64),
	// 	})
	// }

	// utils.PrettyPrint(resCountSku)
	// utils.PrettyPrint(t)
	return nil, nil
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
		return nil, errors.New(fmt.Sprintf("Error.Parsing: %s", err))
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
		return nil, errors.New(fmt.Sprintf("Error.Parsing: %s", err))
	}
	fmt.Printf("Query: %s\n", string(result))

	return query, nil
}

// func BuildQueryCountSku([]*int status_id) *esquery.SearchRequest {

// }
