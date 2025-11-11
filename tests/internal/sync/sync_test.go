package test

import (
	"testing"

	"github.com/db-sync-tools/internal/metadata"
	"github.com/db-sync-tools/internal/sync"
)

func TestDecideObjectType(t *testing.T) {
	if got := sync.DecideObjectType(&metadata.TableMetadata{IsView: true}); got != "view" {
		t.Fatalf("want view, got %s", got)
	}
	if got := sync.DecideObjectType(&metadata.TableMetadata{IsView: false}); got != "table" {
		t.Fatalf("want table, got %s", got)
	}
	if got := sync.DecideObjectType(nil); got != "table" {
		t.Fatalf("nil meta should default to table, got %s", got)
	}
}
