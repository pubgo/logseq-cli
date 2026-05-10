package logseq

import (
	"testing"
)

func TestLimitRegexp(t *testing.T) {
	tests := []struct {
		query string
		limit int
	}{
		{"[:find ?name :where [?p :block/name ?name] :limit 5]", 5},
		{"[:find ?name :where [?p :block/name ?name] :limit 100]", 100},
		{"[:find ?name :where [?p :block/name ?name]]", 0},
		{"[:find ?x :where [?b :block/content ?x] :limit 20]", 20},
	}
	for _, tt := range tests {
		m := limitRegexp.FindStringSubmatch(tt.query)
		got := 0
		if len(m) == 2 {
			got = mustAtoi(m[1])
		}
		if got != tt.limit {
			t.Errorf("query=%q: limit=%d, want %d", tt.query, got, tt.limit)
		}
	}
}

func TestLimitRegexp_Stripping(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			"[:find ?name :where [?p :block/name ?name] :limit 5]",
			"[:find ?name :where [?p :block/name ?name] ]",
		},
		{
			"[:find ?name :where [?p :block/name ?name]]",
			"[:find ?name :where [?p :block/name ?name]]",
		},
	}
	for _, tt := range tests {
		got := limitRegexp.ReplaceAllString(tt.input, "")
		if got != tt.want {
			t.Errorf("strip limit: got=%q, want=%q", got, tt.want)
		}
	}
}

func TestOrderByRegexp_Stripping(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			"[:find ?x ?u :where [?b :block/content ?x] [?b :block/updated-at ?u] :order-by [(desc ?u)] :limit 5]",
			"[:find ?x ?u :where [?b :block/content ?x] [?b :block/updated-at ?u]  :limit 5]",
		},
		{
			"[:find ?name :where [?p :block/name ?name]]",
			"[:find ?name :where [?p :block/name ?name]]",
		},
	}
	for _, tt := range tests {
		got := orderByRegexp.ReplaceAllString(tt.input, "")
		if got != tt.want {
			t.Errorf("strip order-by: got=%q, want=%q", got, tt.want)
		}
	}
}

func mustAtoi(s string) int {
	n := 0
	for _, c := range s {
		n = n*10 + int(c-'0')
	}
	return n
}
