package query

import (
	"fmt"

	"github.com/aquasecurity/esquery"
)

const ES_QUERY_MAX_SIZE = 1000 // Reference: https://www.elastic.co/guide/en/app-search/7.17/limits.html
const KEY_PRODUCT_ID = "product_id"
const KEY_BRAND_ID = "brand_id"
const KEY_WAREHOUSE_ID = "warehouse_id"

func QueryListProduct(skus []float64, printQuery bool) (*esquery.SearchRequest, error) {

	terms := make([]esquery.Mappable, 0)
	if len(skus) > 0 {
		for _, s := range skus {
			terms = append(terms, esquery.Term("product.product_sku", s))
		}
	} else {
		terms = append(terms, esquery.MatchAll())
	}

	query := esquery.Search().
		Query(
			esquery.Bool().Should(terms...),
		).
		SourceIncludes("product").
		Size(ES_QUERY_MAX_SIZE)

	if printQuery {
		result, err := query.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("Error.Parsing: %s", err)
		}
		fmt.Printf("Query: %s\n", string(result))
	}

	return query, nil
}

func QueryList(printQuery bool, extend []esquery.Mappable) (*esquery.SearchRequest, error) {

	query := esquery.Search().
		Query(
			esquery.Bool().Must(extend...),
		).
		Size(ES_QUERY_MAX_SIZE)

	if printQuery {
		result, err := query.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("Error.Parsing: %s", err)
		}
		fmt.Printf("Query: %s\n", string(result))
	}

	return query, nil
}

func QueryAggSku(key string, printQuery bool, status_ids []int, product_status_ids []int, extend []esquery.Mappable) (*esquery.SearchRequest, error) {
	agg := esquery.TermsAgg(key, "sku.keyword").
		Aggs(esquery.TermsAgg(KEY_WAREHOUSE_ID, KEY_WAREHOUSE_ID)).
		Size(ES_QUERY_MAX_SIZE)

	match := make([]esquery.Mappable, 0)
	if len(status_ids) > 0 {
		for _, v := range status_ids {
			match = append(match, esquery.Term("status_id", v))
		}
	}
	if len(product_status_ids) > 0 {
		for _, v := range product_status_ids {
			match = append(match, esquery.Term("product_status_id", v))
		}
	}
	if len(match) == 0 {
		match = append(match, esquery.MatchAll())
	}

	query := esquery.Search().
		Query(
			esquery.Bool().
				Must(extend...).
				Should(match...),
		).
		Aggs(agg).
		Size(0)
	if printQuery {
		result, err := query.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("Error.Parsing: %s", err)
		}
		fmt.Printf("Query: %s\n", string(result))
	}
	return query, nil
}
