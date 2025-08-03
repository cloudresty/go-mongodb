package mongodb

import "go.mongodb.org/mongo-driver/v2/bson"

// SortSpec represents a flexible sort specification that can be:
// - bson.D for ordered sorting
// - map[string]int for simple field:order mapping
// - mongodb.D (re-exported bson.D from mongodb.go)
type SortSpec interface{}

// Convenience functions for common BSON operations

// SortAsc creates an ascending sort specification for a single field
func SortAsc(field string) bson.D {
	return bson.D{{Key: field, Value: 1}}
}

// SortDesc creates a descending sort specification for a single field
func SortDesc(field string) bson.D {
	return bson.D{{Key: field, Value: -1}}
}

// SortMultiple creates a sort specification from a map of field:order pairs
// Note: Go maps are unordered, so use this only when sort order doesn't matter
// For ordered sorting, use SortMultipleOrdered or bson.D directly
func SortMultiple(fields map[string]int) bson.D {
	result := make(bson.D, 0, len(fields))
	for field, order := range fields {
		result = append(result, bson.E{Key: field, Value: order})
	}
	return result
}

// SortMultipleOrdered creates an ordered sort specification from field-order pairs
// Usage: SortMultipleOrdered("created_at", -1, "name", 1)
func SortMultipleOrdered(fieldOrderPairs ...any) bson.D {
	if len(fieldOrderPairs)%2 != 0 {
		panic("SortMultipleOrdered requires an even number of arguments (field, order pairs)")
	}

	result := make(bson.D, 0, len(fieldOrderPairs)/2)
	for i := 0; i < len(fieldOrderPairs); i += 2 {
		field, ok := fieldOrderPairs[i].(string)
		if !ok {
			panic("SortMultipleOrdered field must be string")
		}
		order, ok := fieldOrderPairs[i+1].(int)
		if !ok {
			panic("SortMultipleOrdered order must be int")
		}
		result = append(result, bson.E{Key: field, Value: order})
	}
	return result
}

// Document creates a BSON document from key-value pairs
// Usage: Document("name", "John", "age", 30)
func Document(keyValuePairs ...any) bson.M {
	if len(keyValuePairs)%2 != 0 {
		panic("Document requires an even number of arguments (key, value pairs)")
	}

	result := make(bson.M)
	for i := 0; i < len(keyValuePairs); i += 2 {
		key, ok := keyValuePairs[i].(string)
		if !ok {
			panic("Document key must be string")
		}
		result[key] = keyValuePairs[i+1]
	}
	return result
}

// Projection creates a projection document with included/excluded fields
// Usage: Projection(Include("name", "email"), Exclude("_id"))
func Projection(specs ...ProjectionSpec) bson.D {
	result := bson.D{}
	for _, spec := range specs {
		result = append(result, spec...)
	}
	return result
}

// ProjectionSpec represents field inclusion/exclusion in projections
type ProjectionSpec bson.D

// Include creates a projection spec that includes the specified fields
func Include(fields ...string) ProjectionSpec {
	result := make(bson.D, len(fields))
	for i, field := range fields {
		result[i] = bson.E{Key: field, Value: 1}
	}
	return ProjectionSpec(result)
}

// Exclude creates a projection spec that excludes the specified fields
func Exclude(fields ...string) ProjectionSpec {
	result := make(bson.D, len(fields))
	for i, field := range fields {
		result[i] = bson.E{Key: field, Value: 0}
	}
	return ProjectionSpec(result)
}

// convertSortSpec converts a SortSpec to bson.D
func convertSortSpec(sort SortSpec) bson.D {
	switch s := sort.(type) {
	case bson.D:
		return s
	case map[string]int:
		return SortMultiple(s)
	case nil:
		return bson.D{}
	default:
		panic("invalid sort specification: must be bson.D, mongodb.D, or map[string]int")
	}
}
