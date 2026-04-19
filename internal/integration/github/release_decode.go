// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"encoding/json"
	"fmt"
)

// DecodeReleaseResponse maps POST /releases JSON into [Release].
func DecodeReleaseResponse(body []byte) (Release, error) {
	var raw struct {
		ID      int    `json:"id"`
		HTMLURL string `json:"html_url"`
		TagName string `json:"tag_name"`
		Name    string `json:"name"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return Release{}, fmt.Errorf("github: decode release: %w", err)
	}
	return Release{
		ID:      raw.ID,
		HTMLURL: raw.HTMLURL,
		TagName: raw.TagName,
		Name:    raw.Name,
	}, nil
}
