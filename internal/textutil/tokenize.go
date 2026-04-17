package textutil

import (
	"regexp"
	"sort"
	"strings"
)

var tokenPattern = regexp.MustCompile(`[A-Za-z0-9]+`)

var stopwords = map[string]struct{}{
	"the": {}, "and": {}, "for": {}, "with": {}, "from": {}, "your": {}, "this": {}, "that": {}, "you": {}, "are": {},
	"app": {}, "ios": {}, "music": {}, "free": {}, "get": {}, "use": {}, "can": {}, "will": {}, "into": {}, "more": {},
}

func Keywords(texts ...string) map[string]int {
	out := map[string]int{}
	for _, text := range texts {
		for _, token := range tokenPattern.FindAllString(strings.ToLower(text), -1) {
			if len(token) < 3 {
				continue
			}
			if _, ok := stopwords[token]; ok {
				continue
			}
			out[token]++
		}
	}
	return out
}

func SortedDiff(target, competitors map[string]int, limit int) []string {
	type pair struct {
		word  string
		score int
	}
	var pairs []pair
	for word, score := range competitors {
		if target[word] == 0 {
			pairs = append(pairs, pair{word: word, score: score})
		}
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].score == pairs[j].score {
			return pairs[i].word < pairs[j].word
		}
		return pairs[i].score > pairs[j].score
	})
	if limit > 0 && len(pairs) > limit {
		pairs = pairs[:limit]
	}
	out := make([]string, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, p.word)
	}
	return out
}
