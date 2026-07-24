package auditor

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestExtractData(t *testing.T) {
	tests := []struct {
		name           string
		html           string
		expectedTitle  string
		expectedH1     int
		expectedImgAlt int
		expectedWords  int
	}{
		{
			name: "Happy Path - Standard SEO Page",
			html: `<html>
					<head>
						<title>Test Page</title>
						<meta name="description" content="This is a test description">
					</head>
					<body>
						<h1>Main Title</h1>
						<p>Hello world from the auditor.</p>
						<img src="logo.png" alt="Company Logo">
						<img src="ad.png"> 
					</body>
				   </html>`,
			expectedTitle:  "Test Page",
			expectedH1:     1,
			expectedImgAlt: 1,
			expectedWords:  7, // "Main Title" (2) + "Hello world from the auditor." (5)
		},
		{
			name:           "Failure Case 1 - Missing SEO Elements",
			html:           `<html><body><p>Just some text without titles or headings.</p></body></html>`,
			expectedTitle:  "",
			expectedH1:     0,
			expectedImgAlt: 0,
			expectedWords:  7,
		},
		{
			name: "Failure Case 2 - Noise Handling (Scripts/Styles)",
			html: `<html>
					<head>
						<style>body { color: red; }</style>
						<script>console.log("hidden code");</script>
					</head>
					<body>
						<h1>Title</h1>
						<p>Visible text.</p>
					</body>
				   </html>`,
			expectedTitle:  "",
			expectedH1:     1,
			expectedImgAlt: 0,
			expectedWords:  3, // Only "Title" and "Visible", "text."
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _ := html.Parse(strings.NewReader(tt.html))
			report := &Report{}

			extractData(doc, report, false)

			if report.Title != tt.expectedTitle {
				t.Errorf("%s: Title = %v, want %v", tt.name, report.Title, tt.expectedTitle)
			}
			if report.H1Count != tt.expectedH1 {
				t.Errorf("%s: H1Count = %v, want %v", tt.name, report.H1Count, tt.expectedH1)
			}
			if report.ImagesMissingAlt != tt.expectedImgAlt {
				t.Errorf("%s: ImagesMissingAlt = %v, want %v", tt.name, report.ImagesMissingAlt, tt.expectedImgAlt)
			}
			if report.WordCount != tt.expectedWords {
				t.Errorf("%s: WordCount = %v, want %v", tt.name, report.WordCount, tt.expectedWords)
			}
		})
	}
}
