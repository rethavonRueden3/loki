package main

import (
	"testing"
)

func TestMergeResponsesDeduplicateBeforeLimit(t *testing.T) {
	// Mock responses with overlapping boundary timestamps
	// Window 1 ends with timestamp 10
	// Window 2 starts with timestamp 10
	resp1 := QueryResponse{
		Streams: []Stream{
			{
				Labels: "{app=\"test\"}",
				Entries: []Entry{
					{Timestamp: 8, Line: "log 8"},
					{Timestamp: 9, Line: "log 9"},
					{Timestamp: 10, Line: "log 10"},
				},
			},
		},
	}
	resp2 := QueryResponse{
		Streams: []Stream{
			{
				Labels: "{app=\"test\"}",
				Entries: []Entry{
					{Timestamp: 10, Line: "log 10"},
					{Timestamp: 11, Line: "log 11"},
					{Timestamp: 12, Line: "log 12"},
				},
			},
		},
	}

	// Merge with limit 5
	merged := MergeResponses([]QueryResponse{resp1, resp2}, 5, true)

	// Count total entries
	totalEntries := 0
	for _, stream := range merged.Streams {
		totalEntries += len(stream.Entries)
	}

	if totalEntries != 5 {
		t.Errorf("Expected exactly 5 entries, got %d", totalEntries)
	}

	// Verify no duplicates
	seen := make(map[int64]bool)
	for _, stream := range merged.Streams {
		for _, entry := range stream.Entries {
			if seen[entry.Timestamp] {
				t.Errorf("Found duplicate entry at timestamp %d", entry.Timestamp)
			}
			seen[entry.Timestamp] = true
		}
	}
}
