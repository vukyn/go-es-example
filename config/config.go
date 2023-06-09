package config

import (
	"errors"
	"os"

	"github.com/elastic/go-elasticsearch/v7"
)

func GetConfig() elasticsearch.Config {

	const (
		TEMP   = "config/es-cert.crt"
	)

	cert, err := os.ReadFile(TEMP)
	if err != nil {
		panic(err)
	}

	cfg := elasticsearch.Config{
		Addresses: []string{
			"https://elk-dev:9200",
		},
		Username: "devuser",
		Password: "a2WSzAYDvgF3cZbm",
		CACert:   cert,
	}

	return cfg
}

func CheckHealth(es *elasticsearch.Client) (string, error) {
	res, err := es.Info()
	if err != nil {
		return "", err
	}

	if res.IsError() {
		return "", errors.New(res.String())
	}

	defer res.Body.Close()
	return res.Status(), nil
}
