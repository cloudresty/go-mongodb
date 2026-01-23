package pipeline

import (
	"github.com/cloudresty/go-mongodb/v2/filter"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Builder represents a fluent aggregation pipeline builder
type Builder struct {
	stages []bson.M
}

// New creates a new pipeline builder
func New() *Builder {
	return &Builder{
		stages: make([]bson.M, 0),
	}
}

// Build returns the pipeline as a slice of bson.M
func (b *Builder) Build() []bson.M {
	return b.stages
}

// ToBSONArray converts the pipeline to a bson.A for compatibility
func (b *Builder) ToBSONArray() bson.A {
	result := make(bson.A, len(b.stages))
	for i, stage := range b.stages {
		result[i] = stage
	}
	return result
}

// Match adds a $match stage to the pipeline
func (b *Builder) Match(filterBuilder *filter.Builder) *Builder {
	var matchDoc bson.M
	if filterBuilder != nil {
		matchDoc = filterBuilder.Build()
	} else {
		matchDoc = bson.M{}
	}

	b.stages = append(b.stages, bson.M{"$match": matchDoc})
	return b
}

// MatchRaw adds a $match stage with raw bson.M filter
func (b *Builder) MatchRaw(filter bson.M) *Builder {
	b.stages = append(b.stages, bson.M{"$match": filter})
	return b
}

// Project adds a $project stage to the pipeline
func (b *Builder) Project(fields bson.M) *Builder {
	b.stages = append(b.stages, bson.M{"$project": fields})
	return b
}

// Sort adds a $sort stage to the pipeline
func (b *Builder) Sort(sorts bson.D) *Builder {
	b.stages = append(b.stages, bson.M{"$sort": sorts})
	return b
}

// SortMap adds a $sort stage using a map (converted to bson.D for ordered sorting)
func (b *Builder) SortMap(sorts map[string]int) *Builder {
	sortDoc := make(bson.D, 0, len(sorts))
	for field, order := range sorts {
		sortDoc = append(sortDoc, bson.E{Key: field, Value: order})
	}
	return b.Sort(sortDoc)
}

// Limit adds a $limit stage to the pipeline
func (b *Builder) Limit(limit int64) *Builder {
	b.stages = append(b.stages, bson.M{"$limit": limit})
	return b
}

// Skip adds a $skip stage to the pipeline
func (b *Builder) Skip(skip int64) *Builder {
	b.stages = append(b.stages, bson.M{"$skip": skip})
	return b
}

// Group adds a $group stage to the pipeline
func (b *Builder) Group(id any, fields bson.M) *Builder {
	groupDoc := bson.M{"_id": id}
	for k, v := range fields {
		groupDoc[k] = v
	}
	b.stages = append(b.stages, bson.M{"$group": groupDoc})
	return b
}

// Lookup adds a $lookup stage to the pipeline
func (b *Builder) Lookup(from, localField, foreignField, as string) *Builder {
	b.stages = append(b.stages, bson.M{
		"$lookup": bson.M{
			"from":         from,
			"localField":   localField,
			"foreignField": foreignField,
			"as":           as,
		},
	})
	return b
}

// Unwind adds an $unwind stage to the pipeline
func (b *Builder) Unwind(path string) *Builder {
	b.stages = append(b.stages, bson.M{"$unwind": path})
	return b
}

// UnwindWithOptions adds an $unwind stage with additional options
func (b *Builder) UnwindWithOptions(path string, preserveNullAndEmptyArrays bool, includeArrayIndex string) *Builder {
	unwindDoc := bson.M{"path": path}
	if preserveNullAndEmptyArrays {
		unwindDoc["preserveNullAndEmptyArrays"] = true
	}
	if includeArrayIndex != "" {
		unwindDoc["includeArrayIndex"] = includeArrayIndex
	}
	b.stages = append(b.stages, bson.M{"$unwind": unwindDoc})
	return b
}

// AddFields adds an $addFields stage to the pipeline
func (b *Builder) AddFields(fields bson.M) *Builder {
	b.stages = append(b.stages, bson.M{"$addFields": fields})
	return b
}

// ReplaceRoot adds a $replaceRoot stage to the pipeline
func (b *Builder) ReplaceRoot(newRoot any) *Builder {
	b.stages = append(b.stages, bson.M{"$replaceRoot": bson.M{"newRoot": newRoot}})
	return b
}

// Facet adds a $facet stage to the pipeline
func (b *Builder) Facet(facets map[string][]bson.M) *Builder {
	b.stages = append(b.stages, bson.M{"$facet": facets})
	return b
}

// Count adds a $count stage to the pipeline
func (b *Builder) Count(field string) *Builder {
	b.stages = append(b.stages, bson.M{"$count": field})
	return b
}

// Sample adds a $sample stage to the pipeline
func (b *Builder) Sample(size int64) *Builder {
	b.stages = append(b.stages, bson.M{"$sample": bson.M{"size": size}})
	return b
}

// Raw adds a custom stage to the pipeline
func (b *Builder) Raw(stage bson.M) *Builder {
	b.stages = append(b.stages, stage)
	return b
}

// Helper functions for common pipeline operations

// Match creates a $match stage (standalone function)
func Match(filterBuilder *filter.Builder) *Builder {
	return New().Match(filterBuilder)
}

// MatchRaw creates a $match stage with raw filter (standalone function)
func MatchRaw(filter bson.M) *Builder {
	return New().MatchRaw(filter)
}

// Project creates a $project stage (standalone function)
func Project(fields bson.M) *Builder {
	return New().Project(fields)
}

// Sort creates a $sort stage (standalone function)
func Sort(sorts bson.D) *Builder {
	return New().Sort(sorts)
}

// SortMap creates a $sort stage with map (standalone function)
func SortMap(sorts map[string]int) *Builder {
	return New().SortMap(sorts)
}

// Limit creates a $limit stage (standalone function)
func Limit(limit int64) *Builder {
	return New().Limit(limit)
}

// Skip creates a $skip stage (standalone function)
func Skip(skip int64) *Builder {
	return New().Skip(skip)
}

// Group creates a $group stage (standalone function)
func Group(id any, fields bson.M) *Builder {
	return New().Group(id, fields)
}
