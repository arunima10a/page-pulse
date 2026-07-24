package auditor

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// Audit results for a given URL
type Report struct {
	URL              string        `json:"url"`
	Status           int           `json:"status"`
	ResponseTime     time.Duration `json:"response_time_ms"`
	Title            string        `json:"title"`
	MetaDescription  string        `json:"meta_description"`
	H1Count          int           `json:"h1_count"`
	ImagesMissingAlt int           `json:"images_missing_alt"`
	WordCount        int           `json:"word_count"`
	RequestTime      string        `json:"request_time"`
}

// standardized error format for the API
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

func AnalyzeURL(targetURL string) (*Report, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	start := time.Now()

	// Create a new request
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Professional User-Agent - identifies our tool as a standard browser so sites don't block us
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36 PagePulse/1.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	responseTime := time.Since(start)

	// Validate Content-Type: We only care about HTML
	ctype := strings.ToLower(resp.Header.Get("Content-Type"))
	if !strings.Contains(ctype, "text/html") {
		return nil, fmt.Errorf("unsupported content type: %s (expected text/html)", ctype)
	}

	// Check if the server actually returned a success status
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server returned an error status: %d", resp.StatusCode)
	}

	report := &Report{
		URL:          targetURL,
		Status:       resp.StatusCode,
		ResponseTime: responseTime,
		RequestTime:  time.Now().Format(time.RFC3339),
	}

	// Parse the HTML tree
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	extractData(doc, report, false)

	return report, nil
}


func extractData(n *html.Node, report *Report, inBody bool) {
	if n.Type == html.ElementNode {
		// Track if we have entered the body
		if n.Data == "body" {
			inBody = true
		}

		switch n.Data {
		case "title":
			if n.FirstChild != nil {
				report.Title = n.FirstChild.Data
			}
		case "meta":
			var name, content string
			for _, attr := range n.Attr {
				if attr.Key == "name" {
					name = strings.ToLower(attr.Val)
				}
				if attr.Key == "content" {
					content = attr.Val
				}
			}
			if name == "description" {
				report.MetaDescription = content
			}
		case "h1":
			report.H1Count++
		case "img":
			hasAlt := false
			for _, attr := range n.Attr {
				if attr.Key == "alt" {
					hasAlt = true
					break
				}
			}
			if !hasAlt {
				report.ImagesMissingAlt++
			}
		}
	}

	// Only count words if we are inside the <body> and NOT in a script/style tag
	if n.Type == html.TextNode && inBody {
		parent := n.Parent.Data
		if parent != "script" && parent != "style" && parent != "noscript" {
			words := strings.Fields(n.Data)
			report.WordCount += len(words)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractData(c, report, inBody)
	}
}