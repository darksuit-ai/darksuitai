package google

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	exp "github.com/darksuit-ai/darksuitai/internal/exceptions"
	bst "github.com/darksuit-ai/darksuitai/pkg/tools"
)

var re = regexp.MustCompile(`[^a-zA-Z0-9]`)

type SearchResult struct {
	Link    string
	Snippet string
	Fact    string
}

type GoogleSerperAPIWrapper struct {
	K                int    `json:"k"`
	Gl               string `json:"gl"`
	Hl               string `json:"hl"`
	Type             string `json:"type"`
	ResultKeyForType map[string]string
	Tbs              string `json:"tbs"`
	SerperAPIKey     string `json:"serper_api_key"`
	AioSession       *http.Client
	wg               sync.WaitGroup
}

const API_KEY string = "1703d54836a59537c8256940b15809260dce88ad"

func NewGoogleSerperAPIWrapper() *GoogleSerperAPIWrapper {
	return &GoogleSerperAPIWrapper{
		K:    10,
		Gl:   "us",
		Hl:   "en",
		Type: "search",
		ResultKeyForType: map[string]string{
			"news":   "news",
			"places": "places",
			"images": "images",
			"search": "organic",
		},
		SerperAPIKey: API_KEY,
		AioSession: &http.Client{
			Timeout: 10 * time.Second,
		},
		wg: sync.WaitGroup{},
	}
}

func (w *GoogleSerperAPIWrapper) Results(query string, kwargs map[string]interface{}) map[string]interface{} {
	return w.googleSerperAPIResults(
		query,
		w.Gl,
		w.Hl,
		w.K,
		w.Tbs,
		w.Type,
		kwargs,
	)
}

func (w *GoogleSerperAPIWrapper) Run(query string, kwargs map[string]interface{}) (*string, any) {
	results := w.googleSerperAPIResults(
		query,
		w.Gl,
		w.Hl,
		w.K,
		w.Tbs,
		w.Type,
		kwargs,
	)

	finalResult := w.parseResults(results)
	searchResp := strings.ReplaceAll(finalResult, "...", ".")
	searchResp = strings.ReplaceAll(searchResp, " ... ", ".")
	finalResponse := searchResp + "\n\nI will also notify the user that I have appended images in my search response."
	return &finalResponse, nil
}

func (w *GoogleSerperAPIWrapper) parseSnippets(results map[string]interface{}) string {
	var searchResults strings.Builder

	if message, ok := results["message"].(string); ok {
		if message == "Not enough credits" {
			return "Not enough credits"
		}
	}

	w.wg.Add(2)
	go func() {
		defer w.wg.Done()
		if answerBox, ok := results["answerBox"].(map[string]interface{}); ok {
			searchResults.WriteString("\n------------- TOP SEARCH RESULT ----------\n")
			if snippet, ok := answerBox["snippet"].(string); ok {
				if link, ok := answerBox["link"].(string); ok {
					searchResults.WriteString(fmt.Sprintf("%s <--> %s\n", snippet, link))
				}
			} else if answer, ok := answerBox["answer"].(string); ok {
				searchResults.WriteString(fmt.Sprintf("Answer: %s\n", answer))
			}
			searchResults.WriteString("\n------------- MORE SEARCH RESULT----------\n")
		}

		if organicResults, ok := results["organic"].([]interface{}); ok {
			for _, result := range organicResults {
				if resultMap, ok := result.(map[string]interface{}); ok {
					link, _ := resultMap["link"].(string)
					link = strings.ReplaceAll(link, "...", "")
					snippet, _ := resultMap["snippet"].(string)
					searchResults.WriteString(fmt.Sprintf("%s <--> %s\n", snippet, link))
				}
			}
		}
	}()

	go func() {
		defer w.wg.Done()
		if peopleAlsoAskResults, ok := results["peopleAlsoAsk"].([]interface{}); ok {
			for _, result := range peopleAlsoAskResults {
				if resultMap, ok := result.(map[string]interface{}); ok {
					link, _ := resultMap["link"].(string)
					link = strings.ReplaceAll(link, "...", "")
					snippet, _ := resultMap["snippet"].(string)
					searchResults.WriteString(fmt.Sprintf("%s <--> %s\n", snippet, link))
				}
			}
		}
	}()

	w.wg.Wait()

	if searchResults.Len() == 0 {
		return "No Google Search Result was found"
	}

	return searchResults.String() + "I will infuse the links to my Final Answer for verifiability."
}

func (w *GoogleSerperAPIWrapper) parseResults(results map[string]interface{}) string {
	//return strings.Join(w.parseSnippets(results), " ")
	return w.parseSnippets(results)
}

func (w *GoogleSerperAPIWrapper) googleSerperAPIResults(
	searchTerm string,
	gl string,
	hl string,
	num int,
	tbs string,
	searchType string,
	kwargs map[string]interface{},
) map[string]interface{} {
	cleaned := strings.Trim(searchTerm, "\\\"")
	cleanedStr := strings.ReplaceAll(re.ReplaceAllString(cleaned, ""), " ", "")

	payload := map[string]string{
		"q": cleanedStr,
	}

	// Convert the payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshaling payload:", err)

	}

	// Create a new HTTP request
	url := "https://google.serper.dev/search"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		exp.Loggers.System.Warn(err)
	}

	// Set the request headers
	req.Header.Set("Content-Type", "application/json")
	// Set the request headers
	req.Header.Set("X-API-KEY", API_KEY)

	resp, err := w.AioSession.Do(req)
	if err != nil {
		exp.Loggers.System.Warn(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		exp.Loggers.System.Warn(err)
	}

	var searchResults map[string]interface{}
	err = json.Unmarshal(body, &searchResults)
	if err != nil {
		exp.Loggers.System.Warn(err)
	}
	return searchResults
}

func googleSearchAndImages(query string, toolsMeta map[string]interface{}) (string, []interface{}) {
	var wg sync.WaitGroup
	var searchResult *string
	var imgResult *[]string
	wg.Add(2)

	go func() {
		defer wg.Done()
		imgResult = googleImages(query)

	}()

	go func() {
		defer wg.Done()
		wrapper := NewGoogleSerperAPIWrapper()
		searchResult, _ = wrapper.Run(query, nil)
	}()

	wg.Wait()
	// Convert imgResult to []interface{}
	var imgResultInterface = make([]interface{}, 1)

	imgResultInterface[0] = append(imgResultInterface, *imgResult)

	return *searchResult, imgResultInterface
}

func Google_tool(description string) map[string]bst.BaseTool {
	if description == "" {
		description = `This tool is useful ONLY in the following circumstances:
	- user is asking about recent/current events or something that requires real-time information (sports scores, news, latest happenings/events, any information deemed recent; beyond your knowledge cutoff year. Recently, the currently year is %d)
	- When you need to look up information about topics, these topics can be a wide range of topics especially about humans, stocks e.g who won the superbowl? is Jim Simons still alive? who is in scotland's eur group this year?.
	- user is asking about some term you are totally unfamiliar with (it might be new)
	- user explicitly asks you to browse, show images or provide links to references
Always rephrase the input word in the best search term to get results.`
	}
	return map[string]bst.BaseTool{
		"Google Search": {
			Name:        "Google Search",
			Description: description,
			ToolFunc: func(query string, meta []interface{}) (string, []interface{}) {
				return googleSearchAndImages(query, nil)
			},
		},
	}
}
