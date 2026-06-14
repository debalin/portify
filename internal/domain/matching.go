package domain

import (
	"strings"
)

// LevenshteinDistance calculates the Levenshtein distance between two strings.
// It is rune-aware and handles Unicode characters correctly.
func LevenshteinDistance(s, t string) int {
	sRunes := []rune(strings.ToLower(strings.TrimSpace(s)))
	tRunes := []rune(strings.ToLower(strings.TrimSpace(t)))

	lenS := len(sRunes)
	lenT := len(tRunes)

	if lenS == 0 {
		return lenT
	}
	if lenT == 0 {
		return lenS
	}

	d := make([][]int, lenS+1)
	for i := range d {
		d[i] = make([]int, lenT+1)
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}

	for i := 1; i <= lenS; i++ {
		for j := 1; j <= lenT; j++ {
			cost := 0
			if sRunes[i-1] != tRunes[j-1] {
				cost = 1
			}
			d[i][j] = minOfThree(
				d[i-1][j]+1,      // deletion
				d[i][j-1]+1,      // insertion
				d[i-1][j-1]+cost, // substitution
			)
		}
	}
	return d[lenS][lenT]
}

func minOfThree(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// Similarity calculates the similarity ratio between two strings (0.0 to 1.0).
func Similarity(s, t string) float64 {
	sRunes := []rune(strings.ToLower(strings.TrimSpace(s)))
	tRunes := []rune(strings.ToLower(strings.TrimSpace(t)))
	maxLen := len(sRunes)
	if len(tRunes) > maxLen {
		maxLen = len(tRunes)
	}
	if maxLen == 0 {
		return 1.0
	}
	dist := LevenshteinDistance(s, t)
	return 1.0 - float64(dist)/float64(maxLen)
}

// CleanString normalizes strings for matching by lowering, converting punctuation to spaces,
// and stripping common noise keywords.
func CleanString(s string) string {
	s = strings.ToLower(s)
	// Replace punctuation with spaces
	s = strings.Map(func(r rune) rune {
		if strings.ContainsRune("()[]{}.-_,,:;!?/\\\"'~*&^%$#@`+=", r) {
			return ' '
		}
		return r
	}, s)
	// Remove common extra keywords
	noise := []string{"official audio", "official video", "lyric video", "official music video", "lyrics", "remastered", "remaster", "live", "acoustic"}
	for _, n := range noise {
		s = strings.ReplaceAll(s, n, " ")
	}
	// Normalize spacing
	words := strings.Fields(s)
	return strings.Join(words, " ")
}

// IsMatch checks if candidate track metadata matches the target metadata.
func IsMatch(targetTitle, targetArtist, candidateTitle, candidateArtist string) bool {
	cleanTargetTitle := CleanString(targetTitle)
	cleanCandidateTitle := CleanString(candidateTitle)
	cleanTargetArtist := CleanString(targetArtist)
	cleanCandidateArtist := CleanString(candidateArtist)

	titleSim := Similarity(cleanTargetTitle, cleanCandidateTitle)
	artistSim := Similarity(cleanTargetArtist, cleanCandidateArtist)

	// If candidate title contains both target title and target artist, consider it a match
	if cleanTargetTitle != "" && cleanTargetArtist != "" &&
		strings.Contains(cleanCandidateTitle, cleanTargetTitle) &&
		strings.Contains(cleanCandidateTitle, cleanTargetArtist) {
		return true
	}

	// Accept title if similarity is >= 0.75
	titleMatches := titleSim >= 0.75

	// Accept artist if similarity is >= 0.75, or one contains the other
	artistMatches := artistSim >= 0.75 ||
		strings.Contains(cleanCandidateArtist, cleanTargetArtist) ||
		strings.Contains(cleanTargetArtist, cleanCandidateArtist)

	return titleMatches && artistMatches
}
