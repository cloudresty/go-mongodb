package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/cloudresty/go-mongodb/v2"
	"github.com/cloudresty/go-mongodb/v2/filter"
)

func main() {
	// Connect to MongoDB
	client, err := mongodb.NewClient(mongodb.FromEnv())
	if err != nil {
		log.Fatal("Failed to create MongoDB client:", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Error closing client: %v", err)
		}
	}()

	// Get database and collection
	db := client.Database("bson_convenience_demo")
	col := db.Collection("users")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Clean up any existing data
	_, _ = col.DeleteMany(ctx, filter.New())

	fmt.Println("🧮 BSON Convenience Features Demo")
	fmt.Println("=================================")

	// Demo 1: Using Document() convenience function for easier document creation
	fmt.Println("\n1. Creating documents with Document() convenience function:")

	users := []any{
		mongodb.Document("name", "Alice", "age", 25, "department", "Engineering", "salary", 85000),
		mongodb.Document("name", "Bob", "age", 30, "department", "Marketing", "salary", 70000),
		mongodb.Document("name", "Charlie", "age", 35, "department", "Engineering", "salary", 95000),
		mongodb.Document("name", "Diana", "age", 28, "department", "Sales", "salary", 75000),
		mongodb.Document("name", "Eva", "age", 32, "department", "Engineering", "salary", 90000),
	}

	_, err = col.InsertMany(ctx, users)
	if err != nil {
		log.Fatal("Failed to insert users:", err)
	}
	fmt.Printf("✅ Inserted %d users using Document() helper\n", len(users))

	// Demo 2: Simple sort functions
	fmt.Println("\n2. Using SortAsc() and SortDesc() convenience functions:")

	result, err := col.FindAscending(ctx, filter.Eq("department", "Engineering"), "age")
	if err != nil {
		log.Fatal("Failed to find ascending:", err)
	}
	defer func() {
		if err := result.Close(ctx); err != nil {
			log.Printf("Error closing result: %v", err)
		}
	}()

	var engineeringUsers []mongodb.M
	if err := result.All(ctx, &engineeringUsers); err != nil {
		log.Fatal("Failed to decode results:", err)
	}

	fmt.Println("👨‍💻 Engineering users sorted by age (ascending):")
	for _, user := range engineeringUsers {
		fmt.Printf("   - %s, age %v\n", user["name"], user["age"])
	}

	// Demo 3: Using map-based sorting for flexibility
	fmt.Println("\n3. Using map[string]int for flexible sorting:")

	sortMap := map[string]int{
		"salary": -1, // Descending by salary
	}

	result2, err := col.FindSorted(ctx, filter.New(), sortMap)
	if err != nil {
		log.Fatal("Failed to find with map sort:", err)
	}
	defer func() {
		if err := result2.Close(ctx); err != nil {
			log.Printf("Error closing result2: %v", err)
		}
	}()

	var allUsers []mongodb.M
	if err := result2.All(ctx, &allUsers); err != nil {
		log.Fatal("Failed to decode results:", err)
	}

	fmt.Println("💰 All users sorted by salary (descending):")
	for _, user := range allUsers {
		fmt.Printf("   - %s: $%v\n", user["name"], user["salary"])
	}

	// Demo 4: Complex multi-field sorting with SortMultipleOrdered
	fmt.Println("\n4. Using SortMultipleOrdered() for complex sorting:")

	complexSort := mongodb.SortMultipleOrdered("department", 1, "salary", -1)

	result3, err := col.FindSorted(ctx, filter.New(), complexSort)
	if err != nil {
		log.Fatal("Failed to find with complex sort:", err)
	}
	defer func() {
		if err := result3.Close(ctx); err != nil {
			log.Printf("Error closing result3: %v", err)
		}
	}()

	var sortedUsers []mongodb.M
	if err := result3.All(ctx, &sortedUsers); err != nil {
		log.Fatal("Failed to decode results:", err)
	}

	fmt.Println("🏢 Users sorted by department (asc), then salary (desc):")
	currentDept := ""
	for _, user := range sortedUsers {
		dept := user["department"].(string)
		if dept != currentDept {
			fmt.Printf("\n📁 %s:\n", dept)
			currentDept = dept
		}
		fmt.Printf("   - %s: $%v\n", user["name"], user["salary"])
	}

	// Demo 5: Projection convenience functions
	fmt.Println("\n5. Using Include() and Exclude() for projections:")

	result4, err := col.FindWithProjectionFields(ctx,
		filter.Eq("department", "Engineering"),
		[]string{"name", "salary"}, // Include these fields
		[]string{"_id"},            // Exclude _id
	)
	if err != nil {
		log.Fatal("Failed to find with projection:", err)
	}
	defer func() {
		if err := result4.Close(ctx); err != nil {
			log.Printf("Error closing result4: %v", err)
		}
	}()

	var projectedUsers []mongodb.M
	if err := result4.All(ctx, &projectedUsers); err != nil {
		log.Fatal("Failed to decode results:", err)
	}

	fmt.Println("👀 Engineering users with only name and salary (no _id):")
	for _, user := range projectedUsers {
		fmt.Printf("   - Name: %s, Salary: $%v\n", user["name"], user["salary"])
		if _, hasID := user["_id"]; hasID {
			fmt.Printf("     ⚠️  ERROR: _id should be excluded!\n")
		}
		if _, hasAge := user["age"]; hasAge {
			fmt.Printf("     ⚠️  ERROR: age should not be included!\n")
		}
	}

	// Demo 6: Using re-exported BSON types for cleaner code
	fmt.Println("\n6. Using re-exported BSON types (mongodb.D, mongodb.M):")

	// Instead of importing bson separately, use mongodb.D and mongodb.M
	manualSort := mongodb.D{
		{Key: "age", Value: 1},
		{Key: "name", Value: 1},
	}

	result5, err := col.FindSorted(ctx, filter.Eq("department", "Sales"), manualSort)
	if err != nil {
		log.Fatal("Failed to find with manual sort:", err)
	}
	defer func() {
		if err := result5.Close(ctx); err != nil {
			log.Printf("Error closing result5: %v", err)
		}
	}()

	var salesUsers []mongodb.M
	if err := result5.All(ctx, &salesUsers); err != nil {
		log.Fatal("Failed to decode results:", err)
	}

	fmt.Println("🛒 Sales users sorted by age then name (using mongodb.D):")
	for _, user := range salesUsers {
		fmt.Printf("   - %s, age %v\n", user["name"], user["age"])
	}

	// Demo 7: Find one with convenience methods
	fmt.Println("\n7. Using FindOneAscending() and FindOneDescending():")

	youngest := col.FindOneAscending(ctx, filter.New(), "age")
	var youngestUser mongodb.M
	if err := youngest.Decode(&youngestUser); err != nil {
		log.Fatal("Failed to decode youngest:", err)
	}

	oldest := col.FindOneDescending(ctx, filter.New(), "age")
	var oldestUser mongodb.M
	if err := oldest.Decode(&oldestUser); err != nil {
		log.Fatal("Failed to decode oldest:", err)
	}

	fmt.Printf("👶 Youngest user: %s (age %v)\n", youngestUser["name"], youngestUser["age"])
	fmt.Printf("👴 Oldest user: %s (age %v)\n", oldestUser["name"], oldestUser["age"])

	fmt.Println("\n✨ BSON convenience features make MongoDB operations much cleaner!")
	fmt.Println("   • No need to import bson separately")
	fmt.Println("   • Flexible sorting with maps or convenience functions")
	fmt.Println("   • Easy document creation with Document()")
	fmt.Println("   • Clean projection syntax with Include()/Exclude()")
	fmt.Println("   • Type-safe SortSpec interface accepts multiple formats")
}
