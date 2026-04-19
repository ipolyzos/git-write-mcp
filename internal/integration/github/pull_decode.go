// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"encoding/json"
	"fmt"
)

// DecodePullRequestFromAPI maps a GitHub pulls API JSON object (create or get)
// into [PullRequest]. Unknown fields are ignored.
func DecodePullRequestFromAPI(body []byte) (PullRequest, error) {
	var raw struct {
		Number  int    `json:"number"`
		Title   string `json:"title"`
		State   string `json:"state"`
		HTMLURL string `json:"html_url"`
		Head    struct {
			SHA string `json:"sha"`
		} `json:"head"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return PullRequest{}, fmt.Errorf("github: decode pull request: %w", err)
	}
	return PullRequest{
		Number:  raw.Number,
		Title:   raw.Title,
		State:   raw.State,
		HTMLURL: raw.HTMLURL,
		HeadSHA: raw.Head.SHA,
	}, nil
}
