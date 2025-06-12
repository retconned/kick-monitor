package util

import (
	"regexp"
	"sort" // for sorting time slices
	"strings"
	"time"
)

// NormalizeChatMessage cleans up message content for comparison
func NormalizeChatMessage(message string) string {
	// Convert to lowercase
	normalized := strings.ToLower(message)
	// Remove leading/trailing whitespace
	normalized = strings.TrimSpace(normalized)
	// You might want to remove punctuation or extra spaces here too
	normalized = regexp.MustCompile(`\s+`).ReplaceAllString(normalized, " ") // Replace multiple spaces with single
	return normalized
}

// JaccardSimilarity calculates the Jaccard similarity between two strings.
// It's a simple token-based similarity. Can be used for "similar message" detection.
func JaccardSimilarity(s1, s2 string) float64 {
	words1 := strings.Fields(s1)
	words2 := strings.Fields(s2)

	if len(words1) == 0 && len(words2) == 0 {
		return 1.0 // Both empty, considered identical
	}
	if len(words1) == 0 || len(words2) == 0 {
		return 0.0 // One empty, one not
	}

	set1 := make(map[string]struct{})
	for _, word := range words1 {
		set1[word] = struct{}{}
	}

	set2 := make(map[string]struct{})
	for _, word := range words2 {
		set2[word] = struct{}{}
	}

	intersection := 0
	for word := range set1 {
		if _, found := set2[word]; found {
			intersection++
		}
	}

	union := len(set1) + len(set2) - intersection

	if union == 0 { // Should only happen if both were empty, handled above
		return 0.0
	}

	return float64(intersection) / float64(union)
}

func ContainsString(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// UniqueStrings returns a slice with unique strings, preserving order (roughly).
func UniqueStrings(slice []string) []string {
	seen := make(map[string]struct{})
	result := []string{}
	for _, val := range slice {
		if _, ok := seen[val]; !ok {
			seen[val] = struct{}{}
			result = append(result, val)
		}
	}
	return result
}

// UniqueSortedTimes returns a slice with unique time.Time values, sorted.
func UniqueSortedTimes(slice []time.Time) []time.Time {
	// Convert to string set to find unique values
	timeMap := make(map[string]time.Time)
	for _, t := range slice {
		timeMap[t.Format(time.RFC3339Nano)] = t // Use a precise string representation as map key
	}

	uniqueTimes := make([]time.Time, 0, len(timeMap))
	for _, t := range timeMap {
		uniqueTimes = append(uniqueTimes, t)
	}

	// Sort the unique times
	sort.Slice(uniqueTimes, func(i, j int) bool {
		return uniqueTimes[i].Before(uniqueTimes[j])
	})

	return uniqueTimes
}
