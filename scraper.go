package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type Problem struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Link       string `json:"link"`
	Difficulty string `json:"difficulty"`
}

type ProblemSet struct {
	Name     string    `json:"name"`
	Problems []Problem `json:"problems"`
}

type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
	OperationName string                 `json:"operationName"`
}

type GraphQLResponse struct {
	Data struct {
		FavoriteQuestionList struct {
			Questions []struct {
				Difficulty         string `json:"difficulty"`
				Title              string `json:"title"`
				TitleSlug          string `json:"titleSlug"`
				QuestionFrontendId string `json:"questionFrontendId"`
			} `json:"questions"`
			TotalLength int  `json:"totalLength"`
			HasMore     bool `json:"hasMore"`
		} `json:"favoriteQuestionList"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage: go run scrape_leetcode.go <list-id> <name>\nExample: go run scrape_leetcode.go ajpcecv6 \"Blind 75\"")
	}

	listId := os.Args[1]
	name := os.Args[2]

	fmt.Printf("Fetching list: %s\n", listId)

	// this is what i see in the network tab
	query := `
    query favoriteQuestionList($favoriteSlug: String!, $filter: FavoriteQuestionFilterInput, $filtersV2: QuestionFilterInput, $searchKeyword: String, $sortBy: QuestionSortByInput, $limit: Int, $skip: Int, $version: String = "v2") {
      favoriteQuestionList(
        favoriteSlug: $favoriteSlug
        filter: $filter
        filtersV2: $filtersV2
        searchKeyword: $searchKeyword
        sortBy: $sortBy
        limit: $limit
        skip: $skip
        version: $version
      ) {
        questions {
          difficulty
          questionFrontendId
          title
          titleSlug
        }
        totalLength
        hasMore
      }
    }
    `

	// similar structure acc. to what i recieve
	reqBody := GraphQLRequest{
		Query: query,
		Variables: map[string]interface{}{
			"skip":         0,
			"limit":        100,
			"favoriteSlug": listId,
			"filtersV2": map[string]interface{}{
				"filterCombineType": "ALL",
				"statusFilter": map[string]interface{}{
					"questionStatuses": []string{},
					"operator":         "IS",
				},
				"difficultyFilter": map[string]interface{}{
					"difficulties": []string{},
					"operator":     "IS",
				},
				"languageFilter": map[string]interface{}{
					"languageSlugs": []string{},
					"operator":      "IS",
				},
				"topicFilter": map[string]interface{}{
					"topicSlugs": []string{},
					"operator":   "IS",
				},
				"acceptanceFilter":    map[string]interface{}{},
				"frequencyFilter":     map[string]interface{}{},
				"lastSubmittedFilter": map[string]interface{}{},
				"publishedFilter":     map[string]interface{}{},
				"companyFilter": map[string]interface{}{
					"companySlugs": []string{},
					"operator":     "IS",
				},
				"positionFilter": map[string]interface{}{
					"positionSlugs": []string{},
					"operator":      "IS",
				},
				"premiumFilter": map[string]interface{}{
					"premiumStatus": []string{},
					"operator":      "IS",
				},
			},
			"searchKeyword": "",
			"sortBy": map[string]interface{}{
				"sortField": "CUSTOM",
				"sortOrder": "ASCENDING",
			},
		},
		OperationName: "favoriteQuestionList",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatal("Error marshaling request:", err)
	}

	// req create
	req, err := http.NewRequest("POST", "https://leetcode.com/graphql", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Fatal("Error creating request:", err)
	}

	// very imp. otherwise cloudfare will block
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36")
	req.Header.Set("Referer", "https://leetcode.com")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error making request:", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading response:", err)
	}

	// debug stuff
	if resp.StatusCode != 200 {
		fmt.Printf("Response Status: %d\n", resp.StatusCode)
		fmt.Printf("Response Body: %s\n", string(body))
	}

	// parsing
	var graphQLResp GraphQLResponse
	err = json.Unmarshal(body, &graphQLResp)
	if err != nil {
		log.Fatal("Error parsing response:", err)
	}

	// graphql errors
	if len(graphQLResp.Errors) > 0 {
		log.Fatalf("GraphQL error: %s", graphQLResp.Errors[0].Message)
	}

	// my format
	problemSet := ProblemSet{
		Name:     name,
		Problems: make([]Problem, 0),
	}

	questions := graphQLResp.Data.FavoriteQuestionList.Questions
	for _, q := range questions {
		problem := Problem{
			ID:         q.TitleSlug,
			Name:       q.Title,
			Link:       fmt.Sprintf("https://leetcode.com/problems/%s", q.TitleSlug),
			Difficulty: strings.ToLower(q.Difficulty),
		}
		problemSet.Problems = append(problemSet.Problems, problem)
	}

	// data dir creation
	err = os.MkdirAll("data", 0755)
	if err != nil {
		log.Fatal("Error creating directory:", err)
	}

	// output as json
	outputFile := fmt.Sprintf("data/%s-problems.json", strings.ToLower(strings.ReplaceAll(name, " ", "-")))
	jsonData, err := json.MarshalIndent(problemSet, "", "  ")
	if err != nil {
		log.Fatal("Error marshaling output:", err)
	}

	err = os.WriteFile(outputFile, jsonData, 0644)
	if err != nil {
		log.Fatal("Error writing file:", err)
	}

	fmt.Printf("Successfully scraped %d problems and saved to %s\n", len(problemSet.Problems), outputFile)
}
