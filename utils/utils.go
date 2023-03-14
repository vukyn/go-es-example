package utils

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/elastic/go-elasticsearch/v7/esapi"
)

func PrettyPrint(v interface{}) (err error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
	return
}

func WriteFile(text string, filename string) error {
	data := []byte(text)
	folderName := "data/"
	filename = folderName + filename + ".txt"

	if _, err := os.Stat(folderName); err == nil {
		os.Remove(filename)
	} else {
		if err := os.Mkdir(folderName, os.ModePerm); err != nil {
			return fmt.Errorf("error when make dir: %s", err.Error())
		}
	}

	if err := os.WriteFile(filename, data, 0); err != nil {
		return fmt.Errorf("error when write file: %s", err.Error())
	}

	fmt.Printf("write file \"%s\" done!\n", filename)
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