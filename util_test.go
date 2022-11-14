package main

import (
	"strings"
	"testing"
)

func TestGetUniqueWords(t *testing.T) {
	var tests = []struct {
		input  string
		output []string
	}{
		{input: "", output: []string{""}},
		{input: " ", output: []string{""}},
		{input: "one", output: []string{"one"}},
		{input: "one..", output: []string{"one"}},
		{input: "free'd", output: []string{"free'd"}},
		{input: "l'amour", output: []string{"l'amour"}},
		{input: "one two", output: []string{"one", "two"}},
		{input: "Mr Mrs", output: []string{"Mr", "Mrs"}},
		{input: "Mr. Mrs.", output: []string{"Mr", "Mrs"}},
		{input: "one!!!---", output: []string{"one"}},
		{input: "..  one..", output: []string{"one"}},
		{input: "forget-me-not", output: []string{"forget-me-not"}},
		{input: "!  one--! two-", output: []string{"one", "two"}},
		{input: "Abracadabra", output: []string{"Abracadabra"}},

		{input: "one@(*#... two..@* three#^##@^", output: []string{"one", "two", "three"}},
		{input: "one@(*#... two..@* three aga#^##@^", output: []string{"one", "two", "three", "aga"}},
	}

	for ntest, tt := range tests {
		words := GetUniqueWords([]string{tt.input})
		if len(words) > len(tt.output) {
			const msg = "%d: words len: %d, tt.output len: %d, input: %s, output: %s"
			t.Errorf(msg, ntest, len(words), len(tt.output), tt.input, tt.output)
			break
		}
		for i := 0; i < len(words); i++ {
			for _, out := range strings.Split(tt.output[i], " ") {
				if _, ok := words[out]; !ok {
					t.Errorf("(%s) in words[out] (%t)", out, ok)
				}
			}
		}
	}
}

func TestCountSpacesBetweenWords(t *testing.T) {
	var tests = []struct {
		input  string
		output int
	}{
		{input: "one", output: 0},
		{input: " one", output: 0},
		{input: "one two", output: 1},
		{input: "one two three", output: 2},
		{input: "one  two    three", output: 2},
		{input: "one two three four ", output: 3},
		{input: "one two three four   ", output: 3},
		{input: " one two three four ", output: 3},
		{input: " one two three four    ", output: 3},
	}

	for ntest, tt := range tests {
		count := CountSpacesBetweenWords(tt.input)

		if count != tt.output {
			const msg = "ntest: %d, got: %d, want %d\n"
			t.Errorf(msg, ntest, count, tt.output)
		}
	}
}

func TestOmitTrailingPunctuation(t *testing.T) {
	type test struct {
		in  string
		out string
	}

	tests := []test{
		{in: "foo", out: "foo"},
		{in: "foo.", out: "foo"},
		{in: "foo!", out: "foo"},
		{in: "foo?", out: "foo"},
		{in: "foo-", out: "foo"},

		{in: "foo--", out: "foo"},
		{in: "foo!!", out: "foo"},
		{in: "foo??", out: "foo"},

		{in: "----", out: "----"},
	}

	for ntest, tt := range tests {
		result := OmitTrailingPunctuation(tt.in)

		if len(result) != len(tt.out) && result != tt.out {
			const msg = "ntest: %d, got: %s, want %s\n"
			t.Errorf(msg, ntest, result, tt.out)
		}
	}
}

func TestOmitPrecedingPunctuation(t *testing.T) {
	type test struct {
		in  string
		out string
	}

	tests := []test{
		{in: "\"foo", out: "foo"},
		{in: "'foo", out: "foo"},
		{in: "'foo", out: "foo"},

		{in: "--foo", out: "foo"},
		{in: "!!foo", out: "foo"},
		{in: "??foo", out: "foo"},

		{in: "----", out: "----"},
	}

	for ntest, tt := range tests {
		result := OmitPrecedingPunctuation(tt.in)

		if len(result) != len(tt.out) && result != tt.out {
			const msg = "ntest: %d, got: %s, want %s\n"
			t.Errorf(msg, ntest, result, tt.out)
		}
	}
}

// hardcoded values! (10)
// potential bugs! when width is not the same accross chars
func TestGetSelectedCharLen(t *testing.T) {
	type test struct {
		in  int
		out int
	}

	tests := []test{
		{in: 0, out: 0},
		{in: 10, out: 1},
		{in: 20, out: 2},
	}

	for ntest, tt := range tests {
		result := GetSelectedCharLen(tt.in)
		if result != tt.out {
			const msg = "ntest: %d, got: %d, want %d\n"
			t.Errorf(msg, ntest, result, tt.out)
		}
	}
}
