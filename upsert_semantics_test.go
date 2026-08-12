package mongodb

import (
	"os"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// These tests pin the update DOCUMENTS rather than round-tripping through a
// server, because the defect they guard against was never a server behaviour:
// the wrong operator was chosen in Go, and MongoDB then did exactly what it was
// asked. Asserting the operator is asserting the bug.
//
// The server semantics these rely on were verified against a live MongoDB on
// 2026-08-12:
//
//	$set carrying an _id equal to the stored one     -> accepted
//	$set carrying an _id different from the stored one -> ImmutableField error
//	$setOnInsert against an existing document        -> matched=1 modified=0, write discarded
//	$set (minus _id) + $setOnInsert{_id} on insert   -> document lands under the caller's _id
//	the same pair on update                          -> modifies, _id untouched

type probeDoc struct {
	ID     string `bson:"_id"`
	Name   string `bson:"name"`
	Status string `bson:"status"`
}

func operators(t *testing.T, doc bson.M) []string {
	t.Helper()
	ops := make([]string, 0, len(doc))
	for k := range doc {
		ops = append(ops, k)
	}
	return ops
}

// TestInsertIfAbsentIsSetOnInsertOnly pins the behaviour media-agenda-collector
// deliberately depends on: a matched document must not be touched.
//
// If this ever starts emitting $set, collected events would begin overwriting
// each other on re-collection instead of de-duplicating.
func TestInsertIfAbsentIsSetOnInsertOnly(t *testing.T) {
	builder, err := insertIfAbsentUpdate(probeDoc{ID: "a", Name: "n", Status: "s"})
	if err != nil {
		t.Fatalf("building update: %v", err)
	}
	doc := builder.Build()

	if _, present := doc["$setOnInsert"]; !present {
		t.Fatalf("update has no $setOnInsert, so it would modify an existing document: %v", operators(t, doc))
	}
	if _, present := doc["$set"]; present {
		t.Errorf("update contains $set — an insert-if-absent write must never modify a matched document: %v", operators(t, doc))
	}
	if len(doc) != 1 {
		t.Errorf("update carries operators beyond $setOnInsert: %v", operators(t, doc))
	}
}

// TestSetOrInsertKeepsIDOutOfSetButKeepsItOnInsert pins the two halves of the
// _id handling together, because either alone is a bug.
//
// $set-ing _id makes MongoDB reject the entire write with an immutable-field
// error whenever the document's _id differs from the stored one — verified live.
// Dropping _id altogether would instead lose the caller's chosen identifier on a
// genuine insert, which is silent and worse.
func TestSetOrInsertKeepsIDOutOfSetButKeepsItOnInsert(t *testing.T) {
	builder, err := setOrInsertUpdate(probeDoc{ID: "chosen-id", Name: "n", Status: "s"})
	if err != nil {
		t.Fatalf("building update: %v", err)
	}
	doc := builder.Build()

	set, ok := doc["$set"].(bson.M)
	if !ok {
		t.Fatalf("update has no $set, so it would never modify an existing document: %v", operators(t, doc))
	}
	if _, present := set["_id"]; present {
		t.Error("$set carries _id; MongoDB rejects the whole write with ImmutableField when it differs from the stored value")
	}
	if set["name"] != "n" || set["status"] != "s" {
		t.Errorf("$set lost fields: %v", set)
	}

	onInsert, ok := doc["$setOnInsert"].(bson.M)
	if !ok {
		t.Fatal("update has no $setOnInsert, so a genuine insert would not use the caller's chosen _id")
	}
	if onInsert["_id"] != "chosen-id" {
		t.Errorf("$setOnInsert._id = %v, want the caller's chosen id", onInsert["_id"])
	}
}

// TestSetOrInsertOmitsAZeroID pins the common case that would otherwise fail
// hard: keying on some other field with a struct whose _id was never populated.
//
// Writing _id:"" on insert would store a document under an empty identifier;
// $set-ing it would fail every update. Omitting it entirely lets MongoDB take
// the _id from the filter, or generate one.
func TestSetOrInsertOmitsAZeroID(t *testing.T) {
	builder, err := setOrInsertUpdate(probeDoc{Name: "n", Status: "s"})
	if err != nil {
		t.Fatalf("building update: %v", err)
	}
	doc := builder.Build()

	if _, present := doc["$setOnInsert"]; present {
		t.Errorf("a zero _id was written on insert; the document would be stored under an empty identifier: %v", doc["$setOnInsert"])
	}
	set, ok := doc["$set"].(bson.M)
	if !ok {
		t.Fatalf("update has no $set: %v", operators(t, doc))
	}
	if _, present := set["_id"]; present {
		t.Error("$set carries a zero _id, which fails with ImmutableField against any existing document")
	}
}

// TestSetOrInsertFromMapMatchesTheStructPath pins that the map variant applies
// the same _id rule. A divergence here would be invisible: both compile, both
// look right, and only the map path would fail in production.
func TestSetOrInsertFromMapMatchesTheStructPath(t *testing.T) {
	withID, err := setOrInsertUpdateFromMap(map[string]any{"_id": "chosen-id", "name": "n"})
	if err != nil {
		t.Fatalf("building update: %v", err)
	}
	doc := withID.Build()
	if set := doc["$set"].(bson.M); len(set) != 1 || set["name"] != "n" {
		t.Errorf("$set = %v, want the non-_id fields only", set)
	}
	if onInsert, ok := doc["$setOnInsert"].(bson.M); !ok || onInsert["_id"] != "chosen-id" {
		t.Errorf("$setOnInsert = %v, want the caller's _id", doc["$setOnInsert"])
	}

	zeroID, err := setOrInsertUpdateFromMap(map[string]any{"_id": "", "name": "n"})
	if err != nil {
		t.Fatalf("building update: %v", err)
	}
	if _, present := zeroID.Build()["$setOnInsert"]; present {
		t.Error("the map path wrote a zero _id on insert where the struct path omits it")
	}
}

// TestSetOrInsertFromMapDoesNotMutateTheCaller'sMap is spelled without an
// apostrophe below; the point is that deleting _id must not be visible to the
// caller, who may reuse the map for a second write or for logging.
func TestSetOrInsertFromMapDoesNotMutateCallerMap(t *testing.T) {
	fields := map[string]any{"_id": "x", "name": "n"}
	if _, err := setOrInsertUpdateFromMap(fields); err != nil {
		t.Fatalf("building update: %v", err)
	}
	if _, present := fields["_id"]; !present {
		t.Error("the caller's map had _id deleted out from under it")
	}
}

// TestEmptyDocumentIsRefusedRatherThanSentAsEmptySet pins a failure that would
// otherwise surface as an opaque server error. MongoDB rejects an empty $set,
// and the resulting message names neither the collection nor the caller.
func TestEmptyDocumentIsRefusedRatherThanSentAsEmptySet(t *testing.T) {
	if _, err := setOrInsertUpdate(struct{}{}); err == nil {
		t.Error("an empty document produced an update; MongoDB would reject it with an unhelpful error")
	}
	if _, err := setOrInsertUpdateFromMap(map[string]any{}); err == nil {
		t.Error("an empty map produced an update")
	}
	// A document carrying only an _id is legitimate: insert-under-this-id.
	if _, err := setOrInsertUpdateFromMap(map[string]any{"_id": "only"}); err != nil {
		t.Errorf("a document with only an _id was refused: %v", err)
	}
}

func TestIsZeroID(t *testing.T) {
	var nilObjectID bson.ObjectID

	tests := []struct {
		name string
		id   any
		want bool
	}{
		{"absent", nil, true},
		{"empty string", "", true},
		{"a real string", "abc", false},
		{"zero ObjectID", nilObjectID, true},
		{"a real ObjectID", bson.NewObjectID(), false},
		{"zero int", 0, true},
		{"a real int", 7, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isZeroID(tc.id); got != tc.want {
				t.Errorf("isZeroID(%#v) = %v, want %v", tc.id, got, tc.want)
			}
		})
	}
}

// TestUpdateResultHelpersNameTheSilentCase pins the reporting that would have
// made the original incident visible. MatchedWithoutModifying is exactly what
// the frozen ITS state machines returned on every single write.
func TestUpdateResultHelpersNameTheSilentCase(t *testing.T) {
	tests := []struct {
		name                    string
		result                  *UpdateResult
		insert, update, matched bool
	}{
		{
			name:   "insert-only write landing on an existing document — the incident",
			result: &UpdateResult{MatchedCount: 1, ModifiedCount: 0, UpsertedCount: 0},
			// The write reported no error and changed nothing.
			matched: true,
		},
		{
			name:   "a genuine insert",
			result: &UpdateResult{MatchedCount: 0, ModifiedCount: 0, UpsertedCount: 1},
			insert: true,
		},
		{
			name:   "a genuine update",
			result: &UpdateResult{MatchedCount: 1, ModifiedCount: 1},
			update: true,
		},
		{
			name:   "nothing matched at all",
			result: &UpdateResult{},
		},
		{
			name:   "a nil result must not panic",
			result: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.result.DidInsert(); got != tc.insert {
				t.Errorf("DidInsert() = %v, want %v", got, tc.insert)
			}
			if got := tc.result.DidUpdate(); got != tc.update {
				t.Errorf("DidUpdate() = %v, want %v", got, tc.update)
			}
			if got := tc.result.MatchedWithoutModifying(); got != tc.matched {
				t.Errorf("MatchedWithoutModifying() = %v, want %v", got, tc.matched)
			}
		})
	}
}

// TestDeprecatedUpsertMethodsDelegateRatherThanDuplicate pins that the retained
// methods keep their behaviour by CONSTRUCTION rather than by a copy that can
// drift.
//
// UpsertByField must go on doing precisely what it always did — one caller in
// this estate depends on insert-only semantics on purpose — and the cheapest way
// to guarantee that is for it to be the new method under the old name.
func TestDeprecatedUpsertMethodsDelegateRatherThanDuplicate(t *testing.T) {
	src, err := os.ReadFile("collection.go")
	if err != nil {
		t.Fatalf("reading source: %v", err)
	}
	text := string(src)

	for _, tc := range []struct{ method, delegatesTo string }{
		{"func (col *Collection) UpsertByField(", "col.InsertIfAbsentByField("},
		{"func (col *Collection) UpsertByFieldMap(", "col.InsertIfAbsentByFieldMap("},
	} {
		start := strings.Index(text, tc.method)
		if start < 0 {
			t.Errorf("%s was removed; it is retained for compatibility within v2 and must not be deleted", tc.method)
			continue
		}
		end := strings.Index(text[start:], "\n}\n")
		if end < 0 {
			t.Fatalf("could not find the end of %s", tc.method)
		}
		body := text[start : start+end]

		if !strings.Contains(body, tc.delegatesTo) {
			t.Errorf("%s no longer delegates to %s — its behaviour can now drift from the method it is documented as being identical to", tc.method, tc.delegatesTo)
		}
		if strings.Contains(body, "SetStruct(") {
			t.Errorf("%s now builds a $set update. It is documented as insert-only and a caller depends on that; changing it would silently start overwriting their data", tc.method)
		}
	}
}

// TestEveryDeprecatedIdentifierSaysWhyInItsFirstSentence pins the part that
// actually reaches a developer.
//
// staticcheck reports the Deprecated line at the call site, so the first
// sentence is the whole warning. "Use X instead" would tell someone what to
// migrate to but not that their current code is silently broken, which is the
// only reason this deprecation exists.
func TestEveryDeprecatedIdentifierSaysWhyInItsFirstSentence(t *testing.T) {
	src, err := os.ReadFile("collection.go")
	if err != nil {
		t.Fatalf("reading source: %v", err)
	}

	required := map[string]string{
		"func (col *Collection) UpsertByField(":            "NEVER MODIFIES AN EXISTING DOCUMENT",
		"func (col *Collection) UpsertByFieldMap(":         "never modifies an existing document",
		"func (col *Collection) UpsertByFieldWithOptions(": "never modifies",
		"type UpsertOptions struct":                        "OPPOSITE",
	}

	text := string(src)
	for decl, phrase := range required {
		at := strings.Index(text, decl)
		if at < 0 {
			t.Errorf("%s is missing", decl)
			continue
		}
		// The doc comment is the text immediately preceding the declaration.
		docStart := strings.LastIndex(text[:at], "\n\n")
		doc := text[docStart:at]

		if !strings.Contains(doc, "Deprecated:") {
			t.Errorf("%s carries no Deprecated: marker, so no linter will flag its call sites", decl)
		}
		if !strings.Contains(doc, phrase) {
			t.Errorf("%s is deprecated without stating the trap (%q). A developer reading only the linter warning would migrate for tidiness, not because their writes are being discarded", decl, phrase)
		}
	}
}

// TestNoNewExportedIdentifierIsCalledUpsert pins the naming rule that came out
// of this incident: the word is burned in this package.
//
// It named three different operations, and the one it actually performed was the
// one nobody expected. Anything new must say which of insert / merge / replace
// it does.
func TestNoNewExportedIdentifierIsCalledUpsert(t *testing.T) {
	src, err := os.ReadFile("collection.go")
	if err != nil {
		t.Fatalf("reading source: %v", err)
	}

	// The three retained-for-compatibility methods, plus the options type.
	permitted := map[string]bool{
		"UpsertByField":            true,
		"UpsertByFieldMap":         true,
		"UpsertByFieldWithOptions": true,
		"UpsertOptions":            true,
	}

	for _, line := range strings.Split(string(src), "\n") {
		if !strings.HasPrefix(line, "func (col *Collection) ") && !strings.HasPrefix(line, "type ") {
			continue
		}
		name := exportedNameOf(line)
		if name == "" || !strings.Contains(name, "Upsert") {
			continue
		}
		if !permitted[name] {
			t.Errorf("new exported identifier %q contains \"Upsert\". Name it for what it does to an existing document: InsertIfAbsent, SetOrInsert or ReplaceOrInsert", name)
		}
	}
}

func exportedNameOf(line string) string {
	line = strings.TrimPrefix(line, "func (col *Collection) ")
	line = strings.TrimPrefix(line, "type ")
	cut := strings.IndexAny(line, "( ")
	if cut < 0 {
		return ""
	}
	return line[:cut]
}
