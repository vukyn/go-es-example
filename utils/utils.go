package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/elastic/go-elasticsearch/v7/esapi"
)

func PrettyPrint(v interface{}) (err error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}


// Write text to a file.
// Create if not exist, or overwrite the existing file.
//
// Example:
//
//	WriteFile("Hello World", "temp/output.txt")
func WriteFile(input string, filePath string) error {
	data := []byte(input)
	dir, _ := filepath.Split(filePath)

	if _, err := os.Stat(dir); err == nil {
		os.Remove(filePath)
	} else {
		if err := os.Mkdir(dir, os.ModePerm); err != nil {
			return err
		}
	}
	if err := os.WriteFile(filePath, data, 0); err != nil {
		return err
	}
	return nil
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

func GetHitsResponse(esRes *esapi.Response) []interface{} {
	var response map[string]interface{}

	if esRes.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(esRes.Body).Decode(&e); err != nil {
			fmt.Printf("GetHitsResponse.IsError: Error parsing the response body: %s", err)
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
		fmt.Printf("GetHitsResponse: Error parsing the response body: %s", err)
	} else {
		fmt.Printf(
			// Print the response status and information.
			"[%s] %d hits; took: %dms\n",
			esRes.Status(),
			int(response["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64)),
			int(response["took"].(float64)),
		)
	}

	// PrettyPrint(response)
	return response["hits"].(map[string]interface{})["hits"].([]interface{})
}