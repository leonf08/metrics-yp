package models

// MetricJSON is data structure for JSON request and response.
type MetricJSON struct {
	// ID is a name of the metric
	ID string `json:"id"`

	// MType is a type of the metric
	MType string `json:"type"`

	// Delta is a value of the metric in case of counter type
	Delta *int64 `json:"delta,omitempty"`

	// Value is a value of the metric in case of gauge type
	Value *float64 `json:"value,omitempty"`
}

// MetricDB is data structure for database.
type MetricDB struct {
	// Name is a name of the metric
	Name string `json:"name" db:"name"`
	Metric
}

// Metric is data structure for internal usage.
type Metric struct {
	// Type is a type of the metric
	Type string `json:"type" db:"type"`

	// Val is a value of the metric
	Val any `json:"value" db:"value"`
}
