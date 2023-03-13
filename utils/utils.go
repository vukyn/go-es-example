package utils

import (
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v7/esapi"
)

func PrettyPrint(v interface{}) (err error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
	return
}

func GetAggregationResponse(esRes *esapi.Response, key string) []interface{} {
	var response map[string]interface{}

	if esRes.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(esRes.Body).Decode(&e); err != nil {
			fmt.Printf("GetAggregationResponse.IsError: Error parsing the response body: %s", err)
		} else {
			// Print the response status and error information.
			fmt.Printf("[%s] %s: %s",
				esRes.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}

	if err := json.NewDecoder(esRes.Body).Decode(&response); err != nil {
		fmt.Printf("GetAggregationResponse: Error parsing the response body: %s", err)
	} else {
		fmt.Printf(
			// Print the response status and information.
			"[%s] %s: %d hits; took: %dms\n",
			esRes.Status(),
			key,
			int(response["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
			int(response["took"].(float64)),
		)
	}

	// PrettyPrint(response)
	return response["aggregations"].(map[string]interface{})[key].(map[string]interface{})["buckets"].([]interface{})
}
