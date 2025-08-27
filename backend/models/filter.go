// models/filter.go
package models

// FilterRow Svelte uygulamasından gelen tek bir filtre koşulunu temsil eder.
type FilterRow struct {
	Field    string `json:"field"`    // Örn: "person_id", "activity_code"
	Operator string `json:"operator"` // Örn: "=", "!=", ">", "<", "LIKE"
	Value    string `json:"value"`    // Örn: "109403", "FLT", "2025-07-01"
}

// SavedFilter Svelte uygulamasından gelen kaydedilmiş bir filtrenin tamamını temsil eder.
type SavedFilter struct {
	ID    string      `json:"id"`
	Name  string      `json:"name"`
	Rows  []FilterRow `json:"rows"`
	Logic string      `json:"logic"` // "AND" veya "OR"
}
