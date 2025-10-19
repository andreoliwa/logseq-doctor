package backlog

import (
	"fmt"
	"strings"

	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
	"github.com/andreoliwa/lqd/internal"
)

type SingleBacklogConfig struct {
	BacklogPage string
	Icon        string
	InputPages  []string
}

type Config struct {
	FocusPage string
	Backlogs  []SingleBacklogConfig
}

type ConfigReader interface {
	ReadConfig() (*Config, error)
}

type pageConfigReader struct {
	graph      *logseq.Graph
	configPage string
}

// NewPageConfigReader creates a new ConfigReader that reads the backlog configuration from a Logseq page.
func NewPageConfigReader(graph *logseq.Graph, configPage string) ConfigReader {
	return &pageConfigReader{
		graph:      graph,
		configPage: configPage,
	}
}

// ReadConfig reads the backlog configuration from a Logseq page.
func (p *pageConfigReader) ReadConfig() (*Config, error) { //nolint:cyclop,funlen
	configPage := internal.OpenPage(p.graph, p.configPage)

	var backlogs []SingleBacklogConfig

	for _, block := range configPage.Blocks() {
		var inputPages []string

		firstRegularPage := ""
		backlogPage := ""

		// TODO: simplify and replace by FilterDeep after a test is added
		block.Children().FindDeep(func(node content.Node) bool {
			target := ""
			isLink := false

			if pageLink, ok := node.(*content.PageLink); ok {
				target = pageLink.To
			} else if tag, ok := node.(*content.Hashtag); ok {
				target = tag.To
			} else if link, ok := node.(*content.Link); ok {
				target = link.URL
				isLink = true
			}

			if target == "" {
				return false
			}

			if target == p.configPage {
				return false
			}

			if strings.HasPrefix(target+"/", p.configPage) {
				backlogPage = target

				return false
			} else if isLink {
				// Ignore links to real URLs or any link that doesn't start with the config page
				return false
			}

			if firstRegularPage == "" {
				firstRegularPage = p.configPage + "/" + target
			}

			inputPages = append(inputPages, target)

			return false
		})

		if len(inputPages) > 0 {
			chosenPage := backlogPage
			if chosenPage == "" {
				chosenPage = firstRegularPage
			}

			backlogs = append(backlogs, SingleBacklogConfig{
				BacklogPage: chosenPage,
				Icon:        "",
				InputPages:  inputPages,
			})
		}
	}

	if len(backlogs) == 0 {
		fmt.Println("no pages found in the backlog")
	}

	return &Config{FocusPage: p.configPage + "/Focus", Backlogs: backlogs}, nil
}
