package search

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
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
	url, err := m.buildURL(keyword)
	if err != nil {
		return nil, err
	}
	var list []*cdp.Node
	err = chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitVisible(m.Parser.List),
		chromedp.Nodes(m.Parser.List, &list))
	if err != nil {
		return nil, err
	}
	var ret []Node
	for _, node := range list {
		var n Node
		n.ID, err = m.each(ctx, node, m.Parser.Fields.ID)
		if err != nil {
			logging.Warning("[>> %s <<]get id from node %s: %v", m.Name, node.FullXPath(), err)
			continue
		}
		n.Name, err = m.each(ctx, node, m.Parser.Fields.Name)
		if err != nil {
			logging.Warning("[>> %s <<]get name from node %s: %v", m.Name, node.FullXPath(), err)
			continue
		}
		t, err := m.each(ctx, node, m.Parser.Fields.Type)
		if err != nil {
			logging.Warning("[>> %s <<]get type from node %s: %v", m.Name, node.FullXPath(), err)
			continue
		}
		switch t {
		case "movie":
			n.Type = TypeMovie
		case "tv":
			n.Type = TypeTv
		default:
			n.Type = TypeUnknown
		}
		n.Download, err = m.each(ctx, node, m.Parser.Fields.Download)
		if err != nil {
			logging.Warning("[>> %s <<]get download from node %s: %v", m.Name, node.FullXPath(), err)
			continue
		}
		ts, err := m.each(ctx, node, m.Parser.Fields.Upload)
		if err != nil {
			logging.Warning("[>> %s <<]get upload time from node %s: %v", m.Name, node.FullXPath(), err)
		}
		n.Upload, _ = time.Parse(time.RFC3339, ts)
		size, err := m.each(ctx, node, m.Parser.Fields.Size)
		if err != nil {
			logging.Warning("[>> %s <<]get size from node %s: %v", m.Name, node.FullXPath(), err)
		}
		sz, err := humanize.ParseBytes(size)
		if err == nil {
			n.Size = runtime.Bytes(sz)
		}
		seeders, err := m.each(ctx, node, m.Parser.Fields.Seeders)
		if err != nil {
			logging.Warning("[>> %s <<]get seeders from node %s: %v", m.Name, node.FullXPath(), err)
		}
		cnt, _ := strconv.ParseUint(seeders, 10, 64)
		n.Seeders = uint(cnt)
		leechers, err := m.each(ctx, node, m.Parser.Fields.Leechers)
		if err != nil {
			logging.Warning("[>> %s <<]get leechers from node %s: %v", m.Name, node.FullXPath(), err)
		}
		cnt, _ = strconv.ParseUint(leechers, 10, 64)
		n.Leechers = uint(cnt)
		peers, err := m.each(ctx, node, m.Parser.Fields.Peers)
		if err != nil {
			logging.Warning("[>> %s <<]get peers from node %s: %v", m.Name, node.FullXPath(), err)
		}
		cnt, _ = strconv.ParseUint(peers, 10, 64)
		n.Peers = uint(cnt)
		n.Detail, err = m.each(ctx, node, m.Parser.Fields.Detail)
		if err != nil {
			logging.Warning("[>> %s <<]get detail from node %s: %v", m.Name, node.FullXPath(), err)
		}
		ret = append(ret, n)
	}
	return ret, nil
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

func (m *Model) each(ctx context.Context, node *cdp.Node, field Field) (ret string, err error) {
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
	// TODO: request
	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, time.Second)
	defer cancel()

	var child []*cdp.Node
	err = chromedp.Run(ctx, chromedp.Nodes(field.Selector, &child, chromedp.ByQuery, chromedp.FromNode(node)))
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
