package scrape

import (
	"bytes"
	"fmt"

	"github.com/antchfx/htmlquery"
)

var imagePrefix = "https://pp.tinybeans.com/p/prod/image/upload/t_l/tb/journals"

// ScrapeHTML scans the email text for images, and returns a list of
// image URLS or error
func ScrapeHTML(text []byte) ([]string, error) {

	doc, err := htmlquery.Parse(bytes.NewReader(text))
	if err != nil {
		return nil, err
	}

	// find all img tags with a url that includes the word
	imgs := htmlquery.Find(doc,
		fmt.Sprintf("//img[starts-with(@src, '%s')]", imagePrefix))
	out := []string{}
	for _, img := range imgs {
		for _, attr := range img.Attr {
			if attr.Key == "src" {
				out = append(out, attr.Val)
				break
			}
		}
	}

	return out, nil
}
