package mongodb

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/cloudresty/emit"
	"github.com/cloudresty/go-mongodb/filter"
	"github.com/cloudresty/go-mongodb/update"
	"github.com/cloudresty/ulid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Collection wraps a MongoDB collection with enhanced functionality
type Collection struct {
	collection *mongo.Collection
	client     *Client
	name       string
}

// Result types for modern API

// FindOneResult wraps mongo.SingleResult with additional methods
type FindOneResult struct {
	result *mongo.SingleResult
}

// FindResult wraps mongo.Cursor with additional methods
type FindResult struct {
	cursor *mongo.Cursor
}

// AggregateResult wraps mongo.Cursor for aggregation operations
type AggregateResult struct {
	cursor *mongo.Cursor
}

// Methods for FindOneResult
func (r *FindOneResult) Decode(v any) error {
	return r.result.Decode(v)
}

func (r *FindOneResult) Err() error {
	return r.result.Err()
}

func (r *FindOneResult) Raw() ([]byte, error) {
	return r.result.Raw()
}

// Methods for FindResult
func (r *FindResult) Next(ctx context.Context) bool {
	return r.cursor.Next(ctx)
}

func (r *FindResult) Decode(v any) error {
	return r.cursor.Decode(v)
}

func (r *FindResult) All(ctx context.Context, results any) error {
	return r.cursor.All(ctx, results)
}

func (r *FindResult) Close(ctx context.Context) error {
	return r.cursor.Close(ctx)
}

func (r *FindResult) Err() error {
	return r.cursor.Err()
}

// Methods for AggregateResult
func (r *AggregateResult) Next(ctx context.Context) bool {
	return r.cursor.Next(ctx)
}

func (r *AggregateResult) Decode(v any) error {
	return r.cursor.Decode(v)
}

func (r *AggregateResult) All(ctx context.Context, results any) error {
	return r.cursor.All(ctx, results)
}

func (r *AggregateResult) Close(ctx context.Context) error {
	return r.cursor.Close(ctx)
}

func (r *AggregateResult) Err() error {
	return r.cursor.Err()
}

// Name returns the collection name
func (col *Collection) Name() string {
	return col.name
}

// InsertOne inserts a single document with ULID generation
func (col *Collection) InsertOne(ctx context.Context, document any, opts ...options.Lister[options.InsertOneOptions]) (*InsertOneResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	// Generate ULID for _id if not present
	docMap, ok := document.(bson.M)
	if !ok {
		// Convert to bson.M if possible
		bytes, err := bson.Marshal(document)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal document: %w", err)
		}
		if err := bson.Unmarshal(bytes, &docMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal document: %w", err)
		}
	}

	// Add ULID if _id is not present
	if _, exists := docMap["_id"]; !exists {
		if col.client.config.IDMode == IDModeULID {
			id, err := ulid.New()
			if err != nil {
				return nil, fmt.Errorf("failed to generate ULID: %w", err)
			}
			docMap["_id"] = id
		}
	}

	// Add created_at and updated_at timestamps
	now := time.Now()
	if _, exists := docMap["created_at"]; !exists {
		docMap["created_at"] = now
	}
	if _, exists := docMap["updated_at"]; !exists {
		docMap["updated_at"] = now
	}

	result, err := col.collection.InsertOne(ctx, docMap, opts...)
	if err != nil {
		col.client.incrementFailureCount()
		emit.Error.StructuredFields("Failed to insert document",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return nil, err
	}

	col.client.incrementOperationCount()
	emit.Debug.StructuredFields("Document inserted successfully",
		emit.ZString("collection", col.name),
		emit.ZString("id", result.InsertedID.(string)))

	return &InsertOneResult{
		InsertedID:  result.InsertedID.(string),
		GeneratedAt: now,
	}, nil
}

// InsertMany inserts multiple documents with ULID generation
func (col *Collection) InsertMany(ctx context.Context, documents []any, opts ...options.Lister[options.InsertManyOptions]) (*InsertManyResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	now := time.Now()
	var processedDocs []any
	var generatedIDs []string

	for _, doc := range documents {
		// Convert to bson.M
		docMap, ok := doc.(bson.M)
		if !ok {
			bytes, err := bson.Marshal(doc)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal document: %w", err)
			}
			if err := bson.Unmarshal(bytes, &docMap); err != nil {
				return nil, fmt.Errorf("failed to unmarshal document: %w", err)
			}
		}

		// Add ULID if _id is not present
		if _, exists := docMap["_id"]; !exists {
			if col.client.config.IDMode == IDModeULID {
				newID, err := ulid.New()
				if err != nil {
					return nil, fmt.Errorf("failed to generate ULID: %w", err)
				}
				docMap["_id"] = newID
				generatedIDs = append(generatedIDs, newID)
			}
		} else {
			if idStr, ok := docMap["_id"].(string); ok {
				generatedIDs = append(generatedIDs, idStr)
			}
		}

		// Add timestamps
		if _, exists := docMap["created_at"]; !exists {
			docMap["created_at"] = now
		}
		if _, exists := docMap["updated_at"]; !exists {
			docMap["updated_at"] = now
		}

		processedDocs = append(processedDocs, docMap)
	}

	_, err := col.collection.InsertMany(ctx, processedDocs, opts...)
	if err != nil {
		col.client.incrementFailureCount()
		emit.Error.StructuredFields("Failed to insert documents",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return nil, err
	}

	col.client.incrementOperationCount()
	emit.Debug.StructuredFields("Documents inserted successfully",
		emit.ZString("collection", col.name),
		emit.ZInt("count", len(processedDocs)))

	return &InsertManyResult{
		InsertedIDs:   generatedIDs,
		InsertedCount: int64(len(processedDocs)),
		GeneratedAt:   now,
	}, nil
}

// FindOne finds a single document using a filter builder
func (col *Collection) FindOne(ctx context.Context, filterBuilder *filter.Builder, opts ...options.Lister[options.FindOneOptions]) *FindOneResult {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	// Build filter document
	filterDoc := bson.M{}
	if filterBuilder != nil {
		filterDoc = filterBuilder.Build()
	}

	emit.Debug.StructuredFields("Finding document",
		emit.ZString("collection", col.name))

	result := col.collection.FindOne(ctx, filterDoc, opts...)

	// Track read operation (Note: MongoDB SingleResult doesn't expose error until Decode())
	col.client.incrementOperationCount()

	return &FindOneResult{
		result: result,
	}
}

// Find finds multiple documents using a filter builder
func (col *Collection) Find(ctx context.Context, filterBuilder *filter.Builder, opts ...options.Lister[options.FindOptions]) (*FindResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	// Build filter document
	filterDoc := bson.M{}
	if filterBuilder != nil {
		filterDoc = filterBuilder.Build()
	}

	emit.Debug.StructuredFields("Finding documents",
		emit.ZString("collection", col.name))

	cursor, err := col.collection.Find(ctx, filterDoc, opts...)
	if err != nil {
		emit.Error.StructuredFields("Failed to find documents",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return nil, err
	}

	return &FindResult{
		cursor: cursor,
	}, nil
}

// FindByULID finds a document by its ULID
func (col *Collection) FindByULID(ctx context.Context, ulid string) *FindOneResult {
	filterBuilder := filter.Eq("_id", ulid)
	return col.FindOne(ctx, filterBuilder)
}

// UpdateOne updates a single document
func (col *Collection) UpdateOne(ctx context.Context, filterBuilder *filter.Builder, updateBuilder *update.Builder, opts ...options.Lister[options.UpdateOneOptions]) (*UpdateResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	// Build filter and update documents
	filterDoc := bson.M{}
	if filterBuilder != nil {
		filterDoc = filterBuilder.Build()
	}

	updateDoc := bson.M{}
	if updateBuilder != nil {
		updateDoc = updateBuilder.Build()
	}

	// Add updated_at timestamp
	enhancedUpdate := addUpdatedAt(updateDoc)

	result, err := col.collection.UpdateOne(ctx, filterDoc, enhancedUpdate, opts...)
	if err != nil {
		col.client.incrementFailureCount()
		emit.Error.StructuredFields("Failed to update document",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return nil, err
	}

	col.client.incrementOperationCount()

	updateResult := &UpdateResult{
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		UpsertedCount: result.UpsertedCount,
	}

	if result.UpsertedID != nil {
		updateResult.UpsertedID = result.UpsertedID.(string)
	}

	emit.Debug.StructuredFields("Document updated successfully",
		emit.ZString("collection", col.name),
		emit.ZInt("matched", int(updateResult.MatchedCount)),
		emit.ZInt("modified", int(updateResult.ModifiedCount)))

	return updateResult, nil
}

// UpdateMany updates multiple documents
func (col *Collection) UpdateMany(ctx context.Context, filterBuilder *filter.Builder, updateBuilder *update.Builder, opts ...options.Lister[options.UpdateManyOptions]) (*UpdateResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	// Build filter and update documents
	filterDoc := bson.M{}
	if filterBuilder != nil {
		filterDoc = filterBuilder.Build()
	}

	updateDoc := bson.M{}
	if updateBuilder != nil {
		updateDoc = updateBuilder.Build()
	}

	// Add updated_at timestamp
	enhancedUpdate := addUpdatedAt(updateDoc)

	result, err := col.collection.UpdateMany(ctx, filterDoc, enhancedUpdate, opts...)
	if err != nil {
		emit.Error.StructuredFields("Failed to update documents",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return nil, err
	}

	updateResult := &UpdateResult{
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		UpsertedCount: result.UpsertedCount,
	}

	if result.UpsertedID != nil {
		updateResult.UpsertedID = result.UpsertedID.(string)
	}

	emit.Debug.StructuredFields("Documents updated successfully",
		emit.ZString("collection", col.name),
		emit.ZInt("matched", int(updateResult.MatchedCount)),
		emit.ZInt("modified", int(updateResult.ModifiedCount)))

	return updateResult, nil
}

// ReplaceOne replaces a single document
func (col *Collection) ReplaceOne(ctx context.Context, filterBuilder *filter.Builder, replacement any, opts ...options.Lister[options.ReplaceOptions]) (*UpdateResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	// Build filter document
	filterDoc := bson.M{}
	if filterBuilder != nil {
		filterDoc = filterBuilder.Build()
	}

	// Enhance the replacement document
	enhanced := enhanceReplacementDocument(replacement)

	result, err := col.collection.ReplaceOne(ctx, filterDoc, enhanced, opts...)
	if err != nil {
		emit.Error.StructuredFields("Failed to replace document",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return nil, err
	}

	updateResult := &UpdateResult{
		MatchedCount:  result.MatchedCount,
		ModifiedCount: result.ModifiedCount,
		UpsertedCount: result.UpsertedCount,
	}

	if result.UpsertedID != nil {
		updateResult.UpsertedID = result.UpsertedID.(string)
	}

	emit.Debug.StructuredFields("Document replaced successfully",
		emit.ZString("collection", col.name),
		emit.ZInt("matched", int(updateResult.MatchedCount)),
		emit.ZInt("modified", int(updateResult.ModifiedCount)))

	return updateResult, nil
}

// DeleteOne deletes a single document
func (col *Collection) DeleteOne(ctx context.Context, filterBuilder *filter.Builder, opts ...options.Lister[options.DeleteOneOptions]) (*DeleteResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	// Build filter document
	filterDoc := bson.M{}
	if filterBuilder != nil {
		filterDoc = filterBuilder.Build()
	}

	result, err := col.collection.DeleteOne(ctx, filterDoc, opts...)
	if err != nil {
		col.client.incrementFailureCount()
		emit.Error.StructuredFields("Failed to delete document",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return nil, err
	}

	col.client.incrementOperationCount()
	emit.Debug.StructuredFields("Document deleted successfully",
		emit.ZString("collection", col.name),
		emit.ZInt("deleted", int(result.DeletedCount)))

	return &DeleteResult{
		DeletedCount: result.DeletedCount,
	}, nil
}

// DeleteMany deletes multiple documents
func (col *Collection) DeleteMany(ctx context.Context, filterBuilder *filter.Builder, opts ...options.Lister[options.DeleteManyOptions]) (*DeleteResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	// Build filter document
	filterDoc := bson.M{}
	if filterBuilder != nil {
		filterDoc = filterBuilder.Build()
	}

	result, err := col.collection.DeleteMany(ctx, filterDoc, opts...)
	if err != nil {
		emit.Error.StructuredFields("Failed to delete documents",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return nil, err
	}

	emit.Debug.StructuredFields("Documents deleted successfully",
		emit.ZString("collection", col.name),
		emit.ZInt("deleted", int(result.DeletedCount)))

	return &DeleteResult{
		DeletedCount: result.DeletedCount,
	}, nil
}

// CountDocuments counts documents in the collection
func (col *Collection) CountDocuments(ctx context.Context, filterBuilder *filter.Builder, opts ...options.Lister[options.CountOptions]) (int64, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	// Build filter document
	filterDoc := bson.M{}
	if filterBuilder != nil {
		filterDoc = filterBuilder.Build()
	}

	count, err := col.collection.CountDocuments(ctx, filterDoc, opts...)
	if err != nil {
		emit.Error.StructuredFields("Failed to count documents",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return 0, err
	}

	emit.Debug.StructuredFields("Documents counted successfully",
		emit.ZString("collection", col.name),
		emit.ZInt("count", int(count)))

	return count, nil
}

// Distinct returns distinct values for a field
func (col *Collection) Distinct(ctx context.Context, fieldName string, filterBuilder *filter.Builder, opts ...options.Lister[options.DistinctOptions]) ([]any, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	// Build filter document
	filterDoc := bson.M{}
	if filterBuilder != nil {
		filterDoc = filterBuilder.Build()
	}

	result := col.collection.Distinct(ctx, fieldName, filterDoc, opts...)
	if result.Err() != nil {
		emit.Error.StructuredFields("Failed to get distinct values",
			emit.ZString("error", result.Err().Error()),
			emit.ZString("collection", col.name),
			emit.ZString("field", fieldName))
		return nil, result.Err()
	}

	var values []any
	if err := result.Decode(&values); err != nil {
		return nil, err
	}

	emit.Debug.StructuredFields("Distinct values retrieved successfully",
		emit.ZString("collection", col.name),
		emit.ZString("field", fieldName),
		emit.ZInt("count", len(values)))

	return values, nil
}

// Aggregate performs an aggregation operation
func (col *Collection) Aggregate(ctx context.Context, pipeline any, opts ...options.Lister[options.AggregateOptions]) (*mongo.Cursor, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	cursor, err := col.collection.Aggregate(ctx, pipeline, opts...)
	if err != nil {
		emit.Error.StructuredFields("Failed to aggregate",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return nil, err
	}

	emit.Debug.StructuredFields("Aggregation started successfully",
		emit.ZString("collection", col.name))

	return cursor, nil
}

// Indexes returns the index operations for this collection
func (col *Collection) Indexes() mongo.IndexView {
	return col.collection.Indexes()
}

// CreateIndex creates a single index
func (col *Collection) CreateIndex(ctx context.Context, model mongo.IndexModel, opts ...options.Lister[options.CreateIndexesOptions]) (string, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	name, err := col.collection.Indexes().CreateOne(ctx, model, opts...)
	if err != nil {
		emit.Error.StructuredFields("Failed to create index",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return "", err
	}

	emit.Debug.StructuredFields("Index created successfully",
		emit.ZString("collection", col.name),
		emit.ZString("index", name))

	return name, nil
}

// CreateIndexes creates multiple indexes
func (col *Collection) CreateIndexes(ctx context.Context, models []mongo.IndexModel, opts ...options.Lister[options.CreateIndexesOptions]) ([]string, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	names, err := col.collection.Indexes().CreateMany(ctx, models, opts...)
	if err != nil {
		emit.Error.StructuredFields("Failed to create indexes",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return nil, err
	}

	emit.Debug.StructuredFields("Indexes created successfully",
		emit.ZString("collection", col.name),
		emit.ZInt("count", len(names)))

	return names, nil
}

// DropIndex drops a single index
func (col *Collection) DropIndex(ctx context.Context, name string, opts ...options.Lister[options.DropIndexesOptions]) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	err := col.collection.Indexes().DropOne(ctx, name, opts...)
	if err != nil {
		emit.Error.StructuredFields("Failed to drop index",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name),
			emit.ZString("index", name))
		return err
	}

	emit.Debug.StructuredFields("Index dropped successfully",
		emit.ZString("collection", col.name),
		emit.ZString("index", name))

	return nil
}

// ListIndexes returns a cursor for all indexes in the collection
func (col *Collection) ListIndexes(ctx context.Context, opts ...options.Lister[options.ListIndexesOptions]) (*mongo.Cursor, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	cursor, err := col.collection.Indexes().List(ctx, opts...)
	if err != nil {
		emit.Error.StructuredFields("Failed to list indexes",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return nil, err
	}

	emit.Debug.StructuredFields("Indexes listed successfully",
		emit.ZString("collection", col.name))

	return cursor, nil
}

// Watch returns a change stream for the collection
func (col *Collection) Watch(ctx context.Context, pipeline any, opts ...options.Lister[options.ChangeStreamOptions]) (*mongo.ChangeStream, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	stream, err := col.collection.Watch(ctx, pipeline, opts...)
	if err != nil {
		emit.Error.StructuredFields("Failed to create change stream",
			emit.ZString("error", err.Error()),
			emit.ZString("collection", col.name))
		return nil, err
	}

	emit.Debug.StructuredFields("Change stream created successfully",
		emit.ZString("collection", col.name))

	return stream, nil
}

// Helper functions

// addUpdatedAt adds or updates the updated_at field in an update document
func addUpdatedAt(update any) any {
	switch u := update.(type) {
	case bson.M:
		// Deep copy to avoid modifying the original
		enhanced := make(bson.M, len(u))
		maps.Copy(enhanced, u)

		// Add updated_at to $set operations
		if setOp, exists := enhanced["$set"]; exists {
			if setMap, ok := setOp.(bson.M); ok {
				setMap["updated_at"] = time.Now()
			}
		} else {
			enhanced["$set"] = bson.M{"updated_at": time.Now()}
		}
		return enhanced

	case bson.D:
		// Convert bson.D to bson.M
		m := make(bson.M)
		for _, elem := range u {
			m[elem.Key] = elem.Value
		}
		return addUpdatedAt(m)

	default:
		// For other types, try to convert to bson.M
		bytes, err := bson.Marshal(update)
		if err != nil {
			// If conversion fails, return original with a simple wrapper
			return bson.M{
				"$set": bson.M{
					"updated_at": time.Now(),
				},
			}
		}

		var m bson.M
		if err := bson.Unmarshal(bytes, &m); err != nil {
			return bson.M{
				"$set": bson.M{
					"updated_at": time.Now(),
				},
			}
		}

		return addUpdatedAt(m)
	}
}

// enhanceReplacementDocument adds timestamps and ULID to replacement documents
func enhanceReplacementDocument(doc any) bson.M {
	var enhanced bson.M

	switch d := doc.(type) {
	case bson.M:
		enhanced = make(bson.M, len(d)+2) // +2 for potential timestamps
		maps.Copy(enhanced, d)
	case bson.D:
		enhanced = make(bson.M)
		for _, elem := range d {
			enhanced[elem.Key] = elem.Value
		}
	default:
		// Try to convert to bson.M
		bytes, err := bson.Marshal(doc)
		if err != nil {
			enhanced = bson.M{}
		} else {
			if err := bson.Unmarshal(bytes, &enhanced); err != nil {
				enhanced = bson.M{}
			}
		}
	}

	// Add or update timestamps
	now := time.Now()
	if _, exists := enhanced["updated_at"]; !exists {
		enhanced["updated_at"] = now
	}

	// Only add created_at if it doesn't exist (preserve original creation time)
	if _, exists := enhanced["created_at"]; !exists {
		enhanced["created_at"] = now
	}

	return enhanced
}
