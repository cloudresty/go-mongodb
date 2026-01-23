package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudresty/go-mongodb/v2"
	"github.com/cloudresty/go-mongodb/v2/filter"
	"github.com/cloudresty/go-mongodb/v2/pipeline"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type User struct {
	ID         string    `bson:"_id,omitempty"`
	Name       string    `bson:"name"`
	Email      string    `bson:"email"`
	Age        int       `bson:"age"`
	Status     string    `bson:"status"`
	Department string    `bson:"department"`
	Salary     float64   `bson:"salary"`
	CreatedAt  time.Time `bson:"created_at"`
	Tags       []string  `bson:"tags"`
}

func main() {
	fmt.Println("🚀 Go-MongoDB Enhanced Features Demo")
	fmt.Println("=====================================")

	// Initialize client
	client, err := mongodb.NewClient(
		mongodb.WithHosts("localhost:27017"),
		mongodb.WithDatabase("enhanced_demo"),
		mongodb.WithAppName("enhanced-features-demo"),
	)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Failed to close client: %v", err)
		}
	}()

	collection := client.Collection("users")
	ctx := context.Background()

	// Seed some sample data
	seedSampleData(collection, ctx)

	fmt.Println("\n📊 Running Enhanced Features Examples...")

	// 1. Enhanced Find Operations with QueryOptions
	demonstrateQueryOptions(collection, ctx)

	// 2. Convenience Sort Methods
	demonstrateSortMethods(collection, ctx)

	// 3. Pipeline Builder
	demonstratePipelineBuilder(collection, ctx)

	// 4. Advanced Aggregation Pipelines
	demonstrateAdvancedPipelines(collection, ctx)

	// 5. Pagination Examples
	demonstratePagination(collection, ctx)

	fmt.Println("\n✅ All Enhanced Features Demonstrated Successfully!")
}

func seedSampleData(collection *mongodb.Collection, ctx context.Context) {
	fmt.Println("\n🌱 Seeding sample data...")

	users := []any{
		User{
			Name: "Alice Johnson", Email: "alice@company.com", Age: 28,
			Status: "active", Department: "Engineering", Salary: 85000,
			CreatedAt: time.Now().AddDate(0, -6, 0), Tags: []string{"senior", "backend"},
		},
		User{
			Name: "Bob Smith", Email: "bob@company.com", Age: 32,
			Status: "active", Department: "Engineering", Salary: 95000,
			CreatedAt: time.Now().AddDate(0, -12, 0), Tags: []string{"lead", "fullstack"},
		},
		User{
			Name: "Carol Davis", Email: "carol@company.com", Age: 29,
			Status: "active", Department: "Marketing", Salary: 70000,
			CreatedAt: time.Now().AddDate(0, -3, 0), Tags: []string{"creative", "social"},
		},
		User{
			Name: "David Wilson", Email: "david@company.com", Age: 26,
			Status: "inactive", Department: "Engineering", Salary: 80000,
			CreatedAt: time.Now().AddDate(0, -1, 0), Tags: []string{"junior", "frontend"},
		},
		User{
			Name: "Eve Brown", Email: "eve@company.com", Age: 35,
			Status: "active", Department: "Sales", Salary: 90000,
			CreatedAt: time.Now().AddDate(0, -24, 0), Tags: []string{"senior", "enterprise"},
		},
	}

	// Clear existing data first
	_, err := collection.DeleteMany(ctx, filter.New())
	if err != nil {
		log.Printf("Warning: Could not clear existing data: %v", err)
	}

	// Insert sample data
	_, err = collection.InsertMany(ctx, users)
	if err != nil {
		log.Printf("Warning: Could not seed data: %v", err)
	} else {
		fmt.Printf("✅ Seeded %d sample users\n", len(users))
	}
}

func demonstrateQueryOptions(collection *mongodb.Collection, ctx context.Context) {
	fmt.Println("\n🔍 1. QueryOptions Examples")
	fmt.Println("-----------------------------")

	// Example 1: Complex query with all options
	queryOpts := &mongodb.QueryOptions{
		Sort: bson.D{
			{Key: "salary", Value: -1},    // Highest salary first
			{Key: "created_at", Value: 1}, // Then oldest first
		},
		Limit: func() *int64 { l := int64(3); return &l }(),
		Skip:  func() *int64 { s := int64(0); return &s }(),
		Projection: bson.D{
			{Key: "name", Value: 1},
			{Key: "department", Value: 1},
			{Key: "salary", Value: 1},
			{Key: "_id", Value: 0},
		},
	}

	fmt.Println("📋 Top 3 earners (name, department, salary only):")
	result, err := collection.FindWithOptions(ctx,
		filter.Eq("status", "active"), queryOpts)

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer func() {
		if err := result.Close(ctx); err != nil {
			log.Printf("Failed to close result cursor: %v", err)
		}
	}()

	for result.Next(ctx) {
		var user bson.M
		if err := result.Decode(&user); err != nil {
			log.Printf("Decode error: %v", err)
			continue
		}
		fmt.Printf("  • %s (%s): $%.0f\n",
			user["name"], user["department"], user["salary"])
	}

	// Example 2: Single document with options
	fmt.Println("\n👤 Newest active user:")
	newestOpts := &mongodb.QueryOptions{
		Sort: bson.D{{Key: "created_at", Value: -1}},
		Projection: bson.D{
			{Key: "name", Value: 1},
			{Key: "created_at", Value: 1},
			{Key: "_id", Value: 0},
		},
	}

	newestResult := collection.FindOneWithOptions(ctx,
		filter.Eq("status", "active"), newestOpts)

	var newest bson.M
	if err := newestResult.Decode(&newest); err != nil {
		log.Printf("No active users found: %v", err)
	} else {
		fmt.Printf("  • %s (joined: %v)\n",
			newest["name"], newest["created_at"])
	}
}

func demonstrateSortMethods(collection *mongodb.Collection, ctx context.Context) {
	fmt.Println("\n📈 2. Convenience Sort Methods")
	fmt.Println("--------------------------------")

	// FindSorted - Multiple active users by age (oldest first)
	fmt.Println("👥 Active users by age (oldest first):")
	sortedResult, err := collection.FindSorted(ctx,
		filter.Eq("status", "active"),
		bson.D{{Key: "age", Value: -1}})

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer func() {
		if err := sortedResult.Close(ctx); err != nil {
			log.Printf("Failed to close sortedResult cursor: %v", err)
		}
	}()

	for sortedResult.Next(ctx) {
		var user User
		if err := sortedResult.Decode(&user); err != nil {
			log.Printf("Decode error: %v", err)
			continue
		}
		fmt.Printf("  • %s: %d years old\n", user.Name, user.Age)
	}

	// FindOneSorted - Oldest active user
	fmt.Println("\n🧓 Oldest active user:")
	oldestResult := collection.FindOneSorted(ctx,
		filter.Eq("status", "active"),
		bson.D{{Key: "age", Value: -1}})

	var oldest User
	if err := oldestResult.Decode(&oldest); err != nil {
		log.Printf("No users found: %v", err)
	} else {
		fmt.Printf("  • %s (%d years old)\n", oldest.Name, oldest.Age)
	}

	// FindWithLimit - Limited results
	fmt.Println("\n🔢 First 2 active users:")
	limitResult, err := collection.FindWithLimit(ctx,
		filter.Eq("status", "active"), 2)

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer func() {
		if err := limitResult.Close(ctx); err != nil {
			log.Printf("Failed to close limitResult cursor: %v", err)
		}
	}()

	count := 0
	for limitResult.Next(ctx) {
		count++
		var user User
		if err := limitResult.Decode(&user); err != nil {
			log.Printf("Decode error: %v", err)
			continue
		}
		fmt.Printf("  %d. %s\n", count, user.Name)
	}
}

func demonstratePipelineBuilder(collection *mongodb.Collection, ctx context.Context) {
	fmt.Println("\n🏗️  3. Pipeline Builder Examples")
	fmt.Println("----------------------------------")

	// Example 1: Department statistics
	fmt.Println("📊 Department Statistics:")

	deptPipeline := pipeline.New().
		Match(filter.Eq("status", "active")).
		Group("$department", bson.M{
			"count":       bson.M{"$sum": 1},
			"avgSalary":   bson.M{"$avg": "$salary"},
			"totalSalary": bson.M{"$sum": "$salary"},
		}).
		Sort(bson.D{{Key: "count", Value: -1}})

	deptResult, err := collection.AggregateWithPipeline(ctx, deptPipeline)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer func() {
		if err := deptResult.Close(ctx); err != nil {
			log.Printf("Failed to close deptResult cursor: %v", err)
		}
	}()

	for deptResult.Next(ctx) {
		var stats bson.M
		if err := deptResult.Decode(&stats); err != nil {
			log.Printf("Decode error: %v", err)
			continue
		}
		fmt.Printf("  • %s: %v people, avg $%.0f, total $%.0f\n",
			stats["_id"], stats["count"], stats["avgSalary"], stats["totalSalary"])
	}

	// Example 2: User list with specific fields
	fmt.Println("\n📋 User Contact List (using pipeline):")

	contactPipeline := pipeline.New().
		Match(filter.Eq("status", "active")).
		Project(bson.M{
			"name":  1,
			"email": 1,
			"dept":  "$department",
			"_id":   0,
		}).
		Sort(bson.D{{Key: "name", Value: 1}})

	contactResult, err := collection.AggregateWithPipeline(ctx, contactPipeline)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer func() {
		if err := contactResult.Close(ctx); err != nil {
			log.Printf("Failed to close contactResult cursor: %v", err)
		}
	}()

	for contactResult.Next(ctx) {
		var contact bson.M
		if err := contactResult.Decode(&contact); err != nil {
			log.Printf("Decode error: %v", err)
			continue
		}
		fmt.Printf("  • %s <%s> (%s)\n",
			contact["name"], contact["email"], contact["dept"])
	}
}

func demonstrateAdvancedPipelines(collection *mongodb.Collection, ctx context.Context) {
	fmt.Println("\n🧠 4. Advanced Pipeline Examples")
	fmt.Println("-----------------------------------")

	// Complex pipeline with multiple transformations
	fmt.Println("🏆 Salary Analysis by Experience Level:")

	analysisPipeline := pipeline.New().
		Match(filter.Eq("status", "active")).
		AddFields(bson.M{
			"experienceLevel": bson.M{
				"$switch": bson.M{
					"branches": []bson.M{
						{"case": bson.M{"$gte": []any{"$age", 35}}, "then": "Senior"},
						{"case": bson.M{"$gte": []any{"$age", 30}}, "then": "Mid-level"},
					},
					"default": "Junior",
				},
			},
		}).
		Group("$experienceLevel", bson.M{
			"count":     bson.M{"$sum": 1},
			"avgSalary": bson.M{"$avg": "$salary"},
			"minSalary": bson.M{"$min": "$salary"},
			"maxSalary": bson.M{"$max": "$salary"},
		}).
		Sort(bson.D{{Key: "avgSalary", Value: -1}})

	analysisResult, err := collection.AggregateWithPipeline(ctx, analysisPipeline)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer func() {
		if err := analysisResult.Close(ctx); err != nil {
			log.Printf("Failed to close analysisResult cursor: %v", err)
		}
	}()

	for analysisResult.Next(ctx) {
		var analysis bson.M
		if err := analysisResult.Decode(&analysis); err != nil {
			log.Printf("Decode error: %v", err)
			continue
		}
		fmt.Printf("  • %s: %v people, avg $%.0f (range: $%.0f - $%.0f)\n",
			analysis["_id"], analysis["count"], analysis["avgSalary"],
			analysis["minSalary"], analysis["maxSalary"])
	}

	// Faceted analysis
	fmt.Println("\n📈 Multi-dimensional Analysis:")

	facetPipeline := pipeline.New().
		Match(filter.Eq("status", "active")).
		Facet(map[string][]bson.M{
			"byDepartment": pipeline.Group("$department", bson.M{
				"count": bson.M{"$sum": 1},
			}).Sort(bson.D{{Key: "count", Value: -1}}).Build(),

			"salaryStats": pipeline.Group(nil, bson.M{
				"avgSalary":      bson.M{"$avg": "$salary"},
				"totalEmployees": bson.M{"$sum": 1},
			}).Build(),
		})

	facetResult, err := collection.AggregateWithPipeline(ctx, facetPipeline)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer func() {
		if err := facetResult.Close(ctx); err != nil {
			log.Printf("Failed to close facetResult cursor: %v", err)
		}
	}()

	if facetResult.Next(ctx) {
		var facetData bson.M
		if err := facetResult.Decode(&facetData); err != nil {
			log.Printf("Decode error: %v", err)
		} else {
			fmt.Printf("  • Faceted analysis completed (departments + salary stats)\n")
			// In a real application, you'd process the faceted results
		}
	}
}

func demonstratePagination(collection *mongodb.Collection, ctx context.Context) {
	fmt.Println("\n📄 5. Pagination Examples")
	fmt.Println("---------------------------")

	pageSize := int64(2)
	fmt.Printf("📖 Paginated results (page size: %d):\n", pageSize)

	// Page 1
	fmt.Println("\nPage 1:")
	page1Opts := &mongodb.QueryOptions{
		Sort:  bson.D{{Key: "name", Value: 1}},
		Limit: &pageSize,
		Skip:  func() *int64 { s := int64(0); return &s }(),
	}

	page1Result, err := collection.FindWithOptions(ctx,
		filter.Eq("status", "active"), page1Opts)

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer func() {
		if err := page1Result.Close(ctx); err != nil {
			log.Printf("Failed to close page1Result cursor: %v", err)
		}
	}()

	for page1Result.Next(ctx) {
		var user User
		if err := page1Result.Decode(&user); err != nil {
			log.Printf("Decode error: %v", err)
			continue
		}
		fmt.Printf("  • %s (%s)\n", user.Name, user.Email)
	}

	// Page 2
	fmt.Println("\nPage 2:")
	page2Opts := &mongodb.QueryOptions{
		Sort:  bson.D{{Key: "name", Value: 1}},
		Limit: &pageSize,
		Skip:  func() *int64 { s := int64(2); return &s }(),
	}

	page2Result, err := collection.FindWithOptions(ctx,
		filter.Eq("status", "active"), page2Opts)

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer func() {
		if err := page2Result.Close(ctx); err != nil {
			log.Printf("Failed to close page2Result cursor: %v", err)
		}
	}()

	for page2Result.Next(ctx) {
		var user User
		if err := page2Result.Decode(&user); err != nil {
			log.Printf("Decode error: %v", err)
			continue
		}
		fmt.Printf("  • %s (%s)\n", user.Name, user.Email)
	}

	// Using skip convenience method
	fmt.Println("\nUsing FindWithSkip for page 3:")
	page3Result, err := collection.FindWithSkip(ctx,
		filter.Eq("status", "active"), 4)

	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer func() {
		if err := page3Result.Close(ctx); err != nil {
			log.Printf("Failed to close page3Result cursor: %v", err)
		}
	}()

	found := false
	for page3Result.Next(ctx) {
		found = true
		var user User
		if err := page3Result.Decode(&user); err != nil {
			log.Printf("Decode error: %v", err)
			continue
		}
		fmt.Printf("  • %s (%s)\n", user.Name, user.Email)
	}

	if !found {
		fmt.Println("  (No more users)")
	}
}
