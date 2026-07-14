package main

import (
	"fmt"
	"sort"
)

// Entry represents a log entry.
type Entry struct {
	Timestamp int64
	Line      string
}

// Stream represents a stream of log entries with labels.
type Stream struct {
	Labels  string
	Entries []Entry
}

// QueryResponse represents the response of a query.
type QueryResponse struct {
	Streams []Stream
}

// MergeResponses merges multiple QueryResponses, deduplicates entries with identical timestamps and labels,
// and applies the limit constraint after deduplication.
func MergeResponses(responses []QueryResponse, limit int, forward bool) QueryResponse { 
	// Group entries by labels
	labelStreams := make(map[string][]Entry)
	for _, resp := range responses {
		for _, stream := range resp.Streams {
			labelStreams[stream.Labels] = append(labelStreams[stream.Labels], stream.Entries...)
		}
	}

	mergedStreams := make([]Stream, 0, len(labelStreams))
	for labels, entries := range labelStreams {
		// Sort entries by timestamp
		sort.Slice(entries, func(i, j int) bool {
			if forward {
				return entries[i].Timestamp < entries[j].Timestamp
			} 
			return entries[i].Timestamp > entries[j].Timestamp
		})

		// Deduplicate entries with identical timestamps
		deduplicated := make([]Entry, 0, len(entries))
		for _, entry := range entries {
			if len(deduplicated) > 0 && deduplicated[len(deduplicated)-1].Timestamp == entry.Timestamp && deduplicated[len(deduplicated)-1].Line == entry.Line {
				continue
			}
			deduplicated = append(deduplicated, entry)
		}

		mergedStreams = append(mergedStreams, Stream{
			Labels:  labels,
			Entries: deduplicated,
		})
	}

	// Now merge all streams and apply the global limit
	// To do this correctly, we sort all entries across all streams if we want a global limit, 
	// or we limit the total count of returned entries across all streams.
	// Let's collect all unique entries, sort them globally (or per stream depending on Loki's behavior, but usually limit is global).
	type streamEntry struct {
		entry  Entry
		labels string
	}
	var allEntries []streamEntry
	for _, stream := range mergedStreams {
		for _, entry := range stream.Entries {
			allEntries = append(allEntries, streamEntry{entry: entry, labels: stream.Labels})
		}
	}

	sort.Slice(allEntries, func(i, j int) bool {
		if forward {
			return allEntries[i].entry.Timestamp < allEntries[j].entry.Timestamp
		}
		return allEntries[i].entry.Timestamp > allEntries[j].entry.Timestamp
	})

	// Apply limit
	if limit > 0 && len(allEntries) > limit {
		allEntries = allEntries[:limit]
	}

	// Reconstruct streams
	resultStreamsMap := make(map[string][]Entry)
	for _, se := range allEntries {
		resultStreamsMap[se.labels] = append(resultStreamsMap[se.labels], se.entry)
	}

	var result QueryResponse
	for labels, entries := range resultStreamsMap {
		result.Streams = append(result.Streams, Stream{
			Labels:  labels,
			Entries: entries,
		})
	}

	return result
}

func main() {
	fmt.Println("Hello, Bounty Hunter!")
}
