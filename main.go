package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"log/slog"

	"github.com/caarlos0/env"
)

var desiredFormat string = "<type>(optional: <scope>): <message>"
var defaultConventionTypes []string = []string{"fix", "feat", "chore", "docs", "build", "ci", "refactor", "perf", "test"}

type config struct {
	githubEventName string `env:"GITHUB_EVENT_NAME"`
	githubEventPath string `env:"GITHUB_EVENT_PATH"`
	types           string `env:"INPUT_TYPES"`
	scope           string `env:"INPUT_SCOPE"`
}

type PullRequest struct {
	Title string `json:"title"`
}

type Event struct {
	PullRequest PullRequest `json:"pull_request"`
}

// The pull-request-title-validator function mankes sure that for each pull request created the
// title of the pull request adheres to a desired structure, in this case convention commit style.
func main() {

	var cfg config
	if err := env.Parse(&cfg); err != nil {
		fmt.Printf("unable to parse the environment variables: %v", err)
		os.Exit(1)
	}

	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelInfo,
	})
	logger := slog.New(logHandler)

	logger.Info("starting pull-request-title-validator", slog.String("event", cfg.githubEventName))

	if cfg.githubEventName != "pull_request" && cfg.githubEventName != "pull_request_target" {
		logger.Error("invalid event type", slog.String("event", cfg.githubEventName))
		os.Exit(1)
	}

	title := fetchTitle(logger, cfg.githubEventPath)
	titleType, titleScope, titleMessage := splitTitle(logger, title)

	parsedTypes := parseTypes(logger, cfg.types, defaultConventionTypes)
	parsedScope := parseScopes(logger, cfg.scope)

	if err := checkAgainstConventionTypes(logger, titleType, parsedTypes); err != nil {
		logger.Error("error while checking the type against the allowed types",
			slog.String("input", cfg.githubEventName),
			slog.Any("conventionTypes", parsedTypes),
		)
		os.Exit(1)
	}

	if err := checkAgainstScopes(logger, titleScope, parsedScope); err != nil && len(parsedScope) >= 1 {
		logger.Error("error while checking the scope against the allowed scopes", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Info("commit title validated successfully",
		slog.String("type", titleType),
		slog.String("scope", titleScope),
		slog.String("message", titleMessage),
	)
	logger.Info("the commit message adheres to the configured standard")
}

func fetchTitle(logger *slog.Logger, githubEventPath string) string {
	var event Event
	var eventData []byte
	var err error

	if eventData, err = os.ReadFile(githubEventPath); err != nil {
		logger.Error("Problem reading the event JSON file", slog.String("path", githubEventPath), slog.Any("error", err))
		os.Exit(1) // You might want to return an empty string or handle this error upstream instead.
	}

	if err = json.Unmarshal(eventData, &event); err != nil {
		logger.Error("Failed to unmarshal JSON", slog.Any("error", err))
		os.Exit(1)
	}

	return event.PullRequest.Title
}

func splitTitle(logger *slog.Logger, title string) (titleType string, titleScope string, titleMessage string) {
	if index := strings.Index(title, "("); strings.Contains(title, "(") {
		titleType = title[:index]
	} else if index := strings.Index(title, ":"); strings.Contains(title, ":") {
		titleType = title[:index]
	} else {
		logger.Error("No type was included in the pull request title.", slog.String("desired format", desiredFormat))
		os.Exit(1)
	}

	if strings.Contains(title, "(") && strings.Contains(title, ")") {
		scope := regexp.MustCompile(`\(([^)]+)\)`)
		if matches := scope.FindStringSubmatch(title); len(matches) > 1 {
			titleScope = matches[1]
		}
	}

	if strings.Contains(title, ":") {
		titleMessage = strings.SplitAfter(title, ":")[1]
		titleMessage = strings.TrimSpace(titleMessage)
	} else {
		logger.Error("No message was included in the pull request title.", slog.String("desired format", desiredFormat))
		os.Exit(1)
	}

	return
}

func checkAgainstConventionTypes(logger *slog.Logger, titleType string, conventionTypes []string) error {
	for _, conventionType := range conventionTypes {
		if titleType == conventionType {
			return nil
		}
	}
	logger.Error("Type not allowed by the convention", slog.String("type", titleType), slog.Any("allowedTypes", conventionTypes))
	return fmt.Errorf("type '%s' is not allowed", titleType)
}

func checkAgainstScopes(logger *slog.Logger, titleScope string, scopes []string) error {
	for _, scope := range scopes {
		if regexp.MustCompile("(?i)" + scope + "$").MatchString(titleScope) {
			return nil
		}
	}
	logger.Error("Scope not allowed", slog.String("scope", titleScope), slog.Any("allowedScopes", scopes))
	return fmt.Errorf("scope '%s' is not allowed", titleScope)
}

func parseTypes(logger *slog.Logger, input string, fallback []string) []string {
	if input == "" {
		logger.Warn("No custom list of commit types passed, using fallback.")
		return fallback
	}
	types := strings.Split(input, ",")
	for i := range types {
		types[i] = strings.TrimSpace(types[i])
	}
	return types
}

func parseScopes(logger *slog.Logger, input string) []string {
	if input == "" {
		logger.Warn("No custom list of commit scopes passed, using fallback.")
		return []string{}
	}
	scopes := strings.Split(input, ",")
	for i := range scopes {
		scopes[i] = strings.TrimSpace(scopes[i])
	}
	return scopes
}
