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
	var user = &User{}

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	rd := bufio.NewScanner(file)

	fmt.Fprintln(out, "found users:")

	for i := 0; rd.Scan(); i++ {
		err = user.UnmarshalJSON(rd.Bytes())
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

		buf.WriteString("[" + strconv.Itoa(i) + "] " + user.Name + " <" + email + ">\n")
	}

	buf.WriteString("\nTotal unique browsers " + strconv.Itoa(len(uniqueBrowsers)) + "\n")
	fmt.Fprint(out, buf.String())
}

func main() {
	slowOut := new(bytes.Buffer)
	FastSearch(slowOut)
	fmt.Println(slowOut)
}
