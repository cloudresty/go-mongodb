package mongodb

import (
	"context"
	"fmt"
	"maps"
	"time"

	"github.com/cloudresty/go-mongodb/filter"
	"github.com/cloudresty/go-mongodb/pipeline"
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
		col.client.config.Logger.Error("Failed to insert document",
			"error", err.Error(),
			"collection", col.name)
		return nil, err
	}

	col.client.incrementOperationCount()
	col.client.config.Logger.Debug("Document inserted successfully",
		"collection", col.name,
		"id", result.InsertedID.(string))

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
		col.client.config.Logger.Error("Failed to insert documents",
			"error", err.Error(),
			"collection", col.name)
		return nil, err
	}

	col.client.incrementOperationCount()
	col.client.config.Logger.Debug("Documents inserted successfully",
		"collection", col.name,
		"count", len(processedDocs))

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

	col.client.config.Logger.Debug("Finding document",
		"collection", col.name)

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

	col.client.config.Logger.Debug("Finding documents",
		"collection", col.name)

	cursor, err := col.collection.Find(ctx, filterDoc, opts...)
	if err != nil {
		col.client.config.Logger.Error("Failed to find documents",
			"error", err.Error(),
			"collection", col.name)
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

// FindWithOptions finds documents with QueryOptions for convenient sorting, limiting, etc.
func (col *Collection) FindWithOptions(ctx context.Context, filterBuilder *filter.Builder, queryOpts *QueryOptions) (*FindResult, error) {
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

	// Convert QueryOptions to MongoDB options
	opts := []options.Lister[options.FindOptions]{}
	if queryOpts != nil {
		findOpts := options.Find()

		if len(queryOpts.Sort) > 0 {
			findOpts.SetSort(queryOpts.Sort)
		}

		if queryOpts.Limit != nil && *queryOpts.Limit > 0 {
			findOpts.SetLimit(*queryOpts.Limit)
		}

		if queryOpts.Skip != nil && *queryOpts.Skip > 0 {
			findOpts.SetSkip(*queryOpts.Skip)
		}

		if len(queryOpts.Projection) > 0 {
			findOpts.SetProjection(queryOpts.Projection)
		}

		opts = append(opts, findOpts)
	}

	col.client.config.Logger.Debug("Finding documents with options",
		"collection", col.name,
		"hasSort", queryOpts != nil && len(queryOpts.Sort) > 0,
		"limit", queryOpts != nil && queryOpts.Limit != nil && *queryOpts.Limit > 0,
		"skip", queryOpts != nil && queryOpts.Skip != nil && *queryOpts.Skip > 0)

	cursor, err := col.collection.Find(ctx, filterDoc, opts...)
	if err != nil {
		col.client.config.Logger.Error("Failed to find documents with options",
			"error", err.Error(),
			"collection", col.name)
		return nil, err
	}

	return &FindResult{
		cursor: cursor,
	}, nil
}

// FindOneWithOptions finds a single document with QueryOptions
func (col *Collection) FindOneWithOptions(ctx context.Context, filterBuilder *filter.Builder, queryOpts *QueryOptions) *FindOneResult {
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

	// Convert QueryOptions to MongoDB options
	opts := []options.Lister[options.FindOneOptions]{}
	if queryOpts != nil {
		findOneOpts := options.FindOne()

		if len(queryOpts.Sort) > 0 {
			findOneOpts.SetSort(queryOpts.Sort)
		}

		if queryOpts.Skip != nil && *queryOpts.Skip > 0 {
			findOneOpts.SetSkip(*queryOpts.Skip)
		}

		if len(queryOpts.Projection) > 0 {
			findOneOpts.SetProjection(queryOpts.Projection)
		}

		opts = append(opts, findOneOpts)
	}

	col.client.config.Logger.Debug("Finding one document with options",
		"collection", col.name,
		"hasSort", queryOpts != nil && len(queryOpts.Sort) > 0)

	result := col.collection.FindOne(ctx, filterDoc, opts...)

	// Track read operation
	col.client.incrementOperationCount()

	return &FindOneResult{
		result: result,
	}
}

// Convenience methods for common sort operations

// FindSorted finds documents with a sort order
func (col *Collection) FindSorted(ctx context.Context, filterBuilder *filter.Builder, sort SortSpec, opts ...options.Lister[options.FindOptions]) (*FindResult, error) {
	sortBSON := convertSortSpec(sort)
	queryOpts := &QueryOptions{Sort: sortBSON}

	// For backward compatibility, if additional options are provided,
	// fall back to the original Find method with sort option manually added
	if len(opts) > 0 {
		// Create a sort option and add it to the existing options
		sortOpt := options.Find().SetSort(sortBSON)
		allOpts := append([]options.Lister[options.FindOptions]{sortOpt}, opts...)
		return col.Find(ctx, filterBuilder, allOpts...)
	}

	return col.FindWithOptions(ctx, filterBuilder, queryOpts)
}

// FindOneSorted finds a single document with a sort order
func (col *Collection) FindOneSorted(ctx context.Context, filterBuilder *filter.Builder, sort SortSpec) *FindOneResult {
	sortBSON := convertSortSpec(sort)
	queryOpts := &QueryOptions{Sort: sortBSON}
	return col.FindOneWithOptions(ctx, filterBuilder, queryOpts)
}

// FindWithLimit finds documents with a limit
func (col *Collection) FindWithLimit(ctx context.Context, filterBuilder *filter.Builder, limit int64) (*FindResult, error) {
	queryOpts := &QueryOptions{Limit: &limit}
	return col.FindWithOptions(ctx, filterBuilder, queryOpts)
}

// FindWithSkip finds documents with a skip offset
func (col *Collection) FindWithSkip(ctx context.Context, filterBuilder *filter.Builder, skip int64) (*FindResult, error) {
	queryOpts := &QueryOptions{Skip: &skip}
	return col.FindWithOptions(ctx, filterBuilder, queryOpts)
}

// FindWithProjection finds documents with field projection
func (col *Collection) FindWithProjection(ctx context.Context, filterBuilder *filter.Builder, projection bson.M) (*FindResult, error) {
	// Convert bson.M to bson.D for QueryOptions
	projectionD := make(bson.D, 0, len(projection))
	for key, value := range projection {
		projectionD = append(projectionD, bson.E{Key: key, Value: value})
	}
	queryOpts := &QueryOptions{Projection: projectionD}
	return col.FindWithOptions(ctx, filterBuilder, queryOpts)
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
		col.client.config.Logger.Error("Failed to update document",
			"error", err.Error(),
			"collection", col.name)
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

	col.client.config.Logger.Debug("Document updated successfully",
		"collection", col.name,
		"matched", int(updateResult.MatchedCount),
		"modified", int(updateResult.ModifiedCount))

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
		col.client.config.Logger.Error("Failed to update documents",
			"error", err.Error(),
			"collection", col.name)
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

	col.client.config.Logger.Debug("Documents updated successfully",
		"collection", col.name,
		"matched", int(updateResult.MatchedCount),
		"modified", int(updateResult.ModifiedCount))

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
		col.client.config.Logger.Error("Failed to replace document",
			"error", err.Error(),
			"collection", col.name)
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

	col.client.config.Logger.Debug("Document replaced successfully",
		"collection", col.name,
		"matched", int(updateResult.MatchedCount),
		"modified", int(updateResult.ModifiedCount))

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
		col.client.config.Logger.Error("Failed to delete document",
			"error", err.Error(),
			"collection", col.name)
		return nil, err
	}

	col.client.incrementOperationCount()
	col.client.config.Logger.Debug("Document deleted successfully",
		"collection", col.name,
		"deleted", int(result.DeletedCount))

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
		col.client.config.Logger.Error("Failed to delete documents",
			"error", err.Error(),
			"collection", col.name)
		return nil, err
	}

	col.client.config.Logger.Debug("Documents deleted successfully",
		"collection", col.name,
		"deleted", int(result.DeletedCount))

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
		col.client.config.Logger.Error("Failed to count documents",
			"error", err.Error(),
			"collection", col.name)
		return 0, err
	}

	col.client.config.Logger.Debug("Documents counted successfully",
		"collection", col.name,
		"count", int(count))

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
		col.client.config.Logger.Error("Failed to get distinct values",
			"error", result.Err().Error(),
			"collection", col.name,
			"field", fieldName)
		return nil, result.Err()
	}

	var values []any
	if err := result.Decode(&values); err != nil {
		return nil, err
	}

	col.client.config.Logger.Debug("Distinct values retrieved successfully",
		"collection", col.name,
		"field", fieldName,
		"count", len(values))

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
		col.client.config.Logger.Error("Failed to aggregate",
			"error", err.Error(),
			"collection", col.name)
		return nil, err
	}

	col.client.config.Logger.Debug("Aggregation started successfully",
		"collection", col.name)

	return cursor, nil
}

// AggregateWithPipeline performs an aggregation operation using a pipeline builder
func (col *Collection) AggregateWithPipeline(ctx context.Context, pipelineBuilder *pipeline.Builder, opts ...options.Lister[options.AggregateOptions]) (*AggregateResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	// Build pipeline
	pipelineDoc := bson.A{}
	if pipelineBuilder != nil {
		pipelineDoc = pipelineBuilder.ToBSONArray()
	}

	col.client.config.Logger.Debug("Aggregating with pipeline builder",
		"collection", col.name,
		"stages", len(pipelineDoc))

	cursor, err := col.collection.Aggregate(ctx, pipelineDoc, opts...)
	if err != nil {
		col.client.config.Logger.Error("Failed to aggregate with pipeline",
			"error", err.Error(),
			"collection", col.name)
		return nil, err
	}

	col.client.incrementOperationCount()

	return &AggregateResult{
		cursor: cursor,
	}, nil
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
		col.client.config.Logger.Error("Failed to create index",
			"error", err.Error(),
			"collection", col.name)
		return "", err
	}

	col.client.config.Logger.Debug("Index created successfully",
		"collection", col.name,
		"index", name)

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
		col.client.config.Logger.Error("Failed to create indexes",
			"error", err.Error(),
			"collection", col.name)
		return nil, err
	}

	col.client.config.Logger.Debug("Indexes created successfully",
		"collection", col.name,
		"count", len(names))

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
		col.client.config.Logger.Error("Failed to drop index",
			"error", err.Error(),
			"collection", col.name,
			"index", name)
		return err
	}

	col.client.config.Logger.Debug("Index dropped successfully",
		"collection", col.name,
		"index", name)

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
		col.client.config.Logger.Error("Failed to list indexes",
			"error", err.Error(),
			"collection", col.name)
		return nil, err
	}

	col.client.config.Logger.Debug("Indexes listed successfully",
		"collection", col.name)

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
		col.client.config.Logger.Error("Failed to create change stream",
			"error", err.Error(),
			"collection", col.name)
		return nil, err
	}

	col.client.config.Logger.Debug("Change stream created successfully",
		"collection", col.name)

	return stream, nil
}

// Convenience Upsert Methods

// UpsertByField performs an atomic upsert based on a specific field match
// This is a convenience method that combines filter creation, update building, and upsert execution
func (col *Collection) UpsertByField(ctx context.Context, field string, value any, document any) (*UpdateResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	// Create filter for the specified field
	filterBuilder := filter.Eq(field, value)

	// Create update using $setOnInsert for the entire document
	updateBuilder := update.New().SetOnInsertStruct(document)

	// Enable upsert
	opts := options.UpdateOne().SetUpsert(true)

	return col.UpdateOne(ctx, filterBuilder, updateBuilder, opts)
}

// UpsertByFieldMap performs an atomic upsert based on a specific field match using a map for the document
func (col *Collection) UpsertByFieldMap(ctx context.Context, field string, value any, fields map[string]any) (*UpdateResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	// Create filter for the specified field
	filterBuilder := filter.Eq(field, value)

	// Create update using $setOnInsert for the map fields
	updateBuilder := update.New().SetOnInsertMap(fields)

	// Enable upsert
	opts := options.UpdateOne().SetUpsert(true)

	return col.UpdateOne(ctx, filterBuilder, updateBuilder, opts)
}

// UpsertOptions provides configuration for upsert operations
type UpsertOptions struct {
	// OnlyInsert when true, ensures existing documents are never modified
	// This is the default behavior when using $setOnInsert
	OnlyInsert bool

	// SkipTimestamps when true, disables automatic timestamp addition
	SkipTimestamps bool
}

// UpsertByFieldWithOptions performs an atomic upsert with additional configuration options
func (col *Collection) UpsertByFieldWithOptions(ctx context.Context, field string, value any, document any, upsertOpts *UpsertOptions) (*UpdateResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	if upsertOpts == nil {
		upsertOpts = &UpsertOptions{OnlyInsert: true}
	}

	// Create filter for the specified field
	filterBuilder := filter.Eq(field, value)

	var updateBuilder *update.Builder
	if upsertOpts.OnlyInsert {
		// Use $setOnInsert to ensure existing documents are not modified
		updateBuilder = update.New().SetOnInsertStruct(document)
	} else {
		// Use $set to update existing documents as well
		updateBuilder = update.New().SetStruct(document)
	}

	// Enable upsert
	opts := options.UpdateOne().SetUpsert(true)

	return col.UpdateOne(ctx, filterBuilder, updateBuilder, opts)
}

// Helper functions

// addUpdatedAt adds or updates the updated_at field in an update document
func addUpdatedAt(update any) any {
	switch u := update.(type) {
	case bson.M:
		// Deep copy to avoid modifying the original
		enhanced := make(bson.M, len(u))
		maps.Copy(enhanced, u)

		// Check if updated_at already exists in $setOnInsert (for upserts)
		if setOnInsertOp, exists := enhanced["$setOnInsert"]; exists {
			if setOnInsertMap, ok := setOnInsertOp.(bson.M); ok {
				if _, hasUpdatedAt := setOnInsertMap["updated_at"]; hasUpdatedAt {
					// updated_at already in $setOnInsert, don't add to $set to avoid conflict
					return enhanced
				}
			}
		}

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

// Convenience methods using our BSON helpers

// FindAscending finds documents sorted by a field in ascending order
func (col *Collection) FindAscending(ctx context.Context, filterBuilder *filter.Builder, field string) (*FindResult, error) {
	return col.FindSorted(ctx, filterBuilder, SortAsc(field))
}

// FindDescending finds documents sorted by a field in descending order
func (col *Collection) FindDescending(ctx context.Context, filterBuilder *filter.Builder, field string) (*FindResult, error) {
	return col.FindSorted(ctx, filterBuilder, SortDesc(field))
}

// FindOneAscending finds a single document sorted by a field in ascending order
func (col *Collection) FindOneAscending(ctx context.Context, filterBuilder *filter.Builder, field string) *FindOneResult {
	return col.FindOneSorted(ctx, filterBuilder, SortAsc(field))
}

// FindOneDescending finds a single document sorted by a field in descending order
func (col *Collection) FindOneDescending(ctx context.Context, filterBuilder *filter.Builder, field string) *FindOneResult {
	return col.FindOneSorted(ctx, filterBuilder, SortDesc(field))
}

// FindWithProjectionFields finds documents with specific field projections
func (col *Collection) FindWithProjectionFields(ctx context.Context, filterBuilder *filter.Builder, includeFields, excludeFields []string) (*FindResult, error) {
	var projectionSpecs []ProjectionSpec
	if len(includeFields) > 0 {
		projectionSpecs = append(projectionSpecs, Include(includeFields...))
	}
	if len(excludeFields) > 0 {
		projectionSpecs = append(projectionSpecs, Exclude(excludeFields...))
	}

	projection := Projection(projectionSpecs...)

	// Convert to bson.M for compatibility with existing method
	projectionM := make(bson.M)
	for _, elem := range projection {
		projectionM[elem.Key] = elem.Value
	}

	return col.FindWithProjection(ctx, filterBuilder, projectionM)
}
