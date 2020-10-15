package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

type File struct {
	Name   string
	Size   int64
	IsDir  bool
	Childs *[]*File
}

func (f *File) String() string {
	return f.Name
}

func walkDir(nodes *[]*File, path string, findFiles bool) {
	f, err := os.Open(path)
	if err != nil {
		panic(err.Error())
	}
	defer f.Close()

	files, err := f.Readdir(-1)

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	for _, file := range files {
		if file.IsDir() {
			newDir := &File{file.Name(), file.Size(), true, &[]*File{}}

			walkDir(newDir.Childs, filepath.Join(path, file.Name()), findFiles)

			*nodes = append(*nodes, newDir)
		} else if findFiles {
			newFile := &File{file.Name(), file.Size(), false, nil}

			*nodes = append(*nodes, newFile)
		}
	}
}

func printTree(out io.Writer, nodes *[]*File, prefix string) {
	for i, node := range *nodes {
		if i == len(*nodes)-1 {
			if !node.IsDir {
				size := ""
				if node.Size == 0 {
					size = "empty"
				} else {
					size = strconv.Itoa(int(node.Size)) + "b"
				}
				fmt.Fprintf(out, "%s└───%s (%s)\n", prefix, node.Name, size)
			} else {
				fmt.Fprintf(out, "%s└───%s\n", prefix, node.Name)
				printTree(out, node.Childs, prefix+"\t")
			}
			return
		} else {
			if !node.IsDir {
				size := ""
				if node.Size == 0 {
					size = "empty"
				} else {
					size = strconv.Itoa(int(node.Size)) + "b"
				}
				fmt.Fprintf(out, "%s├───%s (%s)\n", prefix, node.Name, size)
			} else {
				fmt.Fprintf(out, "%s├───%s\n", prefix, node.Name)
				printTree(out, node.Childs, prefix+"│\t")
			}
		}
	}
}

func dirTree(out io.Writer, path string, printFiles bool) error {
	nodes := []*File{}
	walkDir(&nodes, path, printFiles)
	printTree(out, &nodes, "")
	return nil
}

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}
