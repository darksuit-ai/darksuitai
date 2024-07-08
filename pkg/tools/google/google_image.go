package google

import (
	"encoding/json"
	"fmt"
	exp "github.com/darksuit-ai/darksuitai/internal/exceptions"
	utl "github.com/darksuit-ai/darksuitai/internal/utilities"
	"io"
	"net/http"
	"strings"
)

type SearchParameters struct {
	Q      string `json:"q"`
	Type   string `json:"type"`
	Engine string `json:"engine"`
	Num    int    `json:"num"`
}

type Image struct {
	Title           string `json:"title"`
	ImageUrl        string `json:"imageUrl"`
	ImageWidth      int    `json:"imageWidth"`
	ImageHeight     int    `json:"imageHeight"`
	ThumbnailUrl    string `json:"thumbnailUrl"`
	ThumbnailWidth  int    `json:"thumbnailWidth"`
	ThumbnailHeight int    `json:"thumbnailHeight"`
	Source          string `json:"source"`
	Domain          string `json:"domain"`
	Link            string `json:"link"`
	GoogleUrl       string `json:"googleUrl"`
	Position        int    `json:"position"`
}

type ImageSearchResult struct {
	SearchParameters SearchParameters `json:"searchParameters"`
	Images           []Image          `json:"images"`
}

// googleImages fetches image URLs from Google Images based on the given query
func googleImages(query string) *[]string {
	query = strings.Trim(strings.TrimSpace(query), `"`)
	url := "https://google.serper.dev/images"
	q := utl.ConcatWords([]byte(``), []byte(`{"q":`), []byte(fmt.Sprintf(`"%s"`, query)), []byte(`}`))
	payload := strings.NewReader(q)
	req, err := http.NewRequest("POST", url, payload)

	if err != nil {
		exp.Loggers.System.Warn(err.Error())
		return nil
	}

	// Set the request headers
	req.Header.Set("Content-Type", "application/json")
	// Set the request headers
	req.Header.Set("X-API-KEY", API_KEY)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		exp.Loggers.System.Warn(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		exp.Loggers.System.Warn(err)
		return nil
	}
	if resp.StatusCode == 200 {
		var imageSearch ImageSearchResult
		json_err := json.Unmarshal(body, &imageSearch)
		if json_err != nil {
			exp.Loggers.System.Warn(json_err.Error())
			return nil
		}

		imageSlice := []string{}
		for _, img := range imageSearch.Images {
			imageSlice = append(imageSlice, img.ImageUrl)
		}

		return &imageSlice
	} else if resp.StatusCode == 401 {
		var errorResponse any
		err = json.Unmarshal(body, &errorResponse)
		if err != nil {
			exp.Loggers.System.Warn(errorResponse)
		}
		return nil
	} else if resp.StatusCode == 400 {
		return nil
	} else {
		return nil
	}

}
