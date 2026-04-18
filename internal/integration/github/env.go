// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"fmt"
	"os"
	"strings"
)

// LoadConfigFromEnv builds [Config] from environment variables suitable for local
// development and GitHub-hosted runners (see docs/github-integration.md).
//
// Required: GITHUB_TOKEN. Repository: GITHUB_REPOSITORY_OWNER and
// GITHUB_REPOSITORY_NAME, or GITHUB_REPOSITORY as owner/name (e.g. octocat/Hello-World).
func LoadConfigFromEnv() (Config, error) {
	token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
	owner := strings.TrimSpace(os.Getenv("GITHUB_REPOSITORY_OWNER"))
	name := strings.TrimSpace(os.Getenv("GITHUB_REPOSITORY_NAME"))

	if repo := strings.TrimSpace(os.Getenv("GITHUB_REPOSITORY")); repo != "" {
		parts := strings.SplitN(repo, "/", 2)
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			if owner == "" {
				owner = parts[0]
			}
			if name == "" {
				name = parts[1]
			}
		}
	}

	c := Config{
		Repository: Repository{Owner: owner, Name: name},
		Token:      token,
	}
	if token == "" {
		return Config{}, fmt.Errorf("github: %w", &ValidationError{Field: "token", Msg: "GITHUB_TOKEN is required"})
	}
	if err := c.Validate(); err != nil {
		return Config{}, err
	}
	return c, nil
}
