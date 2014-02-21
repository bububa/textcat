package textcat

import (
	"errors"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	defaultThresholdValue = 1.03
	defaultMaxCandidates  = 5
	defaultMinDocSize     = 25
)

var (
	errShort   = errors.New("SHORT")
	errUnknown = errors.New("UNKNOWN")
	errAvail   = errors.New("NOPATTERNS")
)

type TextCat struct {
	category       map[int]bool
	thresholdValue float64
	maxCandidates  int
	minDocSize     int
	extra          map[int]map[string]int
}

type resultType struct {
	score    int
	category int
}

type resultsType []*resultType

func (r resultsType) Len() int {
	return len(r)
}

func (r resultsType) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r resultsType) Less(i, j int) bool {
	if r[i].score != r[j].score {
		return r[i].score < r[j].score
	}
	return r[i].category < r[j].category
}

func NewTextCat() *TextCat {
	tc := &TextCat{
		categories:     make(map[int]bool),
		thresholdValue: defaultThresholdValue,
		maxCandidates:  defaultMaxCandidates,
		minDocSize:     defaultMinDocSize}
	tc.extra = make(map[int]map[string]int)
	return tc
}

func (tc *TextCat) SetThresholdValue(thresholdValue float64) {
	tc.thresholdValue = thresholdValue
}

func (tc *TextCat) GetThresholdValue() float64 {
	return tc.thresholdValue
}

func (tc *TextCat) SetMaxCandidates(maxCandidates int) {
	tc.maxCandidates = maxCandidates
}

func (tc *TextCat) GetMaxCandidates() int {
	return tc.maxCandidates
}

func (tc *TextCat) SetMinDocSize(minDocSize int) {
	tc.minDocSize = minDocSize
}

func (tc *TextCat) GetMinDocSize() int {
	return tc.minDocSize
}

func (tc *TextCat) ActiveCategories() []int {
	a := make([]int, 0, len(tc.categories))
	for category := range tc.categories {
		if tc.categories[category] {
			a = append(a, category)
		}
	}
	sort.Ints(a)
	return a
}

func (tc *TextCat) AvailableCategories() []int {
	a := make([]int, 0, len(tc.categories))
	for category := range tc.categories {
		a = append(a, category)
	}
	sort.Ints(a)
	return a
}

func (tc *TextCat) DisableCategories(categories ...int) {
	for _, category := range categories {
		if _, exists := tc.categories[category]; exists {
			tc.categories[category] = false
		}
	}
}

func (tc *TextCat) DisableAllCategories() {
	for category := range tc.categories {
		tc.categories[category] = false
	}
}

func (tc *TextCat) EnableCategories(categories ...int) {
	for _, category := range categories {
		if _, exists := tc.categories[category]; exists {
			tc.categories[category] = true
		}
	}
}

func (tc *TextCat) EnableAllCategories() {
	for category := range tc.categories {
		tc.categories[category] = true
	}
}

func (tc *TextCat) Classify(text string) (categories []int, err error) {
	var mydata map[int]int

	categories = make([]int, 0, tc.maxCandidates)

	if utf8.RuneCountInString(strings.TrimSpace(reInvalid.ReplaceAllString(text, " "))) < tc.minDocSize {
		err = errShort
		return
	}

	scores := make([]*resultType, 0, len(tc.categories))
	patt := GetPatterns(text, utf8)
	for category := range tc.categories {
		if !tc.categories[category] {
			continue
		}
		if _, ok := tc.extra[category]; ok {
			mydata = tc.extra[category]
		}
		score := 0
		for n, p := range patt {
			i, ok := mydata[p.S]
			if !ok {
				i = MaxPatterns
			}
			if n > i {
				score += n - i
			} else {
				score += i - n
			}
		}
		scores = append(scores, &resultType{score, category})
	}
	if len(scores) == 0 {
		err = errAvail
		return
	}

	minScore := MaxPatterns * MaxPatterns
	for _, sco := range scores {
		if sco.score < minScore {
			minScore = sco.score
		}
	}
	threshold := float64(minScore) * tc.thresholdValue
	nCandidates := 0
	for _, sco := range scores {
		if float64(sco.score) <= threshold {
			nCandidates += 1
		}
	}
	if nCandidates > tc.maxCandidates {
		err = errUnknown
		return
	}

	lowScores := make([]*resultType, 0, nCandidates)
	for _, sco := range scores {
		if float64(sco.score) <= threshold {
			lowScores = append(lowScores, sco)
		}
	}
	sort.Sort(resultsType(lowScores))
	for _, sco := range lowScores {
		categories = append(categories, sco.category)
	}

	return
}

func (tc *TextCat) AddCategoryWords(category, words []string) {
	if len(words) == 0 {
		return
	}
	if len(words) > 0 {
		a := make(map[string]int)
		for word := range words {
			for i, p := range textcat.GetPatterns(str, true) {
				if i == MaxPatterns {
					break
				}
				a[p.S] = i
			}
		}
		tc.extra[category] = a
		if _, ok := tc.categories[category]; !ok {
			tc.categories[category] = false
		}
	}

	return nil
}
