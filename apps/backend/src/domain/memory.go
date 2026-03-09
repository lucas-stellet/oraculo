// apps/backend/src/domain/memory.go
package domain

import "time"

// Category classifies a knowledge finding.
type Category string

const (
	CategoryPattern      Category = "pattern"
	CategoryConvention   Category = "convention"
	CategoryConstraint   Category = "constraint"
	CategoryDependency   Category = "dependency"
	CategoryTest         Category = "test"
	CategoryArchitecture Category = "architecture"
)

var validCategories = map[Category]bool{
	CategoryPattern: true, CategoryConvention: true, CategoryConstraint: true,
	CategoryDependency: true, CategoryTest: true, CategoryArchitecture: true,
}

// Valid reports whether c is a recognized category.
func (c Category) Valid() bool { return validCategories[c] }

// Confidence indicates how certain a knowledge finding is.
type Confidence string

const (
	ConfidenceHigh   Confidence = "high"
	ConfidenceMedium Confidence = "medium"
	ConfidenceLow    Confidence = "low"
)

var validConfidences = map[Confidence]bool{
	ConfidenceHigh: true, ConfidenceMedium: true, ConfidenceLow: true,
}

// Valid reports whether c is a recognized confidence level.
func (c Confidence) Valid() bool { return validConfidences[c] }

// Knowledge represents a codebase finding persisted in the knowledge table.
type Knowledge struct {
	ID          int
	Domain      string
	Category    Category
	Finding     string
	SourceFiles string
	Confidence  Confidence
	CreatedAt   time.Time
}
