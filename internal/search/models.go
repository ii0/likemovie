package search

import (
	"context"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/lwch/logging"
)

type Models []Model

func (list Models) Query(keyword string) []Node {
	ctx, cancel := chromedp.NewExecAllocator(context.Background(),
		chromedp.NoFirstRun,
		chromedp.Flag("headless", false))
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()
	chromedp.Run(ctx, chromedp.Navigate(""))
	var wg sync.WaitGroup
	wg.Add(len(list))
	var ret []Node
	var m sync.Mutex
	for _, model := range list {
		go func(ctx context.Context, model Model) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(ctx, time.Minute)
			defer cancel()
			nodes, err := model.Query(ctx, keyword)
			if err != nil {
				logging.Error("query [%s] from node: %v", keyword, model.Name)
				return
			}
			m.Lock()
			ret = append(ret, nodes...)
			m.Unlock()
		}(ctx, model)
	}
	wg.Wait()
	return ret
}
