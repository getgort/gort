package main

import (
	"fmt"
	"regexp"
)

func main() {
	const text = "curl <http://google.com|google.com> is not <http://foo.com|foo.com> anymore."

	regexp := regexp.MustCompile("<([a-zA-Z0-9]*://[a-zA-Z0-9]*\\.[a-zA-Z0-9]*)\\|([a-zA-Z0-9]*\\.[a-zA-Z0-9]*)>")

	indices := regexp.FindAllSubmatchIndex([]byte(text), -1)
	last := 0

	if len(indices) > 0 {
		str := ""

		for _, z := range indices {
			str += text[last:z[0]]
			str += text[z[4]:z[5]]
			last = z[5] + 1
		}

		str += text[indices[len(indices)-1][5]+1:]

		fmt.Println(str)
	}
}
