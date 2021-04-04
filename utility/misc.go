package utility

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// The seach are AND, meanining must contain all
func Grep(b []byte, search ...string) []string {
	var lines []string
	text_lines := strings.Split(string(b), "\n")
check_line:
	for _, text_line := range text_lines {
		for _, word := range search {
			if match, _ := regexp.MatchString(word, text_line); match == false {
				continue check_line
			}
		}
		lines = append(lines, text_line)
	}
	return lines
}

// The search are OR, meaning contain any of them
func Egrep(b []byte, search ...string) []string {
	var lines []string
	text_lines := strings.Split(string(b), "\n")

	for _, text_line := range text_lines {
		for _, word := range search {
			if match, _ := regexp.MatchString(word, text_line); match {
				lines = append(lines, text_line)
				break
			}
		}
	}
	return lines
}

func Cat(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

// ReadFile returns a file content, logs eny error
func CatS(paths ...string) string {
	path := filepath.Join(paths...)
	b, _ := ioutil.ReadFile(path)
	return strings.TrimSpace(string(b))
}

func GrepFile(rgx string, fp string) []string {
	b, err := Cat(fp)
	if err != nil {
		return []string{}
	}
	return Grep(b, rgx)
}

func GrepFiles(rgx string, path string) (out map[string][]string) {
	out = map[string][]string{}
	if results, err := CatAll(path); err == nil {
		for fname, content := range results {
			b := []byte(content)
			lines := Grep(b, rgx)
			if len(lines) > 0 {
				out[fname] = lines
			}
		}
	}
	return
}

func CatAll(path string) (results map[string]string, err error) {
	files := []os.FileInfo{}
	results = map[string]string{}
	n := 0

	pattern := filepath.Base(path)
	dir := filepath.Dir(path)

	fi, ef := os.Stat(path)
	switch {
	case ef == nil && !fi.IsDir():
		files = append(files, fi)
		path = ""
	case ef == nil && fi.IsDir():
		dir = path
		pattern = ""
		if files, err = ioutil.ReadDir(dir); err != nil {
			return
		}
	default:
		_, ef = os.Stat(dir)
		if os.IsNotExist(ef) {
			err = ef
			return
		}
		if files, err = ioutil.ReadDir(dir); err != nil {
			return
		}
	}

	for _, file := range files {
		var (
			out []byte
			e   error
		)
		if file.IsDir() {
			continue
		}

		fname := file.Name()
		skip := false
		if pattern != "" {
			// regex is much slower than strings match, but more flexible
			match, _ := regexp.MatchString(pattern, fname)
			skip = !match
		}
		if skip {
			continue
		}

		fp := filepath.Join(dir, fname)

		if file.Size() > 3000000 {
			continue
		}

		out, e = ioutil.ReadFile(fp)
		n++

		isBinary := false
		for _, b := range out {
			if rune(b) == 0 {
				isBinary = true
				break
			}
		}
		if isBinary {
			continue
		}

		if e == nil {
			results[fname] = strings.TrimSpace(string(out))
		}
	}
	return
}
