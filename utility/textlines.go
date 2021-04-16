package utility

import (
	"encoding/json"
	"fmt"
	"strings"
)

type TextLines []string

func (ls *TextLines) Append(lines ...string) {
	*ls = append(*ls, lines...)
}

func (ls TextLines) String() (s string) {
	for _, line := range ls {
		s += fmt.Sprintf("%v\n", line)
	}
	return
}

func (ls TextLines) Len() int {
	return len(ls)
}

func (ls *TextLines) Empty() {
	*ls = (*ls)[:0]
}

func (ls *TextLines) Prepend(s string) {
	*ls = append(*ls, "")
	copy((*ls)[1:], *ls)
	(*ls)[0] = s
}

func (ls *TextLines) Insert(j int, s string) {
	*ls = append(*ls, "")
	copy((*ls)[j+1:], (*ls)[j:])
	(*ls)[j] = s
}

func (ls TextLines) ToJson() string {
	j, _ := json.Marshal(ls)
	return string(j)
}

func (ls *TextLines) FromJson(j string) error {
	return json.Unmarshal([]byte(j), ls)
}

type TextLine string

func (l TextLine) FirstAfterPrefix(w string, target *string) (ok bool) {
	line := strings.TrimSpace(string(l))
	if strings.HasPrefix(line, w) {
		if s := strings.Split(line, w); len(s) > 1 {
			s2 := strings.Split(strings.TrimSpace(s[1]), " ")
			if target != nil {
				*target = strings.TrimSpace(s2[0])
			}
			ok = true
		}
	}
	return
}

func (l TextLine) NAfterPrefix(w string, targets ...*string) (ok bool) {
	line := strings.TrimSpace(string(l))
	if strings.HasPrefix(line, w) {
		if s := strings.Split(line, w); len(s) > 1 {
			s2 := strings.Split(strings.TrimSpace(s[1]), " ")
			for i := range targets {
				*targets[i] = strings.TrimSpace(s2[i])
			}
			ok = true
		}
	}
	return
}

func (l TextLine) Contains(w string) bool {
	return strings.Contains(string(l), w)
}

func (l TextLine) ContainsAll(ws ...string) bool {
	for _, w := range ws {
		if !l.Contains(w) {
			return false
		}
	}
	return true
}

func (l TextLine) ContainsAny(ws ...string) bool {
	for _, w := range ws {
		if l.Contains(w) {
			return true
		}
	}
	return false
}

func (l TextLine) FirstAfter(w string, target *string) (ok bool) {
	line := strings.TrimSpace(string(l))
	if strings.Contains(line, w) {
		if s := strings.Split(line, w); len(s) > 1 {
			s2 := strings.Split(strings.TrimSpace(s[1]), " ")
			if target != nil {
				*target = strings.TrimSpace(s2[0])
			}
			ok = true
		}
	}
	return
}

func (l TextLine) HasPrefix(w string) (ok bool) {
	line := strings.TrimSpace(string(l))
	return strings.HasPrefix(line, w)
}

func (l TextLine) Split() (out []string) {
	s := strings.Split(string(l), " ")
	for j := range s {
		s[j] = strings.TrimSpace(s[j])
		if s[j] != "" {
			out = append(out, s[j])
		}
	}
	return
}
