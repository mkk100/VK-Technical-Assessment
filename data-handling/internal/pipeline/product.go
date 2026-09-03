package pipeline

// Product is the normalized representation every source is mapped onto.
type Product struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Source   string  `json:"source"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
}

// BaseURL is the mock API root. Overridable by the caller before fetching.
var BaseURL = "http://localhost:8080"
