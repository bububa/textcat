package textcat

import (
	"bytes"
	"encoding/gob"
	"io/ioutil"
	"regexp"
	"sort"
	"strings"
)

const (
	MaxPatterns = 1000
)

var (
	reInvalid = regexp.MustCompile("[^\\p{L}]+")
)

type countType struct {
	S string
	I int
}

type countsType []*countType

func (c countsType) Len() int {
	return len(c)
}

func (c countsType) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c countsType) Less(i, j int) bool {
	if c[i].I != c[j].I {
		return c[i].I > c[j].I
	}
	return c[i].S < c[j].S
}

func GetPatterns(s string, useRunes bool) []*countType {
	ngrams := make(map[string]int)
	if useRunes {
		s = strings.ToLower(reInvalid.ReplaceAllString(s, " "))
		for _, word := range strings.Fields(s) {
			b := []rune("_" + word + "____")
			n := len(b) - 4
			for i := 0; i < n; i++ {
				for j := 1; j < 6; j++ {
					s = string(b[i : i+j])
					if !strings.HasSuffix(s, "__") {
						ngrams[s] += 1
					}
				}
			}
		}
	} else {
		for _, word := range strings.Fields(s) {
			b := []byte("_" + word + "____")
			n := len(b) - 4
			for i := 0; i < n; i++ {
				for j := 1; j < 6; j++ {
					s = string(b[i : i+j])
					if !strings.HasSuffix(s, "__") {
						ngrams[s] += 1
					}
				}
			}
		}
	}
	size := len(ngrams)
	counts := make([]*countType, 0, size)
	for i := range ngrams {
		counts = append(counts, &countType{i, ngrams[i]})
	}
	sort.Sort(countsType(counts))
	if size > MaxPatterns {
		counts = counts[:MaxPatterns]
	}
	return counts
}

func (tc *TextCat) EncodePatterns() (bytes.Buffer, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(tc.extra)
	if err != nil {
		return buf, err
	}
	return buf, nil
}

func (tc *TextCat) DecodePatterns(buf *bytes.Buffer) error {
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&tc.extra)
	if err != nil {
		return err
	}
	return nil
}

func (tc *TextCat) SavePatterns(filename string) error {
	buf, err := tc.EncodePatterns()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, buf.Bytes(), 755)
}

func (tc *TextCat) LoadPatterns(filename string) error {
	byt, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(byt)
	return tc.DecodePatterns(buf)
}

func (tc *TextCat) Extra() map[int]map[string]int {
	return tc.extra
}
