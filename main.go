package main

import (
	"flag"
	"fmt"
	"likemovie/internal/app"
	"os"
)

func main() {
	search := flag.String("search", "", "search engine rules")
	debug := flag.Bool("debug", false, "run in debug mode")
	flag.Parse()

	if len(*search) == 0 {
		fmt.Println("missing -search param")
		os.Exit(1)
	}

	app := app.New(*debug)
	app.Init(*search)
}
