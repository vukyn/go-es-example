package common

import (
	"context"
	"encoding/json"
	"fmt"

	// "go_es_example/application/dto"
	"go_es_example/application/dto"
	"go_es_example/utils"

	"github.com/aquasecurity/esquery"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

func GetAllIndexes(es *elasticsearch.Client) (string, error) {
	res, err := esapi.CatIndicesRequest{Format: "json"}.Do(context.Background(), es)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	return res.String(), nil
}

func SearchInventoryUseCase(es *elasticsearch.Client, index string) ([]interface{}, error) {

	resExist, err := es.Indices.Exists([]string{index})
	if err != nil {
		return nil, fmt.Errorf("cannot check index existence: %w", err)
	}
	if resExist.StatusCode == 404 {
		return nil, fmt.Errorf("not found index: %s", index)
	}
	SearchInventoryRepo(es, index)

	return nil, nil
}

func SearchInventoryRepo(es *elasticsearch.Client, index string) ([]interface{}, error) {

	// query := esquery.
	// 	Search().
	// 	Query(
	// 		esquery.
	// 			Bool().
	// 			Must(esquery.MatchAll()),
	// 	).
	// 	From(0).
	// 	SourceIncludes("product")

	agg := esquery.TermsAgg("sku_count", "sku.keyword").Size(1000)

	query := esquery.Search().Aggs(agg).Size(0)

	result, err := query.MarshalJSON()
	if err != nil {
		fmt.Printf("Error parsing the query: %s", err)
	}
	fmt.Printf("Query: %s\n", string(result))

	var response map[string]interface{}

	resInventory, err := query.Run(es,
		es.Search.WithContext(context.TODO()),
		es.Search.WithIndex(index),
	)
	if err != nil {
		fmt.Printf("Error getting response: %s", err)
	}
	defer resInventory.Body.Close()

	if resInventory.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(resInventory.Body).Decode(&e); err != nil {
			fmt.Printf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and error information.
			fmt.Printf("[%s] %s: %s",
				resInventory.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}

	if err := json.NewDecoder(resInventory.Body).Decode(&response); err != nil {
		fmt.Printf("Error parsing the response body: %s", err)
	}

	fmt.Printf(
		"[%s] %d hits; took: %dms\n",
		resInventory.Status(),
		int(response["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
		int(response["took"].(float64)),
	)

	t := response["aggregations"].(map[string]interface{})["sku_count"].(map[string]interface{})["buckets"].([]interface{})

	resCountSku := make([]*dto.SkuCount, 0)

	for _, v := range t {
		skuCount := v.(map[string]interface{})
		resCountSku = append(resCountSku, &dto.SkuCount{
			Sku:   skuCount["key"].(string),
			Count: skuCount["doc_count"].(float64),
		})
	}
	utils.PrettyPrint(resCountSku)
	// utils.PrettyPrint(t)

	// Print the ID and document source for each hit.
	// for _, hit := range response["hits"].(map[string]interface{})["hits"].([]interface{}) {
	// 	fmt.Printf(" * ID=%s, %s", hit.(map[string]interface{})["_id"], hit.(map[string]interface{})["_source"])
	// }

	return nil, nil
}