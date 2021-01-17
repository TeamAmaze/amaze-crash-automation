package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"
)

// Environment values, set these in your deployment
type Environment struct {
	repoOwner     string // add GITHUB_REPO_OWNER in deployment variables
	repoName      string // add GITHUB_REPO_NAME in deployment variables
	apiToken      string // add API_TOKEN in deployment variables
	issueAPIToken string // add ISSUE_API_TOKEN in deployment variables
	issueURI      string // add ISSUE_URI in deployment variables
}

// IssueResponse created by github api
type IssueResponse struct {
	Number  int    `json:"number,omitempty"`
	URL     string `json:"html_url,omitempty"`
	Body    string `json:"body,omitempty"`
	Message string `json:"message,omitempty"`
	Errors  []struct {
		Value    interface{} `json:"value"`
		Resource string      `json:"resource"`
		Field    string      `json:"field"`
		Code     string      `json:"code"`
	} `json:"errors,omitempty"`
}

// IssueRequest github issue request
type IssueRequest struct {
	Title string `json:"title"`
	Body  string `json:"body,omitempty"`
}

var (
	environment Environment
	httpClient  = &http.Client{
		Timeout: time.Second * 10,
	}
	regexpCrash *regexp.Regexp

	// GithubIssueURI URL for creating GitHub issue
	GithubIssueURI string
)

func init() {
	environment.repoOwner = os.Getenv("GITHUB_REPO_OWNER")
	environment.repoName = os.Getenv("GITHUB_REPO_NAME")
	environment.apiToken = os.Getenv("API_TOKEN")
	environment.issueAPIToken = os.Getenv("ISSUE_API_TOKEN")
	environment.issueURI = os.Getenv("ISSUE_URI")
	GithubIssueURI = fmt.Sprintf(GithubAPIBase+"/repos/%v/%v/issues?state=all&per_page=100", environment.repoOwner, environment.repoName)
	regexpCrash = regexp.MustCompile("<p>((.|\n)*)</details>")
}

// CreateIssue main function responsible for creating GitHub issue
func CreateIssue(w http.ResponseWriter, r *http.Request) {
	validRequest, issueRequest, channel := isRequestValid(r)
	if !validRequest {
		log.Printf("Invalid request with url %v and request %v", r.URL, r)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	issueResponse := createGithubIssue(environment, &issueRequest, channel)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(issueResponse)
}

func isRequestValid(r *http.Request) (bool, IssueRequest, string) {
	var issueRequest IssueRequest
	body, _ := ioutil.ReadAll(r.Body)
	log.Printf("Processing request for isRequestValid %v", string(body))
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&issueRequest)
	fatal(err)

	params := r.URL.Query()
	if environment.apiToken != params.Get("token") {
		log.Printf("Invalid request: token doesn't match env %v", environment.apiToken)
		return false, IssueRequest{}, ""
	}

	if params.Get("channel") == "" {
		log.Printf("Invalid request: channel param not present")
		return false, IssueRequest{}, ""
	}

	if issueRequest.Title == "" {
		log.Printf("Crash message empty issue title")
		return false, IssueRequest{}, ""
	}
	if issueRequest.Body == "" {
		log.Printf("Crash message empty issue body")
		return false, IssueRequest{}, ""
	}
	channel := params.Get("channel")
	return true, issueRequest, channel
}

func main() {
	http.HandleFunc("/", CreateIssue)
	log.Fatal(http.ListenAndServe(":10000", nil))
}

func fatal(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
