package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"github.com/caarlos0/env"
)

const (
	desiredFormat = "<type>(optional: <scope>): <message>"
)

var defaultConventionTypes = []string{
	"fix", "feat", "chore", "docs", "build", "ci", "refactor", "perf", "test",
}

type Config struct {
	GithubEventName string `env:"GITHUB_EVENT_NAME"`
	GithubEventPath string `env:"GITHUB_EVENT_PATH"`
	Types           string `env:"INPUT_TYPES"`
	Scopes          string `env:"INPUT_SCOPES"`
}

type PullRequest struct {
	Title string `json:"title"`
}

type Event struct {
	PullRequest PullRequest `json:"pull_request"`
}

type TitleComponents struct {
	Type    string
	Scope   string
	Message string
}

type Validator struct {
	logger *slog.Logger
	config Config
}

func main() {
	logger := setupLogger()

	cfg, err := loadConfig()
	if err != nil {
		logger.Error("unable to parse environment variables", slog.Any("error", err))
		os.Exit(1)
	}

	validator := &Validator{
		logger: logger,
		config: cfg,
	}

	if err := validator.run(); err != nil {
		os.Exit(1)
	}
}

func setupLogger() *slog.Logger {
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelInfo,
	})
	return slog.New(logHandler)
}

func loadConfig() (Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	return cfg, err
}

func (v *Validator) run() error {
	v.logger.Info("starting pull-request-title-validator",
		slog.String("event", v.config.GithubEventName))

	if err := v.validateEventType(); err != nil {
		return err
	}

	title, err := v.fetchTitle()
	if err != nil {
		return err
	}

	components, err := v.parseTitle(title)
	if err != nil {
		return err
	}

	if err := v.validateTitle(components); err != nil {
		return err
	}

	v.logger.Info("commit title validated successfully",
		slog.String("type", components.Type),
		slog.String("scope", components.Scope),
		slog.String("message", components.Message),
	)
	v.logger.Info("the commit message adheres to the configured standard")

	return nil
}

func (v *Validator) validateEventType() error {
	if v.config.GithubEventName != "pull_request" && v.config.GithubEventName != "pull_request_target" {
		v.logger.Error("invalid event type", slog.String("event", v.config.GithubEventName))
		return fmt.Errorf("invalid event type: %s", v.config.GithubEventName)
	}
	return nil
}

func (v *Validator) fetchTitle() (string, error) {
	eventData, err := os.ReadFile(v.config.GithubEventPath)
	if err != nil {
		v.logger.Error("problem reading the event JSON file",
			slog.String("path", v.config.GithubEventPath),
			slog.Any("error", err))
		return "", err
	}

	var event Event
	if err := json.Unmarshal(eventData, &event); err != nil {
		v.logger.Error("failed to unmarshal JSON", slog.Any("error", err))
		return "", err
	}

	return event.PullRequest.Title, nil
}

func (v *Validator) parseTitle(title string) (*TitleComponents, error) {
	// Split title into prefix (type/scope) and message parts using colon as separator
	prefix, message, found := strings.Cut(title, ":")
	if !found {
		v.logger.Error("title must include a message after the colon",
			slog.String("desired format", desiredFormat),
			slog.String("title", title))
		return nil, fmt.Errorf("title missing colon separator")
	}

	// Clean up the message part
	titleMessage := strings.TrimSpace(message)

	// Extract type and scope from the prefix
	titleType, titleScope := extractTypeAndScope(prefix)

	// Validate that we found a type
	if titleType == "" {
		v.logger.Error("title must include a type",
			slog.String("desired format", desiredFormat),
			slog.String("title", title))
		return nil, fmt.Errorf("title missing type")
	}

	return &TitleComponents{
		Type:    titleType,
		Scope:   titleScope,
		Message: titleMessage,
	}, nil
}

func extractTypeAndScope(prefix string) (titleType string, titleScope string) {
	prefix = strings.TrimSpace(prefix)

	// Check if prefix contains a scope in parentheses
	if strings.Contains(prefix, "(") && strings.Contains(prefix, ")") {
		// Extract scope using regex
		scopeRegex := regexp.MustCompile(`\(([^)]+)\)`)

		if matches := scopeRegex.FindStringSubmatch(prefix); len(matches) > 1 {
			titleScope = matches[1]
			titleType = strings.TrimSpace(strings.Split(prefix, "(")[0])
			return titleType, titleScope
		}
	}

	// If no scope found or invalid format, use entire prefix as type
	titleType = prefix
	return titleType, titleScope
}

func (v *Validator) validateTitle(components *TitleComponents) error {
	parsedTypes := v.parseTypes()
	parsedScopes := v.parseScopes()

	if err := v.validateType(components.Type, parsedTypes); err != nil {
		v.logger.Error("error while checking the type against the allowed types",
			slog.String("event name", v.config.GithubEventName),
			slog.String("event path", v.config.GithubEventPath),
			slog.Any("convention types", parsedTypes),
		)
		return err
	}

	if err := v.validateScope(components.Scope, parsedScopes); err != nil && len(parsedScopes) >= 1 {
		v.logger.Error("error while checking the scope against the allowed scopes",
			slog.Any("error", err))
		return err
	}

	return nil
}

func (v *Validator) validateType(titleType string, allowedTypes []string) error {
	for _, allowedType := range allowedTypes {
		if titleType == allowedType {
			return nil
		}
	}

	v.logger.Error("type not allowed by the convention",
		slog.String("type", titleType),
		slog.Any("allowedTypes", allowedTypes))
	return fmt.Errorf("type '%s' is not allowed", titleType)
}

func (v *Validator) validateScope(titleScope string, allowedScopes []string) error {
	for _, scope := range allowedScopes {
		if regexp.MustCompile("(?i)" + scope + "$").MatchString(titleScope) {
			return nil
		}
	}

	return fmt.Errorf("scope '%s' is not allowed", titleScope)
}

func (v *Validator) parseTypes() []string {
	if v.config.Types == "" {
		v.logger.Warn("no custom list of commit types passed, using fallback")
		return defaultConventionTypes
	}

	return parseCommaSeparatedList(v.config.Types)
}

func (v *Validator) parseScopes() []string {
	if v.config.Scopes == "" {
		v.logger.Warn("no custom list of commit scopes passed, using fallback")
		return []string{}
	}

	return parseCommaSeparatedList(v.config.Scopes)
}

func parseCommaSeparatedList(input string) []string {
	items := strings.Split(input, ",")
	for i := range items {
		items[i] = strings.TrimSpace(items[i])
	}
	return items
}
