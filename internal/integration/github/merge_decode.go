// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"encoding/json"
	"fmt"
)

// DecodeMergeResponse maps PUT merge API JSON into [MergePullRequestResult].
func DecodeMergeResponse(body []byte) (MergePullRequestResult, error) {
	var raw struct {
		Merged  bool   `json:"merged"`
		Message string `json:"message"`
		SHA     string `json:"sha"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return MergePullRequestResult{}, fmt.Errorf("github: decode merge response: %w", err)
	}
	return MergePullRequestResult{
		Merged:  raw.Merged,
		Message: raw.Message,
		SHA:     raw.SHA,
	}, nil
}
