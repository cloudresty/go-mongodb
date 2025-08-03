package mongodb

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestBSONConvenienceFunctions(t *testing.T) {
	t.Run("SortAsc", func(t *testing.T) {
		result := SortAsc("name")
		expected := bson.D{{Key: "name", Value: 1}}
		if !equalBSOND(expected, result) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("SortDesc", func(t *testing.T) {
		result := SortDesc("created_at")
		expected := bson.D{{Key: "created_at", Value: -1}}
		if !equalBSOND(expected, result) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("SortMultiple", func(t *testing.T) {
		input := map[string]int{
			"name": 1,
			"age":  -1,
		}
		result := SortMultiple(input)

		// Since maps are unordered, check that both fields are present
		if len(result) != 2 {
			t.Errorf("Expected 2 fields, got %d", len(result))
		}

		nameFound := false
		ageFound := false
		for _, elem := range result {
			if elem.Key == "name" && elem.Value == 1 {
				nameFound = true
			}
			if elem.Key == "age" && elem.Value == -1 {
				ageFound = true
			}
		}
		if !nameFound {
			t.Error("name field not found")
		}
		if !ageFound {
			t.Error("age field not found")
		}
	})

	t.Run("SortMultipleOrdered", func(t *testing.T) {
		result := SortMultipleOrdered("created_at", -1, "name", 1, "age", -1)
		expected := bson.D{
			{Key: "created_at", Value: -1},
			{Key: "name", Value: 1},
			{Key: "age", Value: -1},
		}
		if !equalBSOND(expected, result) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("Document", func(t *testing.T) {
		result := Document("name", "John", "age", 30, "active", true)
		expected := bson.M{
			"name":   "John",
			"age":    30,
			"active": true,
		}
		if !equalBSONM(expected, result) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("Include", func(t *testing.T) {
		result := Include("name", "email", "age")
		expected := ProjectionSpec{
			{Key: "name", Value: 1},
			{Key: "email", Value: 1},
			{Key: "age", Value: 1},
		}
		if !equalProjectionSpec(expected, result) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("Exclude", func(t *testing.T) {
		result := Exclude("_id", "password")
		expected := ProjectionSpec{
			{Key: "_id", Value: 0},
			{Key: "password", Value: 0},
		}
		if !equalProjectionSpec(expected, result) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("Projection", func(t *testing.T) {
		result := Projection(Include("name", "email"), Exclude("_id"))
		expected := bson.D{
			{Key: "name", Value: 1},
			{Key: "email", Value: 1},
			{Key: "_id", Value: 0},
		}
		if !equalBSOND(expected, result) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}

func TestConvertSortSpec(t *testing.T) {
	t.Run("bson.D", func(t *testing.T) {
		input := bson.D{{Key: "name", Value: 1}}
		result := convertSortSpec(input)
		if !equalBSOND(input, result) {
			t.Errorf("Expected %v, got %v", input, result)
		}
	})

	t.Run("map[string]int", func(t *testing.T) {
		input := map[string]int{"name": 1, "age": -1}
		result := convertSortSpec(input)
		if len(result) != 2 {
			t.Errorf("Expected 2 fields, got %d", len(result))
		}

		// Check that both fields are present
		nameFound := false
		ageFound := false
		for _, elem := range result {
			if elem.Key == "name" && elem.Value == 1 {
				nameFound = true
			}
			if elem.Key == "age" && elem.Value == -1 {
				ageFound = true
			}
		}
		if !nameFound {
			t.Error("name field not found")
		}
		if !ageFound {
			t.Error("age field not found")
		}
	})

	t.Run("nil", func(t *testing.T) {
		result := convertSortSpec(nil)
		expected := bson.D{}
		if !equalBSOND(expected, result) {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	t.Run("invalid_type_panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for invalid type")
			}
		}()
		convertSortSpec("invalid")
	})
}

// Helper functions for comparison
func equalBSOND(a, b bson.D) bool {
	if len(a) != len(b) {
		return false
	}
	for i, elem := range a {
		if elem.Key != b[i].Key || elem.Value != b[i].Value {
			return false
		}
	}
	return true
}

func equalBSONM(a, b bson.M) bool {
	if len(a) != len(b) {
		return false
	}
	for key, value := range a {
		if bValue, exists := b[key]; !exists || value != bValue {
			return false
		}
	}
	return true
}

func equalProjectionSpec(a, b ProjectionSpec) bool {
	return equalBSOND(bson.D(a), bson.D(b))
}
