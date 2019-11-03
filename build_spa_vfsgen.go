// +build ignore

package main

import (
	"log"
	"net/http"

	"github.com/shurcooL/vfsgen"
)

func main() {
	var fs http.FileSystem = http.Dir("spa/build")
	err := vfsgen.Generate(fs, vfsgen.Options{
		VariableName: "spaAssets",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
