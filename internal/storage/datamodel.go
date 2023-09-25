package storage

type MetricsDB struct {
	Name string `json:"name" db:"name"`
	Metric
}

type Metric struct {
	Type string `json:"type" db:"type"`
	Val  any    `json:"value" db:"value"`
}
