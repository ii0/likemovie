package main

import (
	"context"

	"github.com/chromedp/chromedp"
	"github.com/lwch/runtime"
)

func main() {
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()
	err := chromedp.Run(ctx,
		chromedp.Navigate("https://www.baidu.com"))
	runtime.Assert(err)
}
