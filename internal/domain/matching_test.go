package domain

import (
	"testing"
)

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		s, t     string
		expected int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "b", 1},
		{"abc", "abc", 0},
		{"abc", "ab", 1},
		{"abc", "ac", 1},
		{"sitting", "kitten", 3},
		{"cake", "dake", 1},
		{"Saturday", "Sunday", 3},
	}

	for _, tt := range tests {
		got := LevenshteinDistance(tt.s, tt.t)
		if got != tt.expected {
			t.Errorf("LevenshteinDistance(%q, %q) = %d; want %d", tt.s, tt.t, got, tt.expected)
		}
	}
}

func TestSimilarity(t *testing.T) {
	tests := []struct {
		s, t      string
		threshold float64
		expected  bool
	}{
		{"", "", 1.0, true},
		{"hello", "hello", 1.0, true},
		{"hello", "hallo", 0.8, true},
		{"hello", "world", 0.5, false},
	}

	for _, tt := range tests {
		got := Similarity(tt.s, tt.t)
		isOk := got >= tt.threshold
		if isOk != tt.expected {
			t.Errorf("Similarity(%q, %q) = %f; threshold %f matched=%t; want matched=%t", tt.s, tt.t, got, tt.threshold, isOk, tt.expected)
		}
	}
}

func TestCleanString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hey Jude (Remastered 2015)", "hey jude 2015"},
		{"Bohemian Rhapsody - Official Video", "bohemian rhapsody"},
		{"Let It Be [Acoustic Version]", "let it be version"}, // "acoustic" is noise, "version" is kept (unless in noise list)
		{"Some Song... (Live)!!", "some song"},
		{"  Spaces   and   Punctuation! ", "spaces and punctuation"},
	}

	for _, tt := range tests {
		got := CleanString(tt.input)
		if got != tt.expected {
			t.Errorf("CleanString(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}

func TestIsMatch(t *testing.T) {
	tests := []struct {
		targetTitle, targetArtist       string
		candidateTitle, candidateArtist string
		expected                        bool
	}{
		{"Hey Jude", "The Beatles", "Hey Jude - Remastered", "Beatles", true},
		{"Bohemian Rhapsody", "Queen", "Bohemian Rhapsody (Official Video)", "Queen", true},
		{"Bohemian Rhapsody", "Queen", "Bohemian Rhapsody", "Princess", false},
		{"Let It Be", "The Beatles", "Hey Jude", "The Beatles", false},
		{"Track Name", "Artist A", "track name", "artist a feat. artist b", true},
	}

	for _, tt := range tests {
		got := IsMatch(tt.targetTitle, tt.targetArtist, tt.candidateTitle, tt.candidateArtist)
		if got != tt.expected {
			t.Errorf("IsMatch(%q, %q, %q, %q) = %t; want %t", tt.targetTitle, tt.targetArtist, tt.candidateTitle, tt.candidateArtist, got, tt.expected)
		}
	}
}
