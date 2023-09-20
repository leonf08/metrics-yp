package storage

type dbEntry struct {
	Name string
	Metric
}

type Metric struct {
	Type string `json:"type" db:"type"`
	Val  any    `json:"value" db:"value"`
}
