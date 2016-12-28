package sp

import (
	"encoding/json"
	"fmt"

	"github.com/bitly/go-simplejson"
)

// ESAgg enum
type ESAgg int

// These are a comprehensive list of es aggregations.
const (
	IllegalAgg ESAgg = iota

	metricBegin
	//metric aggregations method
	Avg
	Cardinality
	ExtendedStats
	GeoBounds
	GeoCentroid
	Max
	Min
	Percentiles
	PercentileRanks
	Stats
	Sum
	Top
	ValueCount

	metricEnd

	bucketBegin
	//bucket aggregations method
	DateHistogram
	DateRange
	Filter
	Filters
	GeoDistance
	GeoHashGrid
	Global
	Histogram
	IPRange
	Missing
	Nested
	Range
	ReverseNested
	Sampler
	SignificantTerms
	Terms

	bucketEnd
)

var aggs = [...]string{
	IllegalAgg: "ILLEGAL",

	Avg:             "avg",
	Cardinality:     "cardinality",
	ExtendedStats:   "extended_stats",
	GeoBounds:       "geo_bounds",
	GeoCentroid:     "geo_centroid",
	Max:             "max",
	Min:             "min",
	Percentiles:     "percentiles",
	PercentileRanks: "percentile_ranks",
	Stats:           "stats",
	Sum:             "sum",
	Top:             "top",
	ValueCount:      "value_count",

	DateHistogram:    "date_histogram",
	DateRange:        "date_range",
	Filter:           "date_range",
	Filters:          "filters",
	GeoDistance:      "geo_distance",
	GeoHashGrid:      "geohash_grid",
	Global:           "global",
	Histogram:        "histogram",
	IPRange:          "ip_range",
	Missing:          "missing",
	Nested:           "nested",
	Range:            "range",
	ReverseNested:    "reverse_nested",
	Sampler:          "sampler",
	SignificantTerms: "significant_terms",
	Terms:            "terms",
}

type Agg struct {
	name   string
	typ    ESAgg
	params map[string]interface{}
}

type Aggs []*Agg

//EsDsl return dsl json string
func (s *SelectStatement) EsDsl() string {

	js := simplejson.New()
	//from and size
	js.Set("from", s.Offset)
	js.Set("size", s.Limit)

	//sort
	var sort []map[string]string
	for _, sf := range s.SortFields {
		m := make(map[string]string)
		if sf.Ascending {
			m[sf.Name] = "asc"
		} else {
			m[sf.Name] = "desc"
		}
		sort = append(sort, m)
	}
	js.Set("sort", sort)

	//fields
	//scirpt fields

	//query
	if s.Condition != nil {
		branch := []string{"query", "bool", "filter", "script", "script"}
		js.SetPath(branch, s.Condition.String())
	}

	// build aggregates
	path := []string{"aggs"}
	//bucket aggregates
	baggs := s.bucketAggregates()
	for _, a := range baggs {
		_path := append(path, []string{a.name, aggs[a.typ]}...)
		js.SetPath(_path, a.params)

		if a.typ == Terms {
			path = append(path, a.name)
		}
		path = append(path, "aggs")
	}
	//metric aggregates
	maggs := s.metricAggs()
	for _, a := range maggs {
		_path := append(path, []string{a.name, aggs[a.typ]}...)
		js.SetPath(_path, a.params)
	}

	_s, err := js.MarshalJSON()
	if err != nil {
		return ""
	}
	t, _ := json.MarshalIndent(js.MustMap(), "", "  ")
	fmt.Println(string(t))

	return string(_s)
}

func (s *SelectStatement) bucketAggregates() Aggs {
	var aggs Aggs
	for _, dim := range s.Dimensions {
		agg := &Agg{}
		agg.params = make(map[string]interface{})
		agg.name = dim.String()

		switch expr := dim.Expr.(type) {
		case *Call:
			fn := expr.Name
			switch fn {
			case "date_histogram":

				agg.typ = DateHistogram
				agg.params["field"] = expr.Args[0].String()
				agg.params["interval"] = expr.Args[1].String()
			default:
				panic(fmt.Errorf("not support bucket aggregation"))
			}

		default:
			agg.typ = Terms
			agg.params["script"] = expr.String()
		}
		aggs = append(aggs, agg)
	}

	return aggs
}

func (s *SelectStatement) metricAggs() Aggs {
	var aggs Aggs
	for _, field := range s.Fields {
		fn, ok := field.Expr.(*Call)
		if !ok {
			continue
		}
		agg := &Agg{}
		agg.params = make(map[string]interface{})
		agg.name = fn.String()

		switch fn.Name {
		case "avg":
			agg.typ = Avg
			agg.params["script"] = fn.Args[0].String()
		case "cardinality":
			agg.typ = Cardinality
			agg.params["script"] = fn.Args[0].String()
		case "sum":
			agg.typ = Sum
			agg.params["script"] = fn.Args[0].String()
		case "max":
			agg.typ = Max
			agg.params["script"] = fn.Args[0].String()
		case "min":
			agg.typ = Min
			agg.params["script"] = fn.Args[0].String()
		case "top":
			agg.typ = Top
			agg.params["script"] = fn.Args[0].String()
		case "count":
			agg.typ = ValueCount
			agg.params["script"] = fn.Args[0].String()
		case "stats":
			agg.typ = Stats
			agg.params["script"] = fn.Args[0].String()
		case "extended_stats":
			agg.typ = ExtendedStats
			agg.params["script"] = fn.Args[0].String()
		default:
			panic(fmt.Errorf("not support agg aggregation"))
		}

		aggs = append(aggs, agg)
	}

	return aggs
}