package query

import (
	"fmt"

	"github.com/aquasecurity/esquery"
)

const ES_QUERY_MAX_SIZE = 1000 // Reference: https://www.elastic.co/guide/en/app-search/7.17/limits.html
const KEY_PRODUCT_ID = "product_id"
const KEY_BRAND_ID = "brand_id"
const KEY_WAREHOUSE_ID = "warehouse_id"

func QueryAggAllSku(key string) (*esquery.SearchRequest, error) {
	agg := esquery.TermsAgg(key, "sku.keyword").
		Aggs(esquery.
			TermsAgg(KEY_PRODUCT_ID, KEY_PRODUCT_ID).
			Aggs(esquery.
				TermsAgg(KEY_BRAND_ID, KEY_BRAND_ID).
				Aggs(esquery.
					TermsAgg(KEY_WAREHOUSE_ID, KEY_WAREHOUSE_ID)))).
		Size(ES_QUERY_MAX_SIZE)

	query := esquery.Search().Aggs(agg).Size(0)
	result, err := query.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("Error.Parsing: %s", err)
	}
	fmt.Printf("Query: %s\n", string(result))
	return query, nil
}

func QueryAggInStock(key string) (*esquery.SearchRequest, error) {
	agg := esquery.TermsAgg(key, "sku.keyword").
		Aggs(esquery.TermsAgg(KEY_WAREHOUSE_ID, KEY_WAREHOUSE_ID)).
		Size(ES_QUERY_MAX_SIZE)

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

func QueryAggCommitted(key string) (*esquery.SearchRequest, error) {
	agg := esquery.TermsAgg(key, "sku.keyword").
		Aggs(esquery.TermsAgg(KEY_WAREHOUSE_ID, KEY_WAREHOUSE_ID)).
		Size(ES_QUERY_MAX_SIZE)

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

func QueryAggReceiving(key string) (*esquery.SearchRequest, error) {
	agg := esquery.TermsAgg(key, "sku.keyword").
		Aggs(esquery.TermsAgg(KEY_WAREHOUSE_ID, KEY_WAREHOUSE_ID)).
		Size(ES_QUERY_MAX_SIZE)

	query := esquery.Search().
		Query(
			esquery.Bool().Should(
				esquery.Match("status_id", 2),
				esquery.Match("status_id", 26),
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

// func BuildQueryCountSku([]*int status_id) *esquery.SearchRequest {

// }
