package main

import (
	"encoding/json"
	"go_es_example/application/common"
	"go_es_example/application/dto"
	"go_es_example/config"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
)

func main() {
	cfg := config.GetConfig()

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		panic(err)
	}

	if es != nil {
		log.Println("Yes, you're ready to go ...")
	}

	health, err := config.CheckHealth(es)
	if err != nil {
		panic(err)
	}
	log.Printf("Status: %s\n", health)

	res, err := common.GetAllIndexes(es)
	if err != nil {
		panic(err)
	}

	var indices []dto.EsIndice
	err = json.Unmarshal([]byte(strings.Replace(res, "[200 OK] ", "", 1)), &indices)
	if err != nil {
		panic(err)
	}
	common.SearchInventoryUseCase(es, "qc_fulfillment.qc_fulfillment_inventory.wms_inventory")
	// common.SearchInventoryUseCase(es, "fulfillment.qc.cdc.final_product")
}
