package search

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/dustin/go-humanize"
	"github.com/lwch/logging"
	"github.com/lwch/runtime"
)

type ResourceType int

const (
	TypeUnknown ResourceType = iota
	TypeMovie
	TypeTv
)

type Node struct {
	ID       string
	Name     string
	Type     ResourceType
	Upload   time.Time
	Download string
	Size     runtime.Bytes
	Seeders  uint
	Leechers uint
	Peers    uint
	Detail   string
}

func (m *Model) Query(ctx context.Context, keyword string) ([]Node, error) {
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()
	u, err := m.buildURL(keyword)
	if err != nil {
		return nil, err
	}
	up, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	prefix := up.Scheme + "://" + up.Host
	var list []*cdp.Node
	err = chromedp.Run(ctx,
		chromedp.Navigate(u),
		chromedp.WaitVisible(m.Parser.List),
		chromedp.Nodes(m.Parser.List, &list))
	if err != nil {
		return nil, err
	}
	var ret []Node
	var wg sync.WaitGroup
	var mu sync.Mutex
	wg.Add(len(list))
	for _, node := range list {
		go func(node *cdp.Node) {
			defer wg.Done()
			if node := m.fetch(ctx, node, prefix); node != nil {
				mu.Lock()
				ret = append(ret, *node)
				mu.Unlock()
			}
		}(node)
	}
	wg.Wait()
	return ret, nil
}

func (m *Model) fetch(ctx context.Context, node *cdp.Node, prefix string) *Node {
	var n Node
	var err error
	n.ID, err = m.each(ctx, node, m.Parser.Fields.ID, prefix)
	if err != nil {
		logging.Warning("[>> %s <<]get id from node %s: %v", m.Name, node.FullXPath(), err)
		return nil
	}
	n.Name, err = m.each(ctx, node, m.Parser.Fields.Name, prefix)
	if err != nil {
		logging.Warning("[>> %s <<]get name from node %s: %v", m.Name, node.FullXPath(), err)
		return nil
	}
	t, err := m.each(ctx, node, m.Parser.Fields.Type, prefix)
	if err != nil {
		logging.Warning("[>> %s <<]get type from node %s: %v", m.Name, node.FullXPath(), err)
		return nil
	}
	switch t {
	case "movie":
		n.Type = TypeMovie
	case "tv":
		n.Type = TypeTv
	default:
		n.Type = TypeUnknown
	}
	n.Download, err = m.each(ctx, node, m.Parser.Fields.Download, prefix)
	if err != nil {
		logging.Warning("[>> %s <<]get download from node %s: %v", m.Name, node.FullXPath(), err)
		return nil
	}
	ts, err := m.each(ctx, node, m.Parser.Fields.Upload, prefix)
	if err != nil {
		logging.Warning("[>> %s <<]get upload time from node %s: %v", m.Name, node.FullXPath(), err)
	}
	n.Upload, _ = time.Parse(time.RFC3339, ts)
	size, err := m.each(ctx, node, m.Parser.Fields.Size, prefix)
	if err != nil {
		logging.Warning("[>> %s <<]get size from node %s: %v", m.Name, node.FullXPath(), err)
	}
	sz, err := humanize.ParseBytes(size)
	if err == nil {
		n.Size = runtime.Bytes(sz)
	}
	seeders, err := m.each(ctx, node, m.Parser.Fields.Seeders, prefix)
	if err != nil {
		logging.Warning("[>> %s <<]get seeders from node %s: %v", m.Name, node.FullXPath(), err)
	}
	cnt, _ := strconv.ParseUint(seeders, 10, 64)
	n.Seeders = uint(cnt)
	leechers, err := m.each(ctx, node, m.Parser.Fields.Leechers, prefix)
	if err != nil {
		logging.Warning("[>> %s <<]get leechers from node %s: %v", m.Name, node.FullXPath(), err)
	}
	cnt, _ = strconv.ParseUint(leechers, 10, 64)
	n.Leechers = uint(cnt)
	peers, err := m.each(ctx, node, m.Parser.Fields.Peers, prefix)
	if err != nil {
		logging.Warning("[>> %s <<]get peers from node %s: %v", m.Name, node.FullXPath(), err)
	}
	cnt, _ = strconv.ParseUint(peers, 10, 64)
	n.Peers = uint(cnt)
	n.Detail, err = m.each(ctx, node, m.Parser.Fields.Detail, prefix)
	if err != nil {
		logging.Warning("[>> %s <<]get detail from node %s: %v", m.Name, node.FullXPath(), err)
	}
	if !strings.HasPrefix(n.Detail, "http://") &&
		!strings.HasPrefix(n.Detail, "https://") {
		if !strings.HasPrefix(n.Detail, "/") {
			n.Detail = "/" + n.Detail
		}
		n.Detail = prefix + n.Detail
	}
	return &n
}

func (m *Model) buildURL(keyword string) (string, error) {
	tpl := template.New(m.Name)
	var err error
	tpl, err = tpl.Parse(m.Search)
	if err != nil {
		return "", err
	}
	var args struct {
		Keyword string
	}
	args.Keyword = keyword
	var buf bytes.Buffer
	err = tpl.Execute(&buf, args)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (m *Model) each(ctx context.Context, node *cdp.Node, field Field, prefix string) (ret string, err error) {
	defer func() {
		ret = strings.ReplaceAll(ret, "\xc2\xa0", " ") // &nbsp; => space
		ret = strings.TrimSpace(ret)
		if err == nil {
			ret = m.mapReduce(ret, field.Maps)
			ret = m.timeFormats(ret, field.TimeFormats)
		}
	}()
	if len(field.Selector) == 0 {
		return "", nil
	}
	if field.Request != nil {
		addr, err := m.each(ctx, node, *field.Request, prefix)
		if err != nil {
			return "", err
		}
		if !strings.HasPrefix(addr, "http://") &&
			!strings.HasPrefix(addr, "https://") {
			if !strings.HasPrefix(addr, "/") {
				addr = "/" + addr
			}
			addr = prefix + addr
		}
		var cancel context.CancelFunc
		ctx, cancel = chromedp.NewContext(ctx)
		defer cancel()
		timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		err = chromedp.Run(timeout, chromedp.Navigate(addr))
		if err != nil {
			return "", err
		}
		node = nil
	}
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	var child []*cdp.Node
	if node == nil {
		err = chromedp.Run(ctx, chromedp.Nodes(field.Selector, &child))
	} else {
		err = chromedp.Run(ctx, chromedp.Nodes(field.Selector, &child,
			chromedp.ByQuery, chromedp.FromNode(node)))
	}
	if err != nil {
		return "", err
	}
	if len(child) == 0 {
		return "", errNotfound
	}
	if len(field.Attribute) > 0 {
		str, ok := child[0].Attribute(field.Attribute)
		if !ok {
			return "", fmt.Errorf("attribute %s not found", field.Attribute)
		}
		return str, nil
	}
	var text string
	err = chromedp.Run(ctx, chromedp.Text(child[0].FullXPath(), &text))
	if err != nil {
		return "", err
	}
	return text, nil
}

func (m *Model) mapReduce(str string, maps []MapRule) string {
	for _, rule := range maps {
		if len(rule.Contain) > 0 && strings.Contains(str, rule.Contain) {
			return rule.To
		}
	}
	return str
}

func (m *Model) timeFormats(str string, formats []string) string {
	for _, format := range formats {
		t, err := time.ParseInLocation(format, str, time.Local)
		if err == nil {
			if t.Year() == 0 {
				t = t.AddDate(time.Now().Year(), 0, 0)
			}
			return t.Format(time.RFC3339)
		}
	}
	return str
}
