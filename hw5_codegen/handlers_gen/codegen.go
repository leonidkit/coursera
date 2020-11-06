package main

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strings"
	"text/template"
)

type Struct struct {
	Name   string
	Fields []StructField
}

type StructField struct {
	Type string
	Name string
	Tag  reflect.StructTag
}

type StructFunc struct {
	Name      string
	ParamType string
	RecvType  string
	Instr     Comment
}

type Comment struct {
	URL    string
	Auth   bool
	Method string
}

func getMeta(node *ast.File) (structs []*Struct, structFuncs []*StructFunc) {
	for _, decl := range node.Decls {
		switch decl.(type) {
		case (*ast.GenDecl):
			g := decl.(*ast.GenDecl)
			for _, spec := range g.Specs {
				currType, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				currStruct, ok := currType.Type.(*ast.StructType)
				if !ok {
					continue
				}

				var sf []StructField
				for _, field := range currStruct.Fields.List {
					var tag reflect.StructTag
					if field.Tag != nil {
						tag = reflect.StructTag(field.Tag.Value[1 : len(field.Tag.Value)-1])
						if tag.Get("apivalidator") == "-" || tag.Get("apivalidator") == "" {
							continue
						}
					}
					if fieldType, ok := field.Type.(*ast.Ident); ok {
						sf = append(sf, StructField{
							Name: field.Names[0].Name,
							Type: fieldType.Name,
							Tag:  tag,
						})
					}
				}
				structs = append(structs, &Struct{
					Name:   currType.Name.Name,
					Fields: sf,
				})
				_ = sf
			}
		case (*ast.FuncDecl):
			f := decl.(*ast.FuncDecl)
			if f.Doc != nil && f.Recv != nil {
				fComm := strings.TrimPrefix(f.Doc.Text(), "apigen:api ")
				var fInstr Comment
				json.Unmarshal([]byte(fComm), &fInstr)

				fRecvType := f.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
				fName := f.Name.Name

				var fParam string
				for _, field := range f.Type.Params.List {
					ident, ok := field.Type.(*ast.Ident)
					if ok {
						fParam = ident.Name
						break
					}
				}

				structFuncs = append(structFuncs, &StructFunc{
					Name:      fName,
					ParamType: fParam,
					RecvType:  fRecvType,
					Instr:     fInstr,
				})
			}
		}
	}
	return
}

func main() {
	resFile, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, os.Args[1], nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	structs, structFuncs := getMeta(node)

	tmpl := template.Must(template.ParseFiles("templates/header.go.tpl"))
	tmpl.Execute(resFile, struct{}{})

	fByRecv := map[string][]StructFunc{}
	for _, f := range structFuncs {
		fByRecv[f.RecvType] = append(fByRecv[f.RecvType], *f)
	}

	tmpl = template.Must(template.ParseFiles("templates/ServeHTTP.go.tpl"))

	for RecvType, Funcs := range fByRecv {
		tmpl.Execute(resFile, struct {
			RecvType string
			Funcs    []StructFunc
		}{RecvType, Funcs})
	}

	funcMap := template.FuncMap{
		"ToLower": strings.ToLower,
		"GetTag": func(name string, tag reflect.StructTag, subtag string) string {
			for _, val := range strings.Split(tag.Get(name), ",") {
				params := strings.Split(val, "=")
				if len(params) == 1 {
					if params[0] == subtag {
						return params[0]
					}
				}
				if params[0] == subtag {
					return params[1]
				}
			}
			return ""
		},
		"GetRange": func(rstr string) []string {
			return strings.Split(rstr, "|")
		},
		"Last": func(x int, a []string) bool {
			return x == len(a)-1
		},
	}
	for _, f := range structFuncs {
		for _, s := range structs {
			if s.Name == f.ParamType {
				tmpl = template.Must(template.New("handler.go.tpl").Funcs(funcMap).ParseFiles("templates/handler.go.tpl"))
				tmpl.Execute(resFile, struct {
					Param *Struct
					Func  *StructFunc
				}{s, f})
			}
		}
	}
}
