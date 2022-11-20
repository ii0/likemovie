package main

import (
	"flag"
	"fmt"
	"likemovie/internal/app"
	"os"
)

func main() {
	// ctx, cancel := chromedp.NewExecAllocator(context.Background(),
	// 	chromedp.NoFirstRun,
	// 	chromedp.Flag("headless", false))
	// defer cancel()
	// ctx, cancel = chromedp.NewContext(ctx)
	// defer cancel()
	// err := chromedp.Run(ctx,
	// 	chromedp.Navigate("https://www.baidu.com"))
	// runtime.Assert(err)

	search := flag.String("search", "", "search engine rules")
	flag.Parse()

	if len(*search) == 0 {
		fmt.Println("missing -search param")
		os.Exit(1)
	}

	app := app.New()
	app.Init(*search)
}
