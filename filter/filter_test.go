package filter

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestEqualityFilter(t *testing.T) {
	f := Eq("status", "active")
	expected := bson.M{"status": "active"}

	if !equalBSON(f.Build(), expected) {
		t.Errorf("Expected %v, got %v", expected, f.Build())
	}
}

func TestComparisonFilters(t *testing.T) {
	tests := []struct {
		name     string
		filter   *Builder
		expected bson.M
	}{
		{
			name:     "Greater than",
			filter:   Gt("age", 21),
			expected: bson.M{"age": bson.M{"$gt": 21}},
		},
		{
			name:     "Less than or equal",
			filter:   Lte("price", 100.0),
			expected: bson.M{"price": bson.M{"$lte": 100.0}},
		},
		{
			name:     "Not equal",
			filter:   Ne("status", "inactive"),
			expected: bson.M{"status": bson.M{"$ne": "inactive"}},
		},
		{
			name:     "In array",
			filter:   In("tags", "go", "mongodb", "database"),
			expected: bson.M{"tags": bson.M{"$in": []any{"go", "mongodb", "database"}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !equalBSON(tt.filter.Build(), tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, tt.filter.Build())
			}
		})
	}
}

func TestLogicalOperators(t *testing.T) {
	// Test AND operation
	f := And(
		Eq("status", "active"),
		Gt("age", 21),
	)

	expected := bson.M{
		"$and": []bson.M{
			{"status": "active"},
			{"age": bson.M{"$gt": 21}},
		},
	}

	if !equalBSON(f.Build(), expected) {
		t.Errorf("AND filter: Expected %v, got %v", expected, f.Build())
	}

	// Test OR operation
	f2 := Or(
		Eq("category", "electronics"),
		Eq("category", "books"),
	)

	expected2 := bson.M{
		"$or": []bson.M{
			{"category": "electronics"},
			{"category": "books"},
		},
	}

	if !equalBSON(f2.Build(), expected2) {
		t.Errorf("OR filter: Expected %v, got %v", expected2, f2.Build())
	}
}

func TestComplexFilter(t *testing.T) {
	// Test complex nested filter
	f := And(
		Eq("category", "electronics"),
		Or(
			Gt("price", 100),
			Eq("brand", "premium"),
		),
		In("tags", "new", "featured"),
	)

	expected := bson.M{
		"$and": []bson.M{
			{"category": "electronics"},
			{
				"$or": []bson.M{
					{"price": bson.M{"$gt": 100}},
					{"brand": "premium"},
				},
			},
			{"tags": bson.M{"$in": []any{"new", "featured"}}},
		},
	}

	if !equalBSON(f.Build(), expected) {
		t.Errorf("Complex filter: Expected %v, got %v", expected, f.Build())
	}
}

func TestArrayOperators(t *testing.T) {
	// Test ElemMatch
	f := ElemMatch("orders", And(
		Gt("amount", 1000),
		Eq("status", "completed"),
	))

	expected := bson.M{
		"orders": bson.M{
			"$elemMatch": bson.M{
				"$and": []bson.M{
					{"amount": bson.M{"$gt": 1000}},
					{"status": "completed"},
				},
			},
		},
	}

	if !equalBSON(f.Build(), expected) {
		t.Errorf("ElemMatch filter: Expected %v, got %v", expected, f.Build())
	}

	// Test Size
	f2 := Size("reviews", 5)
	expected2 := bson.M{"reviews": bson.M{"$size": 5}}

	if !equalBSON(f2.Build(), expected2) {
		t.Errorf("Size filter: Expected %v, got %v", expected2, f2.Build())
	}
}

func TestStringOperators(t *testing.T) {
	// Test Regex
	f := Regex("name", "^John", "i")
	expected := bson.M{
		"name": bson.M{
			"$regex":   "^John",
			"$options": "i",
		},
	}

	if !equalBSON(f.Build(), expected) {
		t.Errorf("Regex filter: Expected %v, got %v", expected, f.Build())
	}

	// Test Text search
	f2 := Text("mongodb go driver")
	expected2 := bson.M{
		"$text": bson.M{
			"$search": "mongodb go driver",
		},
	}

	if !equalBSON(f2.Build(), expected2) {
		t.Errorf("Text filter: Expected %v, got %v", expected2, f2.Build())
	}
}

func TestExistenceOperators(t *testing.T) {
	// Test Exists
	f := Exists("email", true)
	expected := bson.M{"email": bson.M{"$exists": true}}

	if !equalBSON(f.Build(), expected) {
		t.Errorf("Exists filter: Expected %v, got %v", expected, f.Build())
	}

	// Test Type
	f2 := Type("created_at", BSONTypeDateTime)
	expected2 := bson.M{"created_at": bson.M{"$type": 9}}

	if !equalBSON(f2.Build(), expected2) {
		t.Errorf("Type filter: Expected %v, got %v", expected2, f2.Build())
	}
}

// Helper function to compare BSON documents
func equalBSON(a, b bson.M) bool {
	// Use deep equality check for robust comparison
	return reflect.DeepEqual(a, b)
}
