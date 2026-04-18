// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// DecodePullRequestResponse maps a GitHub GET /repos/{owner}/{repo}/pulls/{n}
// JSON body into [PullRequest]. Extra fields are ignored for forward
// compatibility.
func DecodePullRequestResponse(body []byte) (PullRequest, error) {
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	var raw struct {
		Number  int    `json:"number"`
		Title   string `json:"title"`
		State   string `json:"state"`
		HTMLURL string `json:"html_url"`
		Head    struct {
			SHA string `json:"sha"`
		} `json:"head"`
	}
	if err := dec.Decode(&raw); err != nil {
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
