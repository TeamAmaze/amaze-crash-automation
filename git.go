package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

// GithubAPIBase github api uri
var GithubAPIBase = "https://api.github.com"

func createGithubIssue(env Environment, issueRequest *IssueRequest, channel string) IssueResponse {
	versionMatch := parseBodyAndCheckVersion(env, issueRequest.Body)
	if !versionMatch {
		log.Printf("Version not supported for request body %v", issueRequest.Body)
		return IssueResponse{
			IsUnofficial: true,
		}
	}
	found, issueResponse := parseBodyAndFindExisting(issueRequest.Body)
	if found {
		log.Printf("Found existing issue at github %v", issueResponse.Body)
		return issueResponse
	}
	log.Printf("No existing issue at found, proceeding to create new for %v", issueRequest.Body)
	request := getCreateIssueRequest(env, issueRequest, channel)
	body := processRequest(request)
	err3 := json.NewDecoder(bytes.NewReader(body)).Decode(&issueResponse)
	fatal(err3)
	return issueResponse
}

func parseBodyAndCheckVersion(env Environment, requestBody string) bool {
	match := regexpVersion.FindStringSubmatch(requestBody)
	if len(match) == 0 {
		log.Printf("Request body doesn't have a version %v. Proceeding to create issue", requestBody)
		return true
	}
	version := match[1]
	return strings.Contains(env.supportedVersions, version)
}

func parseBodyAndFindExisting(requestBody string) (bool, IssueResponse) {
	match := regexpCrash.FindStringSubmatch(requestBody)
	if len(match) == 0 {
		log.Printf("Request body doesn't have a crash %v. Proceeding to create issue", requestBody)
		return false, IssueResponse{}
	}
	requestBodyCrash := match[1]
	log.Printf("Extracted crash from request body, proceed to find all issues %v", requestBodyCrash)
	findAllIssueRequest := getAllIssuesRequest(1)
	body := processRequest(findAllIssueRequest)
	var findAllIssueResponseList []IssueResponse
	err := json.NewDecoder(bytes.NewReader(body)).Decode(&findAllIssueResponseList)
	fatal(err)
	found, issueResponse := parseIssueResponseListForBody(findAllIssueResponseList, requestBodyCrash)
	for i := 2; !found && len(findAllIssueResponseList) != 0; i++ {
		findAllIssueRequest = getAllIssuesRequest(i)
		body = processRequest(findAllIssueRequest)
		err = json.NewDecoder(bytes.NewReader(body)).Decode(&findAllIssueResponseList)
		fatal(err)
		found, issueResponse = parseIssueResponseListForBody(findAllIssueResponseList, requestBodyCrash)
	}
	return found, issueResponse
}

func parseIssueResponseListForBody(issueResponseList []IssueResponse, requestCrashBody string) (bool, IssueResponse) {
	if len(issueResponseList) == 0 {
		log.Printf("Issue response list empty for crashBody %v, returning", requestCrashBody)
		return false, IssueResponse{}
	}
	var issueResponse IssueResponse
	requestCrashFirst, requestCrashSecond := getCrashLines(requestCrashBody)
	log.Printf("Found request crash first line %v, second line %v", requestCrashFirst, requestCrashSecond)
	if len(requestCrashSecond) == 0 {
		log.Printf("Didn't find package name com.amaze.filemanager is stacktrack, return")
		return true, IssueResponse{
			IsUnofficial: true,
		}
	}
	var found bool = false
	for _, element := range issueResponseList {
		match := regexpCrash.FindStringSubmatch(element.Body)
		if len(match) != 0 {
			crashBody := match[1]
			crashFirst, crashSecond := getCrashLines(crashBody)
			if crashFirst == requestCrashFirst && crashSecond == requestCrashSecond {
				log.Printf("Found a match for crash first line %v, second line %v", crashFirst, crashSecond)
				issueResponse = element
				found = true
				break
			}
		}
	}
	return found, issueResponse
}

func getCrashLines(crashBody string) (string, string) {
	var firstLine string
	var secondLine string
	for _, line := range strings.Split(crashBody, "\n") {
		if len(line) != 0 && len(firstLine) == 0 && line != "```" &&
			!strings.Contains(line, "{") && !strings.Contains(line, "}") && !strings.Contains(line, "$") {
			firstLine = strings.ReplaceAll(line, "at ", "")
		}
		if len(line) != 0 && len(secondLine) == 0 && line != "```" && strings.Contains(line, "com.amaze.filemanager") {
			secondLine = strings.ReplaceAll(line, "at ", "")
		}
		if len(firstLine) != 0 && len(secondLine) != 0 {
			break
		}
	}
	log.Printf("Found lines for crash body, first %v, second %v", firstLine, secondLine)
	return strings.TrimSpace(firstLine), strings.TrimSpace(secondLine)
}

func getCreateIssueRequest(env Environment, issueRequest *IssueRequest, channel string) *http.Request {
	postBody, err := json.Marshal(issueRequest)
	fatal(err)
	log.Printf("Create request for new issue with request body %v", string(postBody))
	requestBody := bytes.NewBuffer(postBody)
	request, err2 := http.NewRequest("POST", fmt.Sprintf("%v?token=%v&channel=%v", env.issueURI, env.issueAPIToken, channel), requestBody)
	fatal(err2)
	return request
}

func getAllIssuesRequest(page int) *http.Request {
	log.Printf("Create request to find all issues for url %v page %v", GithubIssueURI, page)
	request, err := http.NewRequest("GET", fmt.Sprintf(GithubIssueURI+"&page=%v", page), nil)
	fatal(err)
	return request
}

func processRequest(request *http.Request) []byte {
	log.Printf("Final request %v", request)
	resp, err2 := httpClient.Do(request)
	fatal(err2)
	body, err4 := ioutil.ReadAll(resp.Body)
	fatal(err4)
	defer resp.Body.Close()
	return body
}
