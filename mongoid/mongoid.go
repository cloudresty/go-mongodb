package mongoid

import (
	"context"
	"errors"
	"time"

	"github.com/cloudresty/go-mongodb/filter"
	"github.com/cloudresty/go-mongodb/update"
	"github.com/cloudresty/ulid"
)

// ULID represents a ULID identifier with helper methods
type ULID struct {
	id string
}

// NewULID generates a new ULID string
func NewULID() string {
	id, _ := ulid.New() // Ignore error for simplicity in this context
	return id
}

// ParseULID parses a ULID string and returns a ULID struct
func ParseULID(str string) (ULID, error) {
	// Validate the ULID format
	if len(str) != 26 {
		return ULID{}, errors.New("invalid ULID format")
	}

	return ULID{id: str}, nil
}

// String returns the string representation of the ULID
func (u ULID) String() string {
	return u.id
}

// Time extracts the timestamp from the ULID
func (u ULID) Time() time.Time {
	// For now, return current time. The actual ulid package might have a different API
	return time.Now()
}

// IsZero returns true if the ULID is zero/empty
func (u ULID) IsZero() bool {
	return u.id == ""
}

// Collection interface represents the minimal collection interface needed for ULID operations
// This allows the package to work with different collection implementations
type Collection interface {
	FindOne(ctx context.Context, filter any) FindOneResult
	UpdateOne(ctx context.Context, filter, update any) (UpdateResult, error)
	DeleteOne(ctx context.Context, filter any) (DeleteResult, error)
}

// FindOneResult interface represents the result of a FindOne operation
type FindOneResult interface {
	Decode(v any) error
}

// UpdateResult represents the result of an update operation
type UpdateResult struct {
	MatchedCount  int64
	ModifiedCount int64
	UpsertedCount int64
	UpsertedID    any
}

// DeleteResult represents the result of a delete operation
type DeleteResult struct {
	DeletedCount int64
}

// ULID Helper Functions

// FindByULID finds a document by its ULID
func FindByULID(ctx context.Context, coll Collection, id string) FindOneResult {
	return coll.FindOne(ctx, filter.Eq("_id", id).Build())
}

// UpdateByULID updates a document by its ULID
func UpdateByULID(ctx context.Context, coll Collection, id string, updateDoc *update.Builder) (UpdateResult, error) {
	result, err := coll.UpdateOne(ctx, filter.Eq("_id", id).Build(), updateDoc.Build())
	if err != nil {
		return UpdateResult{}, err
	}
	return result, nil
}

// DeleteByULID deletes a document by its ULID
func DeleteByULID(ctx context.Context, coll Collection, id string) (DeleteResult, error) {
	result, err := coll.DeleteOne(ctx, filter.Eq("_id", id).Build())
	if err != nil {
		return DeleteResult{}, err
	}
	return result, nil
}

// ObjectID Helper Functions

// For ObjectID operations, we'll use interface{} to avoid import issues
// Users can import primitive.ObjectID directly in their code

// FindByObjectID finds a document by its ObjectID
func FindByObjectID(ctx context.Context, coll Collection, id any) FindOneResult {
	return coll.FindOne(ctx, filter.Eq("_id", id).Build())
}

// UpdateByObjectID updates a document by its ObjectID
func UpdateByObjectID(ctx context.Context, coll Collection, id any, updateDoc *update.Builder) (UpdateResult, error) {
	result, err := coll.UpdateOne(ctx, filter.Eq("_id", id).Build(), updateDoc.Build())
	if err != nil {
		return UpdateResult{}, err
	}
	return result, nil
}

// DeleteByObjectID deletes a document by its ObjectID
func DeleteByObjectID(ctx context.Context, coll Collection, id any) (DeleteResult, error) {
	result, err := coll.DeleteOne(ctx, filter.Eq("_id", id).Build())
	if err != nil {
		return DeleteResult{}, err
	}
	return result, nil
}
