package pipeline

import (
	"testing"

	"github.com/cloudresty/go-mongodb/v2/filter"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestNew(t *testing.T) {
	builder := New()
	if builder == nil {
		t.Error("New() should return a non-nil builder")
		return
	}
	if len(builder.stages) != 0 {
		t.Error("New builder should have empty stages")
	}
}

func TestMatch(t *testing.T) {
	// Test with filter builder
	filterBuilder := filter.Eq("status", "active")
	pipeline := New().Match(filterBuilder)

	stages := pipeline.Build()
	if len(stages) != 1 {
		t.Errorf("Expected 1 stage, got %d", len(stages))
	}

	matchStage, ok := stages[0]["$match"]
	if !ok {
		t.Error("Expected $match stage")
	}

	matchDoc, ok := matchStage.(bson.M)
	if !ok {
		t.Error("Match stage should be bson.M")
	}

	if matchDoc["status"] != "active" {
		t.Errorf("Expected status=active, got %v", matchDoc["status"])
	}

	// Test with nil filter
	pipeline2 := New().Match(nil)
	stages2 := pipeline2.Build()
	if len(stages2) != 1 {
		t.Errorf("Expected 1 stage, got %d", len(stages2))
	}
}

func TestMatchRaw(t *testing.T) {
	filter := bson.M{"age": bson.M{"$gte": 18}}
	pipeline := New().MatchRaw(filter)

	stages := pipeline.Build()
	if len(stages) != 1 {
		t.Errorf("Expected 1 stage, got %d", len(stages))
	}

	matchStage := stages[0]["$match"].(bson.M)
	ageFilter := matchStage["age"].(bson.M)
	if ageFilter["$gte"] != 18 {
		t.Errorf("Expected age.$gte=18, got %v", ageFilter["$gte"])
	}
}

func TestProject(t *testing.T) {
	projection := bson.M{"name": 1, "email": 1, "_id": 0}
	pipeline := New().Project(projection)

	stages := pipeline.Build()
	if len(stages) != 1 {
		t.Errorf("Expected 1 stage, got %d", len(stages))
	}

	projectStage := stages[0]["$project"].(bson.M)
	if projectStage["name"] != 1 {
		t.Errorf("Expected name=1, got %v", projectStage["name"])
	}
	if projectStage["_id"] != 0 {
		t.Errorf("Expected _id=0, got %v", projectStage["_id"])
	}
}

func TestSort(t *testing.T) {
	sorts := bson.D{
		{Key: "created_at", Value: -1},
		{Key: "name", Value: 1},
	}
	pipeline := New().Sort(sorts)

	stages := pipeline.Build()
	if len(stages) != 1 {
		t.Errorf("Expected 1 stage, got %d", len(stages))
	}

	sortStage := stages[0]["$sort"].(bson.D)
	if len(sortStage) != 2 {
		t.Errorf("Expected 2 sort fields, got %d", len(sortStage))
	}
	if sortStage[0].Key != "created_at" || sortStage[0].Value != -1 {
		t.Errorf("Expected created_at=-1, got %s=%v", sortStage[0].Key, sortStage[0].Value)
	}
}

func TestSortMap(t *testing.T) {
	sorts := map[string]int{
		"created_at": -1,
		"name":       1,
	}
	pipeline := New().SortMap(sorts)

	stages := pipeline.Build()
	if len(stages) != 1 {
		t.Errorf("Expected 1 stage, got %d", len(stages))
	}

	sortStage := stages[0]["$sort"].(bson.D)
	if len(sortStage) != 2 {
		t.Errorf("Expected 2 sort fields, got %d", len(sortStage))
	}
}

func TestLimit(t *testing.T) {
	pipeline := New().Limit(10)

	stages := pipeline.Build()
	if len(stages) != 1 {
		t.Errorf("Expected 1 stage, got %d", len(stages))
	}

	limitStage := stages[0]["$limit"]
	if limitStage != int64(10) {
		t.Errorf("Expected limit=10, got %v", limitStage)
	}
}

func TestSkip(t *testing.T) {
	pipeline := New().Skip(5)

	stages := pipeline.Build()
	if len(stages) != 1 {
		t.Errorf("Expected 1 stage, got %d", len(stages))
	}

	skipStage := stages[0]["$skip"]
	if skipStage != int64(5) {
		t.Errorf("Expected skip=5, got %v", skipStage)
	}
}

func TestGroup(t *testing.T) {
	pipeline := New().Group("$category", bson.M{
		"total":     bson.M{"$sum": 1},
		"avg_price": bson.M{"$avg": "$price"},
	})

	stages := pipeline.Build()
	if len(stages) != 1 {
		t.Errorf("Expected 1 stage, got %d", len(stages))
	}

	groupStage := stages[0]["$group"].(bson.M)
	if groupStage["_id"] != "$category" {
		t.Errorf("Expected _id=$category, got %v", groupStage["_id"])
	}

	total := groupStage["total"].(bson.M)
	if total["$sum"] != 1 {
		t.Errorf("Expected total.$sum=1, got %v", total["$sum"])
	}
}

func TestLookup(t *testing.T) {
	pipeline := New().Lookup("orders", "user_id", "_id", "userOrders")

	stages := pipeline.Build()
	if len(stages) != 1 {
		t.Errorf("Expected 1 stage, got %d", len(stages))
	}

	lookupStage := stages[0]["$lookup"].(bson.M)
	if lookupStage["from"] != "orders" {
		t.Errorf("Expected from=orders, got %v", lookupStage["from"])
	}
	if lookupStage["localField"] != "user_id" {
		t.Errorf("Expected localField=user_id, got %v", lookupStage["localField"])
	}
	if lookupStage["foreignField"] != "_id" {
		t.Errorf("Expected foreignField=_id, got %v", lookupStage["foreignField"])
	}
	if lookupStage["as"] != "userOrders" {
		t.Errorf("Expected as=userOrders, got %v", lookupStage["as"])
	}
}

func TestUnwind(t *testing.T) {
	pipeline := New().Unwind("$tags")

	stages := pipeline.Build()
	if len(stages) != 1 {
		t.Errorf("Expected 1 stage, got %d", len(stages))
	}

	unwindStage := stages[0]["$unwind"]
	if unwindStage != "$tags" {
		t.Errorf("Expected $unwind=$tags, got %v", unwindStage)
	}
}

func TestUnwindWithOptions(t *testing.T) {
	pipeline := New().UnwindWithOptions("$tags", true, "tagIndex")

	stages := pipeline.Build()
	if len(stages) != 1 {
		t.Errorf("Expected 1 stage, got %d", len(stages))
	}

	unwindStage := stages[0]["$unwind"].(bson.M)
	if unwindStage["path"] != "$tags" {
		t.Errorf("Expected path=$tags, got %v", unwindStage["path"])
	}
	if unwindStage["preserveNullAndEmptyArrays"] != true {
		t.Errorf("Expected preserveNullAndEmptyArrays=true, got %v", unwindStage["preserveNullAndEmptyArrays"])
	}
	if unwindStage["includeArrayIndex"] != "tagIndex" {
		t.Errorf("Expected includeArrayIndex=tagIndex, got %v", unwindStage["includeArrayIndex"])
	}
}

func TestChaining(t *testing.T) {
	filterBuilder := filter.Eq("status", "active")
	pipeline := New().
		Match(filterBuilder).
		Project(bson.M{"name": 1, "email": 1}).
		Sort(bson.D{{Key: "created_at", Value: -1}}).
		Limit(10).
		Skip(5)

	stages := pipeline.Build()
	if len(stages) != 5 {
		t.Errorf("Expected 5 stages, got %d", len(stages))
	}

	// Verify stage order
	expectedStages := []string{"$match", "$project", "$sort", "$limit", "$skip"}
	for i, expectedStage := range expectedStages {
		if _, ok := stages[i][expectedStage]; !ok {
			t.Errorf("Expected stage %d to be %s", i, expectedStage)
		}
	}
}

func TestToBSONArray(t *testing.T) {
	pipeline := New().
		MatchRaw(bson.M{"status": "active"}).
		Limit(10)

	bsonArray := pipeline.ToBSONArray()
	if len(bsonArray) != 2 {
		t.Errorf("Expected 2 elements in bson.A, got %d", len(bsonArray))
	}

	// Check first element
	firstStage, ok := bsonArray[0].(bson.M)
	if !ok {
		t.Error("First element should be bson.M")
	}
	if _, ok := firstStage["$match"]; !ok {
		t.Error("First stage should be $match")
	}

	// Check second element
	secondStage, ok := bsonArray[1].(bson.M)
	if !ok {
		t.Error("Second element should be bson.M")
	}
	if _, ok := secondStage["$limit"]; !ok {
		t.Error("Second stage should be $limit")
	}
}

func TestStandaloneFunctions(t *testing.T) {
	// Test Match standalone function
	filterBuilder := filter.Eq("status", "active")
	pipeline := Match(filterBuilder)

	stages := pipeline.Build()
	if len(stages) != 1 {
		t.Errorf("Expected 1 stage, got %d", len(stages))
	}

	// Test chaining with standalone function
	pipeline = Match(filterBuilder).Limit(10)
	stages = pipeline.Build()
	if len(stages) != 2 {
		t.Errorf("Expected 2 stages, got %d", len(stages))
	}
}

func TestComplexPipeline(t *testing.T) {
	// Complex aggregation pipeline example
	filterBuilder := filter.Eq("status", "active").And(filter.Gt("age", 18))

	pipeline := New().
		Match(filterBuilder).
		Lookup("orders", "user_id", "_id", "userOrders").
		Unwind("$userOrders").
		Group("$category", bson.M{
			"totalOrders": bson.M{"$sum": 1},
			"avgAmount":   bson.M{"$avg": "$userOrders.amount"},
		}).
		Sort(bson.D{{Key: "totalOrders", Value: -1}}).
		Limit(5)

	stages := pipeline.Build()
	if len(stages) != 6 {
		t.Errorf("Expected 6 stages, got %d", len(stages))
	}

	// Verify the pipeline structure
	expectedStages := []string{"$match", "$lookup", "$unwind", "$group", "$sort", "$limit"}
	for i, expectedStage := range expectedStages {
		if _, ok := stages[i][expectedStage]; !ok {
			t.Errorf("Expected stage %d to be %s", i, expectedStage)
		}
	}
}
