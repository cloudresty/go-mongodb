package update

import (
	"github.com/cloudresty/go-mongodb/filter"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Builder represents a fluent update builder for MongoDB update operations
type Builder struct {
	update bson.M
}

// New creates a new update builder
func New() *Builder {
	return &Builder{
		update: bson.M{},
	}
}

// Build returns the underlying BSON update document
func (b *Builder) Build() bson.M {
	if len(b.update) == 0 {
		return bson.M{}
	}
	return b.update
}

// ToBSONM converts the update to a bson.M for compatibility
func (b *Builder) ToBSONM() bson.M {
	return b.Build()
}

// Field Update Operators

// Set sets the value of a field
func Set(field string, value any) *Builder {
	return &Builder{
		update: bson.M{"$set": bson.M{field: value}},
	}
}

// Set sets the value of a field (method version)
func (b *Builder) Set(field string, value any) *Builder {
	if b.update["$set"] == nil {
		b.update["$set"] = bson.M{}
	}
	b.update["$set"].(bson.M)[field] = value
	return b
}

// Unset removes the specified fields
func Unset(fields ...string) *Builder {
	unsetDoc := bson.M{}
	for _, field := range fields {
		unsetDoc[field] = ""
	}
	return &Builder{
		update: bson.M{"$unset": unsetDoc},
	}
}

// Unset removes the specified fields (method version)
func (b *Builder) Unset(fields ...string) *Builder {
	if b.update["$unset"] == nil {
		b.update["$unset"] = bson.M{}
	}
	for _, field := range fields {
		b.update["$unset"].(bson.M)[field] = ""
	}
	return b
}

// Inc increments the value of a field by a specified amount
func Inc(field string, value any) *Builder {
	return &Builder{
		update: bson.M{"$inc": bson.M{field: value}},
	}
}

// Inc increments the value of a field by a specified amount (method version)
func (b *Builder) Inc(field string, value any) *Builder {
	if b.update["$inc"] == nil {
		b.update["$inc"] = bson.M{}
	}
	b.update["$inc"].(bson.M)[field] = value
	return b
}

// Mul multiplies the value of a field by a specified amount
func Mul(field string, value any) *Builder {
	return &Builder{
		update: bson.M{"$mul": bson.M{field: value}},
	}
}

// Mul multiplies the value of a field by a specified amount (method version)
func (b *Builder) Mul(field string, value any) *Builder {
	if b.update["$mul"] == nil {
		b.update["$mul"] = bson.M{}
	}
	b.update["$mul"].(bson.M)[field] = value
	return b
}

// Rename renames a field
func Rename(from, to string) *Builder {
	return &Builder{
		update: bson.M{"$rename": bson.M{from: to}},
	}
}

// Rename renames a field (method version)
func (b *Builder) Rename(from, to string) *Builder {
	if b.update["$rename"] == nil {
		b.update["$rename"] = bson.M{}
	}
	b.update["$rename"].(bson.M)[from] = to
	return b
}

// SetOnInsert sets the value of a field only when an upsert inserts a document
func SetOnInsert(field string, value any) *Builder {
	return &Builder{
		update: bson.M{"$setOnInsert": bson.M{field: value}},
	}
}

// SetOnInsert sets the value of a field only when an upsert inserts a document (method version)
func (b *Builder) SetOnInsert(field string, value any) *Builder {
	if b.update["$setOnInsert"] == nil {
		b.update["$setOnInsert"] = bson.M{}
	}
	b.update["$setOnInsert"].(bson.M)[field] = value
	return b
}

// Array Update Operators

// Push appends a value to an array
func Push(field string, value any) *Builder {
	return &Builder{
		update: bson.M{"$push": bson.M{field: value}},
	}
}

// Push appends a value to an array (method version)
func (b *Builder) Push(field string, value any) *Builder {
	if b.update["$push"] == nil {
		b.update["$push"] = bson.M{}
	}
	b.update["$push"].(bson.M)[field] = value
	return b
}

// PushEach appends multiple values to an array
func PushEach(field string, values ...any) *Builder {
	return &Builder{
		update: bson.M{"$push": bson.M{field: bson.M{"$each": values}}},
	}
}

// PushEach appends multiple values to an array (method version)
func (b *Builder) PushEach(field string, values ...any) *Builder {
	if b.update["$push"] == nil {
		b.update["$push"] = bson.M{}
	}
	b.update["$push"].(bson.M)[field] = bson.M{"$each": values}
	return b
}

// Pull removes all instances of a value from an array that match a specified condition
func Pull(field string, filterBuilder *filter.Builder) *Builder {
	var condition any
	if filterBuilder != nil {
		condition = filterBuilder.Build()
	} else {
		condition = bson.M{}
	}

	return &Builder{
		update: bson.M{"$pull": bson.M{field: condition}},
	}
}

// Pull removes all instances of a value from an array that match a specified condition (method version)
func (b *Builder) Pull(field string, filterBuilder *filter.Builder) *Builder {
	if b.update["$pull"] == nil {
		b.update["$pull"] = bson.M{}
	}

	var condition any
	if filterBuilder != nil {
		condition = filterBuilder.Build()
	} else {
		condition = bson.M{}
	}

	b.update["$pull"].(bson.M)[field] = condition
	return b
}

// AddToSet adds a value to an array only if it doesn't already exist
func AddToSet(field string, value any) *Builder {
	return &Builder{
		update: bson.M{"$addToSet": bson.M{field: value}},
	}
}

// AddToSet adds a value to an array only if it doesn't already exist (method version)
func (b *Builder) AddToSet(field string, value any) *Builder {
	if b.update["$addToSet"] == nil {
		b.update["$addToSet"] = bson.M{}
	}
	b.update["$addToSet"].(bson.M)[field] = value
	return b
}

// PopFirst removes the first element from an array
func PopFirst(field string) *Builder {
	return &Builder{
		update: bson.M{"$pop": bson.M{field: -1}},
	}
}

// PopFirst removes the first element from an array (method version)
func (b *Builder) PopFirst(field string) *Builder {
	if b.update["$pop"] == nil {
		b.update["$pop"] = bson.M{}
	}
	b.update["$pop"].(bson.M)[field] = -1
	return b
}

// PopLast removes the last element from an array
func PopLast(field string) *Builder {
	return &Builder{
		update: bson.M{"$pop": bson.M{field: 1}},
	}
}

// PopLast removes the last element from an array (method version)
func (b *Builder) PopLast(field string) *Builder {
	if b.update["$pop"] == nil {
		b.update["$pop"] = bson.M{}
	}
	b.update["$pop"].(bson.M)[field] = 1
	return b
}

// Combining Updates

// And combines multiple update operations
func (b *Builder) And(updates ...*Builder) *Builder {
	for _, update := range updates {
		if update == nil {
			continue
		}

		for operator, fields := range update.update {
			if b.update[operator] == nil {
				b.update[operator] = bson.M{}
			}

			// Merge the fields for this operator
			for field, value := range fields.(bson.M) {
				b.update[operator].(bson.M)[field] = value
			}
		}
	}
	return b
}
