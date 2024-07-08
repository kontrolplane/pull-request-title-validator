package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var desiredFormat string = "<type>(optional: <scope>): <message>"

type PullRequest struct {
	Title string `json:"title"`
}

type Event struct {
	PullRequest PullRequest `json:"pull_request"`
}

// The pull-request-title-validator function mankes sure that for each pull request created the
// title of the pull request adheres to a desired structure, in this case convention commit style.
func main() {
	githubEventName := os.Getenv("GITHUB_EVENT_NAME")
	githubEventPath := os.Getenv("GITHUB_EVENT_PATH")
	conventionTypes := []string{"fix", "feat", "chore", "docs", "build", "ci", "refactor", "perf", "test"}

	if githubEventName != "pull_request" && githubEventName != "pull_request_target" {
		fmt.Printf("Error: the 'pull_request' trigger type should be used, received '%s'\n", githubEventName)
		os.Exit(1)
	}

	title := fetchTitle(githubEventPath)
	titleType, titleScope, titleMessage := splitTitle(title)

	if err := checkAgainstConventionTypes(titleType, conventionTypes); err != nil {
		fmt.Printf("The type passed '%s' is not present in the types allowed by the convention: %s\n", titleType, conventionTypes)
		os.Exit(1)
	}

	fmt.Printf("commit title type used: %s\n", titleType)
	fmt.Printf("commit title scope used: %s\n", titleScope)
	fmt.Printf("commit title message used: %s\n\n", titleMessage)
	fmt.Printf("the commit message adheres to the configured standard")
}

func fetchTitle(githubEventPath string) string {

	var event Event
	var eventData []byte
	var err error

	if eventData, err = os.ReadFile(githubEventPath); err != nil {
		fmt.Printf("Problem reading the event json file: %v\n", err)
		os.Exit(1)
	}

	if err = json.Unmarshal(eventData, &event); err != nil {
		fmt.Printf("Failed to unmarshal JSON: %v", err)
		os.Exit(1)
	}

	return event.PullRequest.Title
}

func splitTitle(title string) (titleType string, titleScope string, titleMessage string) {

	// this part of the function extracts the type
	if index := strings.Index(title, "("); strings.Contains(title, "(") {
		titleType = title[:index]
	} else if index := strings.Index(title, ":"); strings.Contains(title, ":") {
		titleType = title[:index]
	} else {
		fmt.Println("No type was included in the pull request title.")
		fmt.Println(desiredFormat)
		os.Exit(1)
	}

	// this part of the function extracts the optional scope
	if strings.Contains(title, "(") && strings.Contains(title, ")") {
		scope := regexp.MustCompile(`\(([^)]+)\)`)
		titleScope = scope.FindStringSubmatch(title)[1]
	}

	// this part of the function extracts the message
	if strings.Contains(title, ":") {
		titleMessage = strings.SplitAfter(title, ":")[1]
		titleMessage = strings.TrimSpace(titleMessage)
	} else {
		fmt.Println("No message was included in the pull request title.")
		fmt.Println(desiredFormat)
		os.Exit(1)
	}

	return
}

func checkAgainstConventionTypes(titleType string, conventionTypes []string) error {
	for _, conventionType := range conventionTypes {
		if titleType == conventionType {
			return nil
		}
	}

	return fmt.Errorf("the type passed '%s' is not present in the types allowed by the convention: %s", titleType, conventionTypes)
}
