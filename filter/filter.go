package filter

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Builder represents a fluent filter builder for MongoDB queries
type Builder struct {
	filter bson.M
}

// New creates a new filter builder
func New() *Builder {
	return &Builder{
		filter: bson.M{},
	}
}

// Build returns the underlying BSON filter document
func (b *Builder) Build() bson.M {
	if len(b.filter) == 0 {
		return bson.M{}
	}
	return b.filter
}

// ToBSONM converts the filter to a bson.M for compatibility
func (b *Builder) ToBSONM() bson.M {
	return b.Build()
}

// Comparison Operators

// Eq creates an equality filter
func Eq(field string, value any) *Builder {
	return &Builder{
		filter: bson.M{field: value},
	}
}

// Ne creates a not-equal filter
func Ne(field string, value any) *Builder {
	return &Builder{
		filter: bson.M{field: bson.M{"$ne": value}},
	}
}

// Gt creates a greater-than filter
func Gt(field string, value any) *Builder {
	return &Builder{
		filter: bson.M{field: bson.M{"$gt": value}},
	}
}

// Gte creates a greater-than-or-equal filter
func Gte(field string, value any) *Builder {
	return &Builder{
		filter: bson.M{field: bson.M{"$gte": value}},
	}
}

// Lt creates a less-than filter
func Lt(field string, value any) *Builder {
	return &Builder{
		filter: bson.M{field: bson.M{"$lt": value}},
	}
}

// Lte creates a less-than-or-equal filter
func Lte(field string, value any) *Builder {
	return &Builder{
		filter: bson.M{field: bson.M{"$lte": value}},
	}
}

// In creates an "in" filter for array membership
func In(field string, values ...any) *Builder {
	return &Builder{
		filter: bson.M{field: bson.M{"$in": values}},
	}
}

// Nin creates a "not in" filter for array non-membership
func Nin(field string, values ...any) *Builder {
	return &Builder{
		filter: bson.M{field: bson.M{"$nin": values}},
	}
}

// Logical Operators

// And combines multiple filters with logical AND
func (b *Builder) And(filters ...*Builder) *Builder {
	if len(filters) == 0 {
		return b
	}

	conditions := make([]bson.M, 0, len(filters)+1)

	// Add current filter if it exists
	if len(b.filter) > 0 {
		conditions = append(conditions, b.filter)
	}

	// Add all provided filters
	for _, filter := range filters {
		if filter != nil && len(filter.filter) > 0 {
			conditions = append(conditions, filter.filter)
		}
	}

	if len(conditions) == 0 {
		return &Builder{filter: bson.M{}}
	}

	if len(conditions) == 1 {
		return &Builder{filter: conditions[0]}
	}

	return &Builder{
		filter: bson.M{"$and": conditions},
	}
}

// Or combines multiple filters with logical OR
func (b *Builder) Or(filters ...*Builder) *Builder {
	if len(filters) == 0 {
		return b
	}

	conditions := make([]bson.M, 0, len(filters)+1)

	// Add current filter if it exists
	if len(b.filter) > 0 {
		conditions = append(conditions, b.filter)
	}

	// Add all provided filters
	for _, filter := range filters {
		if filter != nil && len(filter.filter) > 0 {
			conditions = append(conditions, filter.filter)
		}
	}

	if len(conditions) == 0 {
		return &Builder{filter: bson.M{}}
	}

	if len(conditions) == 1 {
		return &Builder{filter: conditions[0]}
	}

	return &Builder{
		filter: bson.M{"$or": conditions},
	}
}

// Not negates the current filter
func (b *Builder) Not() *Builder {
	if len(b.filter) == 0 {
		return b
	}

	return &Builder{
		filter: bson.M{"$not": b.filter},
	}
}

// Convenience function for And that can be called statically
func And(filters ...*Builder) *Builder {
	return New().And(filters...)
}

// Convenience function for Or that can be called statically
func Or(filters ...*Builder) *Builder {
	return New().Or(filters...)
}

// Array Operators

// ElemMatch creates an element match filter for arrays
func ElemMatch(field string, filter *Builder) *Builder {
	if filter == nil {
		return &Builder{filter: bson.M{}}
	}

	return &Builder{
		filter: bson.M{field: bson.M{"$elemMatch": filter.filter}},
	}
}

// Size creates a filter for array size
func Size(field string, size int) *Builder {
	return &Builder{
		filter: bson.M{field: bson.M{"$size": size}},
	}
}

// All creates a filter that matches arrays containing all specified values
func All(field string, values ...any) *Builder {
	return &Builder{
		filter: bson.M{field: bson.M{"$all": values}},
	}
}

// String Operators

// Regex creates a regular expression filter
func Regex(field string, pattern string, options ...string) *Builder {
	regexFilter := bson.M{"$regex": pattern}

	if len(options) > 0 {
		regexFilter["$options"] = options[0]
	}

	return &Builder{
		filter: bson.M{field: regexFilter},
	}
}

// Text creates a text search filter
func Text(query string) *Builder {
	return &Builder{
		filter: bson.M{"$text": bson.M{"$search": query}},
	}
}

// Existence Operators

// Exists creates a filter for field existence
func Exists(field string, exists bool) *Builder {
	return &Builder{
		filter: bson.M{field: bson.M{"$exists": exists}},
	}
}

// BSONType represents BSON type constants for type checking
type BSONType int

const (
	BSONTypeDouble     BSONType = 1
	BSONTypeString     BSONType = 2
	BSONTypeObject     BSONType = 3
	BSONTypeArray      BSONType = 4
	BSONTypeBinary     BSONType = 5
	BSONTypeObjectID   BSONType = 7
	BSONTypeBoolean    BSONType = 8
	BSONTypeDateTime   BSONType = 9
	BSONTypeNull       BSONType = 10
	BSONTypeRegex      BSONType = 11
	BSONTypeJavaScript BSONType = 13
	BSONTypeInt32      BSONType = 16
	BSONTypeTimestamp  BSONType = 17
	BSONTypeInt64      BSONType = 18
	BSONTypeDecimal128 BSONType = 19
)

// Type creates a filter for BSON type checking
func Type(field string, bsonType BSONType) *Builder {
	return &Builder{
		filter: bson.M{field: bson.M{"$type": int(bsonType)}},
	}
}
