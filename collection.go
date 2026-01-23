package mongodb

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/cloudresty/go-mongodb/v2/filter"
	"github.com/cloudresty/go-mongodb/v2/pipeline"
	"github.com/cloudresty/go-mongodb/v2/update"
	"github.com/cloudresty/ulid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// idFieldInfo holds cached information about a struct's ID field.
// This avoids repeated reflection on the same type.
type idFieldInfo struct {
	// hasIDField indicates if the struct has an _id field
	hasIDField bool
	// fieldIndex is the index of the ID field in the struct (valid only if hasIDField is true)
	fieldIndex int
	// isStringType indicates if the ID field is a string type (can accept ULID directly)
	isStringType bool
	// isObjectIDType indicates if the ID field is primitive.ObjectID (incompatible with ULID)
	isObjectIDType bool
	// fieldTypeName stores the type name for error messages
	fieldTypeName string
}

// idFieldCache caches ID field metadata by reflect.Type to avoid repeated reflection.
// This is a package-level cache that persists across all collection operations.
var idFieldCache sync.Map // map[reflect.Type]*idFieldInfo

// ErrULIDIncompatibleType is returned when IDMode is ULID but the struct has a non-string ID field.
// This prevents data corruption where a string ULID would be inserted but cannot be decoded
// back into the actual field type (e.g., ObjectID, int64, []byte, custom types).
// Only string or interface{} fields are compatible with ULID mode.
var ErrULIDIncompatibleType = fmt.Errorf("IDMode is ULID but struct ID field has incompatible type")

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

// hasID checks if a document already has an _id field without full marshal/unmarshal.
// Returns (hasID, existingID) where existingID is only valid if hasID is true.
func hasID(document any) (bool, any) {
	switch doc := document.(type) {
	case bson.M:
		if id, exists := doc["_id"]; exists {
			return true, id
		}
		return false, nil
	case bson.D:
		for _, elem := range doc {
			if elem.Key == "_id" {
				return true, elem.Value
			}
		}
		return false, nil
	case map[string]any:
		if id, exists := doc["_id"]; exists {
			return true, id
		}
		return false, nil
	default:
		// Use reflection for structs
		return hasIDReflect(document)
	}
}

// hasIDReflect uses reflection to check if a struct has a non-zero _id field.
// Uses a type cache to avoid repeated reflection on the same struct type.
func hasIDReflect(document any) (bool, any) {
	v := reflect.ValueOf(document)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false, nil
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return false, nil
	}

	t := v.Type()

	// Check the cache first
	if cached, ok := idFieldCache.Load(t); ok {
		info := cached.(*idFieldInfo)
		if !info.hasIDField {
			return false, nil
		}
		fieldVal := v.Field(info.fieldIndex)
		if !fieldVal.IsZero() {
			return true, fieldVal.Interface()
		}
		return false, nil
	}

	// Not in cache - perform reflection and cache the result
	info := findIDFieldInfo(t)
	idFieldCache.Store(t, info)

	if !info.hasIDField {
		return false, nil
	}
	fieldVal := v.Field(info.fieldIndex)
	if !fieldVal.IsZero() {
		return true, fieldVal.Interface()
	}
	return false, nil
}

// findIDFieldInfo performs reflection to find the ID field in a struct type.
// This is called once per type and the result is cached.
func findIDFieldInfo(t reflect.Type) *idFieldInfo {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// Check bson tag first, then json tag, then field name
		tag := field.Tag.Get("bson")
		if tag == "" {
			tag = field.Tag.Get("json")
		}
		// Parse tag to get field name (before comma)
		if idx := len(tag); idx > 0 {
			for j := 0; j < len(tag); j++ {
				if tag[j] == ',' {
					tag = tag[:j]
					break
				}
			}
		}

		if tag == "_id" || (tag == "" && field.Name == "ID") {
			// Check if field is ObjectID type (primitive.ObjectID)
			isObjectID := field.Type.String() == "primitive.ObjectID"
			return &idFieldInfo{
				hasIDField:     true,
				fieldIndex:     i,
				isStringType:   field.Type.Kind() == reflect.String,
				isObjectIDType: isObjectID,
				fieldTypeName:  field.Type.String(),
			}
		}
	}
	return &idFieldInfo{hasIDField: false, fieldIndex: -1, isStringType: false, isObjectIDType: false, fieldTypeName: ""}
}

// inspectResult holds the result of inspecting a document for ID information.
// This consolidates all ID-related checks into a single pass.
type inspectResult struct {
	hasID    bool          // Whether the document already has a non-zero ID
	idValue  any           // The current ID value (if hasID is true)
	info     *idFieldInfo  // Cached field info (for structs only)
	isStruct bool          // Whether the document is a struct (or pointer to struct)
	isPtr    bool          // Whether the document is a pointer (can be modified in place)
	elemVal  reflect.Value // The struct value (for setting fields)
}

// inspectStruct performs a single-pass inspection of a struct document.
// Returns all ID-related information needed for insert operations.
// This avoids multiple cache lookups by consolidating hasID and field info checks.
func inspectStruct(document any) *inspectResult {
	result := &inspectResult{}

	v := reflect.ValueOf(document)
	if v.Kind() == reflect.Ptr {
		result.isPtr = true
		if v.IsNil() {
			return result
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return result
	}

	result.isStruct = true
	result.elemVal = v
	t := v.Type()

	// Get cached field info (single lookup)
	if cached, ok := idFieldCache.Load(t); ok {
		result.info = cached.(*idFieldInfo)
	} else {
		result.info = findIDFieldInfo(t)
		idFieldCache.Store(t, result.info)
	}

	// Check if ID field has a value
	if result.info.hasIDField {
		fieldVal := v.Field(result.info.fieldIndex)
		if !fieldVal.IsZero() {
			result.hasID = true
			result.idValue = fieldVal.Interface()
		}
	}

	return result
}

// trySetULIDOnStruct attempts to set a ULID directly on a struct's ID field using reflection.
// Uses pre-computed inspectResult to avoid duplicate cache lookups.
// Returns (modifiedDocument, generatedID, success). If success is false, caller should fall back
// to the marshal/unmarshal approach.
func trySetULIDOnStructWithInfo(document any, result *inspectResult) (any, string, bool) {
	// Can only do zero-allocation injection if:
	// 1. Document is a pointer (so we can modify it)
	// 2. The struct has an ID field
	// 3. The ID field is a string type
	// 4. The field is settable
	if !result.isPtr || !result.isStruct || result.info == nil {
		return document, "", false
	}
	if !result.info.hasIDField || !result.info.isStringType {
		return document, "", false
	}

	field := result.elemVal.Field(result.info.fieldIndex)
	if !field.CanSet() {
		return document, "", false
	}

	// Generate ULID
	id, err := ulid.New()
	if err != nil {
		return document, "", false
	}

	// Set the ULID directly on the struct field
	field.SetString(id)

	return document, id, true
}

// prepareDocumentForInsert prepares a document for insertion, adding ULID if needed.
// Returns the document to insert and the generated ID (if any).
//
// Performance: For struct pointers with string ID fields, this uses zero-allocation
// ID injection by setting the ULID directly on the struct field, avoiding marshal/unmarshal.
//
// Safety: Returns ErrULIDObjectIDMismatch if IDMode is ULID but the struct has an ObjectID field,
// preventing data corruption where a string ULID would be inserted but cannot be decoded back.
func (col *Collection) prepareDocumentForInsert(document any) (any, error) {
	// Fast path for non-ULID modes
	if col.client.config.IDMode != IDModeULID {
		return document, nil
	}

	// Handle known map types first (no reflection needed)
	switch doc := document.(type) {
	case bson.M:
		if _, exists := doc["_id"]; exists {
			return document, nil
		}
		// Add ULID to map
		id, err := ulid.New()
		if err != nil {
			return nil, fmt.Errorf("failed to generate ULID: %w", err)
		}
		doc["_id"] = id
		return doc, nil
	case bson.D:
		for _, elem := range doc {
			if elem.Key == "_id" {
				return document, nil
			}
		}
		// Convert to bson.M and add ULID
		docMap := make(bson.M, len(doc)+1)
		for _, elem := range doc {
			docMap[elem.Key] = elem.Value
		}
		id, err := ulid.New()
		if err != nil {
			return nil, fmt.Errorf("failed to generate ULID: %w", err)
		}
		docMap["_id"] = id
		return docMap, nil
	case map[string]any:
		if _, exists := doc["_id"]; exists {
			return document, nil
		}
		// Add ULID to map
		id, err := ulid.New()
		if err != nil {
			return nil, fmt.Errorf("failed to generate ULID: %w", err)
		}
		doc["_id"] = id
		return doc, nil
	}

	// For structs, use single-pass inspection
	result := inspectStruct(document)

	// If document already has an ID, pass through
	if result.hasID {
		return document, nil
	}

	// Safety check: If struct has an ID field that is NOT string or interface{}, reject it.
	// A ULID is a string, so inserting it into a non-string field (int64, ObjectID, []byte, etc.)
	// would cause decode failures when reading the document back.
	if result.isStruct && result.info != nil && result.info.hasIDField {
		// Only string and interface{} are compatible with ULID mode
		if !result.info.isStringType && result.info.fieldTypeName != "interface {}" {
			return nil, fmt.Errorf("%w: field type is %s; use string or interface{}", ErrULIDIncompatibleType, result.info.fieldTypeName)
		}
	}

	// Try zero-allocation ID injection for struct pointers with string ID fields
	if modifiedDoc, _, ok := trySetULIDOnStructWithInfo(document, result); ok {
		return modifiedDoc, nil
	}

	// Fall back to marshal/unmarshal for non-pointer structs or non-string ID fields
	var docMap bson.M
	bytes, err := bson.Marshal(document)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal document: %w", err)
	}
	if err := bson.Unmarshal(bytes, &docMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal document: %w", err)
	}

	// Generate and add ULID
	id, err := ulid.New()
	if err != nil {
		return nil, fmt.Errorf("failed to generate ULID: %w", err)
	}
	docMap["_id"] = id

	return docMap, nil
}

// InsertOne inserts a single document with automatic ULID generation when IDMode is IDModeULID.
//
// Performance: For struct pointers with string ID fields (e.g., *User with ID string `bson:"_id"`),
// this uses zero-allocation ID injection by setting the ULID directly on the struct field.
// This avoids the marshal/unmarshal overhead and provides optimal performance.
//
// For non-pointer structs or non-string ID fields, the document is converted to bson.M.
func (col *Collection) InsertOne(ctx context.Context, document any, opts ...options.Lister[options.InsertOneOptions]) (*InsertOneResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	// Prepare document (add ULID if needed)
	docToInsert, err := col.prepareDocumentForInsert(document)
	if err != nil {
		return nil, err
	}

	result, err := col.collection.InsertOne(ctx, docToInsert, opts...)
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
		"id", result.InsertedID)

	return &InsertOneResult{
		InsertedID:  result.InsertedID,
		GeneratedAt: time.Now(),
	}, nil
}

// InsertMany inserts multiple documents with automatic ULID generation when IDMode is IDModeULID.
//
// Performance Note: When using IDModeULID with struct documents that don't have an _id field set,
// each document undergoes marshal/unmarshal to add the ULID. For maximum performance in high-throughput
// scenarios, either:
//   - Pass bson.M or bson.D directly (no conversion needed)
//   - Pre-set the ID field on your structs before insertion
//   - Use IDModeObjectID or IDModeCustom to skip ULID generation
func (col *Collection) InsertMany(ctx context.Context, documents []any, opts ...options.Lister[options.InsertManyOptions]) (*InsertManyResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	now := time.Now()
	processedDocs := make([]any, 0, len(documents))
	generatedIDs := make([]any, 0, len(documents))

	for _, doc := range documents {
		// Use prepareDocumentForInsert for consistent handling (includes safety checks)
		preparedDoc, err := col.prepareDocumentForInsert(doc)
		if err != nil {
			return nil, err
		}

		// Extract the ID from the prepared document
		_, docID := hasID(preparedDoc)
		processedDocs = append(processedDocs, preparedDoc)
		generatedIDs = append(generatedIDs, docID)
	}

	result, err := col.collection.InsertMany(ctx, processedDocs, opts...)
	if err != nil {
		col.client.incrementFailureCount()
		col.client.config.Logger.Error("Failed to insert documents",
			"error", err.Error(),
			"collection", col.name)
		return nil, err
	}

	// Update generatedIDs with actual IDs from result for documents where we didn't set ID
	for i, id := range generatedIDs {
		if id == nil && i < len(result.InsertedIDs) {
			generatedIDs[i] = result.InsertedIDs[i]
		}
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

	result, err := col.collection.UpdateOne(ctx, filterDoc, updateDoc, opts...)
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
		UpsertedID:    result.UpsertedID,
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

	result, err := col.collection.UpdateMany(ctx, filterDoc, updateDoc, opts...)
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
		UpsertedID:    result.UpsertedID,
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

	result, err := col.collection.ReplaceOne(ctx, filterDoc, replacement, opts...)
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
		UpsertedID:    result.UpsertedID,
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

// CreateIndex creates a single index using the library's IndexModel type.
// This eliminates the need to import mongo-driver directly for index operations.
// Use helper functions like IndexAsc(), IndexDesc(), IndexUnique(), IndexText() to create IndexModel.
func (col *Collection) CreateIndex(ctx context.Context, model IndexModel, opts ...options.Lister[options.CreateIndexesOptions]) (string, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	// Convert library's IndexModel to mongo.IndexModel
	mongoModel := mongo.IndexModel{
		Keys:    model.Keys,
		Options: model.Options,
	}

	name, err := col.collection.Indexes().CreateOne(ctx, mongoModel, opts...)
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

// CreateIndexes creates multiple indexes using the library's IndexModel type.
// This eliminates the need to import mongo-driver directly for index operations.
// Use helper functions like IndexAsc(), IndexDesc(), IndexUnique(), IndexText() to create IndexModel.
func (col *Collection) CreateIndexes(ctx context.Context, models []IndexModel, opts ...options.Lister[options.CreateIndexesOptions]) ([]string, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
	}

	// Convert library's IndexModels to mongo.IndexModels
	mongoModels := make([]mongo.IndexModel, len(models))
	for i, model := range models {
		mongoModels[i] = mongo.IndexModel{
			Keys:    model.Keys,
			Options: model.Options,
		}
	}

	names, err := col.collection.Indexes().CreateMany(ctx, mongoModels, opts...)
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
	updateBuilder, err := update.New().SetOnInsertStruct(document)
	if err != nil {
		return nil, fmt.Errorf("failed to create update builder: %w", err)
	}

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
	var err error
	if upsertOpts.OnlyInsert {
		// Use $setOnInsert to ensure existing documents are not modified
		updateBuilder, err = update.New().SetOnInsertStruct(document)
	} else {
		// Use $set to update existing documents as well
		updateBuilder, err = update.New().SetStruct(document)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create update builder: %w", err)
	}

	// Enable upsert
	opts := options.UpdateOne().SetUpsert(true)

	return col.UpdateOne(ctx, filterBuilder, updateBuilder, opts)
}

// ReturnDocument specifies when to capture the document for FindOneAnd* operations.
type ReturnDocument int

const (
	// ReturnBefore returns the document before the modification (default)
	ReturnBefore ReturnDocument = iota
	// ReturnAfter returns the document after the modification
	ReturnAfter
)

// FindOneAndUpdateOptions configures FindOneAndUpdate operations.
type FindOneAndUpdateOptions struct {
	// ReturnDocument specifies whether to return the document before or after the update.
	// Use ReturnBefore (default) or ReturnAfter.
	ReturnDocument ReturnDocument

	// Upsert, when true, creates a new document if no document matches the filter.
	Upsert bool

	// Sort determines which document to update if multiple match.
	Sort bson.D

	// Projection limits the fields returned in the document.
	Projection bson.D
}

// FindOneAndUpdateOpts creates a new FindOneAndUpdateOptions with default values.
func FindOneAndUpdateOpts() *FindOneAndUpdateOptions {
	return &FindOneAndUpdateOptions{}
}

// SetReturnDocument sets the return document option.
func (o *FindOneAndUpdateOptions) SetReturnDocument(rd ReturnDocument) *FindOneAndUpdateOptions {
	o.ReturnDocument = rd
	return o
}

// SetUpsert sets the upsert option.
func (o *FindOneAndUpdateOptions) SetUpsert(upsert bool) *FindOneAndUpdateOptions {
	o.Upsert = upsert
	return o
}

// SetSort sets the sort order for determining which document to update.
func (o *FindOneAndUpdateOptions) SetSort(sort bson.D) *FindOneAndUpdateOptions {
	o.Sort = sort
	return o
}

// SetProjection sets the fields to return in the document.
func (o *FindOneAndUpdateOptions) SetProjection(projection bson.D) *FindOneAndUpdateOptions {
	o.Projection = projection
	return o
}

// FindOneAndReplaceOptions configures FindOneAndReplace operations.
type FindOneAndReplaceOptions struct {
	// ReturnDocument specifies whether to return the document before or after the replacement.
	ReturnDocument ReturnDocument

	// Upsert, when true, creates a new document if no document matches the filter.
	Upsert bool

	// Sort determines which document to replace if multiple match.
	Sort bson.D

	// Projection limits the fields returned in the document.
	Projection bson.D
}

// FindOneAndReplaceOpts creates a new FindOneAndReplaceOptions with default values.
func FindOneAndReplaceOpts() *FindOneAndReplaceOptions {
	return &FindOneAndReplaceOptions{}
}

// SetReturnDocument sets the return document option.
func (o *FindOneAndReplaceOptions) SetReturnDocument(rd ReturnDocument) *FindOneAndReplaceOptions {
	o.ReturnDocument = rd
	return o
}

// SetUpsert sets the upsert option.
func (o *FindOneAndReplaceOptions) SetUpsert(upsert bool) *FindOneAndReplaceOptions {
	o.Upsert = upsert
	return o
}

// SetSort sets the sort order for determining which document to replace.
func (o *FindOneAndReplaceOptions) SetSort(sort bson.D) *FindOneAndReplaceOptions {
	o.Sort = sort
	return o
}

// SetProjection sets the fields to return in the document.
func (o *FindOneAndReplaceOptions) SetProjection(projection bson.D) *FindOneAndReplaceOptions {
	o.Projection = projection
	return o
}

// FindOneAndDeleteOptions configures FindOneAndDelete operations.
type FindOneAndDeleteOptions struct {
	// Sort determines which document to delete if multiple match.
	Sort bson.D

	// Projection limits the fields returned in the document.
	Projection bson.D
}

// FindOneAndDeleteOpts creates a new FindOneAndDeleteOptions with default values.
func FindOneAndDeleteOpts() *FindOneAndDeleteOptions {
	return &FindOneAndDeleteOptions{}
}

// SetSort sets the sort order for determining which document to delete.
func (o *FindOneAndDeleteOptions) SetSort(sort bson.D) *FindOneAndDeleteOptions {
	o.Sort = sort
	return o
}

// SetProjection sets the fields to return in the document.
func (o *FindOneAndDeleteOptions) SetProjection(projection bson.D) *FindOneAndDeleteOptions {
	o.Projection = projection
	return o
}

// FindOneAndUpdate atomically finds a document, applies an update, and returns
// either the original or the modified document based on options.
// This is essential for atomic operations like counters, reservations, and queue processing.
func (col *Collection) FindOneAndUpdate(ctx context.Context, filterBuilder *filter.Builder, updateBuilder *update.Builder, opts ...*FindOneAndUpdateOptions) *FindOneResult {
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

	// Build update document
	updateDoc := bson.M{}
	if updateBuilder != nil {
		updateDoc = updateBuilder.Build()
	}

	// Convert our options to mongo driver options
	driverOpts := options.FindOneAndUpdate()

	if len(opts) > 0 && opts[0] != nil {
		opt := opts[0]

		if opt.ReturnDocument == ReturnAfter {
			driverOpts.SetReturnDocument(options.After)
		} else {
			driverOpts.SetReturnDocument(options.Before)
		}

		if opt.Upsert {
			driverOpts.SetUpsert(true)
		}

		if len(opt.Sort) > 0 {
			driverOpts.SetSort(opt.Sort)
		}

		if len(opt.Projection) > 0 {
			driverOpts.SetProjection(opt.Projection)
		}
	}

	col.client.config.Logger.Debug("FindOneAndUpdate",
		"collection", col.name)

	result := col.collection.FindOneAndUpdate(ctx, filterDoc, updateDoc, driverOpts)

	col.client.incrementOperationCount()

	return &FindOneResult{
		result: result,
	}
}

// FindOneAndReplace atomically finds a document, replaces it, and returns
// either the original or the replacement document based on options.
func (col *Collection) FindOneAndReplace(ctx context.Context, filterBuilder *filter.Builder, replacement any, opts ...*FindOneAndReplaceOptions) *FindOneResult {
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

	// Convert our options to mongo driver options
	driverOpts := options.FindOneAndReplace()

	if len(opts) > 0 && opts[0] != nil {
		opt := opts[0]

		if opt.ReturnDocument == ReturnAfter {
			driverOpts.SetReturnDocument(options.After)
		} else {
			driverOpts.SetReturnDocument(options.Before)
		}

		if opt.Upsert {
			driverOpts.SetUpsert(true)
		}

		if len(opt.Sort) > 0 {
			driverOpts.SetSort(opt.Sort)
		}

		if len(opt.Projection) > 0 {
			driverOpts.SetProjection(opt.Projection)
		}
	}

	col.client.config.Logger.Debug("FindOneAndReplace",
		"collection", col.name)

	result := col.collection.FindOneAndReplace(ctx, filterDoc, replacement, driverOpts)

	col.client.incrementOperationCount()

	return &FindOneResult{
		result: result,
	}
}

// FindOneAndDelete atomically finds a document and deletes it, returning the deleted document.
// This is useful for queue-like operations where you need to atomically claim and remove an item.
func (col *Collection) FindOneAndDelete(ctx context.Context, filterBuilder *filter.Builder, opts ...*FindOneAndDeleteOptions) *FindOneResult {
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

	// Convert our options to mongo driver options
	driverOpts := options.FindOneAndDelete()

	if len(opts) > 0 && opts[0] != nil {
		opt := opts[0]

		if len(opt.Sort) > 0 {
			driverOpts.SetSort(opt.Sort)
		}

		if len(opt.Projection) > 0 {
			driverOpts.SetProjection(opt.Projection)
		}
	}

	col.client.config.Logger.Debug("FindOneAndDelete",
		"collection", col.name)

	result := col.collection.FindOneAndDelete(ctx, filterDoc, driverOpts)

	col.client.incrementOperationCount()

	return &FindOneResult{
		result: result,
	}
}

// Helper functions

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

// =============================================================================
// ID-Based Helper Methods
// =============================================================================

// FindByID finds a single document by its _id field.
// This is a convenience method that works with any ID type (ULID string, ObjectID, etc.).
func (col *Collection) FindByID(ctx context.Context, id any) *FindOneResult {
	return col.FindOne(ctx, filter.Eq("_id", id))
}

// UpdateByID updates a single document by its _id field.
// This is a convenience method that works with any ID type (ULID string, ObjectID, etc.).
func (col *Collection) UpdateByID(ctx context.Context, id any, updateBuilder *update.Builder) (*UpdateResult, error) {
	return col.UpdateOne(ctx, filter.Eq("_id", id), updateBuilder)
}

// DeleteByID deletes a single document by its _id field.
// This is a convenience method that works with any ID type (ULID string, ObjectID, etc.).
func (col *Collection) DeleteByID(ctx context.Context, id any) (*DeleteResult, error) {
	return col.DeleteOne(ctx, filter.Eq("_id", id))
}
