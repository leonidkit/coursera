package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func FastSearch(out io.Writer) {
	uniqueBrowsers := map[string]bool{}
	var buf bytes.Buffer

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	rd := bufio.NewReader(file)

	fmt.Fprintln(out, "found users:")

	for i := 0; ; i++ {
		line, err := rd.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("error reading file %s", err)
		}

		var user = &User{}
		err = user.UnmarshalJSON(line)
		if err != nil {
			panic(err)
		}

		isAndroid := false
		isMSIE := false

		for _, browserRaw := range user.Browsers {
			switch {
			case strings.Contains(browserRaw, "Android"):
				isAndroid = true
			case strings.Contains(browserRaw, "MSIE"):
				isMSIE = true
			default:
				continue
			}

			uniqueBrowsers[browserRaw] = true
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		email := strings.Replace(user.Email, "@", " [at] ", -1)

		buf.WriteString("[" + strconv.Itoa(i) + "] ")
		buf.WriteString(user.Name)
		buf.WriteString(" <" + email + ">\n")
	}

	fmt.Fprint(out, buf.String())
	fmt.Fprintln(out, "\nTotal unique browsers", len(uniqueBrowsers))
}

func main() {
	slowOut := new(bytes.Buffer)
	FastSearch(slowOut)
	fmt.Println(slowOut)
}
