package backlog

import (
	"fmt"
	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
	"strings"
)

type SingleBacklogConfig struct {
	Icon       string
	InputPages []string
	OutputPage string
}

type Config struct {
	FocusPage string
	Backlogs  []SingleBacklogConfig
}

type ConfigReader interface {
	ReadConfig() (*Config, error)
}

type pageConfigReader struct {
	graph    *logseq.Graph
	rootPage string
}

// NewPageConfigReader creates a new ConfigReader that reads the backlog configuration from a Logseq page.
func NewPageConfigReader(graph *logseq.Graph, rootPage string) ConfigReader {
	return &pageConfigReader{
		graph:    graph,
		rootPage: rootPage,
	}
}

// ReadConfig reads the backlog configuration from a Logseq page.
func (p *pageConfigReader) ReadConfig() (*Config, error) {
	rootPage, err := p.graph.OpenPage(p.rootPage)
	if err != nil {
		return nil, fmt.Errorf("failed to open backlog page: %w", err)
	}

	var backlogs []SingleBacklogConfig

	for _, block := range rootPage.Blocks() {
		var inputPages []string

		firstPage := ""

		// TODO: simplify and replace by FilterDeep after a test is added
		block.Children().FindDeep(func(n content.Node) bool {
			link := ""
			if pageLink, ok := n.(*content.PageLink); ok {
				link = pageLink.To
			} else if tag, ok := n.(*content.Hashtag); ok {
				link = tag.To
			}

			if link == "" {
				return false
			}

			// Skip this page if it's a link to a backlog
			if strings.HasPrefix(link, p.rootPage) {
				return false
			}

			if firstPage == "" {
				firstPage = p.rootPage + "/" + link
			}

			inputPages = append(inputPages, link)

			return false
		})

		if len(inputPages) > 0 {
			backlogs = append(backlogs, SingleBacklogConfig{
				Icon:       "",
				InputPages: inputPages,
				OutputPage: firstPage,
			})
		}
	}

	if len(backlogs) == 0 {
		fmt.Println("no pages found in the backlog")
	}

	return &Config{FocusPage: p.rootPage + "/Focus", Backlogs: backlogs}, nil
}
