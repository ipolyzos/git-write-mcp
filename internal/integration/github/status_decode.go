// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package github

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// DecodePullRequestForStatus maps GET /repos/{owner}/{repo}/pulls/{n} JSON into [PullRequestStatus]
// (checks and reviews are filled later by the client).
func DecodePullRequestForStatus(body []byte) (PullRequestStatus, error) {
	var raw struct {
		Number         int    `json:"number"`
		Title          string `json:"title"`
		State          string `json:"state"`
		Merged         bool   `json:"merged"`
		Draft          bool   `json:"draft"`
		Mergeable      *bool  `json:"mergeable"`
		MergeableState string `json:"mergeable_state"`
		HTMLURL        string `json:"html_url"`
		Head           struct {
			Ref string `json:"ref"`
			SHA string `json:"sha"`
		} `json:"head"`
		Base struct {
			Ref string `json:"ref"`
		} `json:"base"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return PullRequestStatus{}, fmt.Errorf("github: decode pull request: %w", err)
	}
	return PullRequestStatus{
		Number:         raw.Number,
		Title:          raw.Title,
		State:          raw.State,
		Merged:         raw.Merged,
		Draft:          raw.Draft,
		Mergeable:      raw.Mergeable,
		MergeableState: strings.TrimSpace(raw.MergeableState),
		HTMLURL:        raw.HTMLURL,
		HeadRef:        strings.TrimSpace(raw.Head.Ref),
		BaseRef:        strings.TrimSpace(raw.Base.Ref),
		HeadSHA:        strings.TrimSpace(raw.Head.SHA),
		ChecksSummary:  ChecksSummary{},
		ReviewsSummary: ReviewsSummary{},
	}, nil
}

// DecodeCombinedStatus maps GET /repos/{owner}/{repo}/commits/{sha}/status JSON into [ChecksSummary].
func DecodeCombinedStatus(body []byte) (ChecksSummary, error) {
	var raw struct {
		State      string `json:"state"`
		TotalCount int    `json:"total_count"`
		Statuses   []struct {
			State   string `json:"state"`
			Context string `json:"context"`
		} `json:"statuses"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return ChecksSummary{}, fmt.Errorf("github: decode combined status: %w", err)
	}
	ctxs := make([]StatusContext, 0, len(raw.Statuses))
	for _, s := range raw.Statuses {
		ctxs = append(ctxs, StatusContext{
			Context: strings.TrimSpace(s.Context),
			State:   strings.TrimSpace(s.State),
		})
	}
	sort.Slice(ctxs, func(i, j int) bool {
		if ctxs[i].Context != ctxs[j].Context {
			return ctxs[i].Context < ctxs[j].Context
		}
		return ctxs[i].State < ctxs[j].State
	})
	return ChecksSummary{
		State:      strings.TrimSpace(raw.State),
		TotalCount: raw.TotalCount,
		Contexts:   ctxs,
	}, nil
}

// DecodeReviewsSummary maps GET /repos/{owner}/{repo}/pulls/{n}/reviews JSON into [ReviewsSummary].
func DecodeReviewsSummary(body []byte) (ReviewsSummary, error) {
	var reviews []struct {
		State string `json:"state"`
	}
	if err := json.Unmarshal(body, &reviews); err != nil {
		return ReviewsSummary{}, fmt.Errorf("github: decode reviews: %w", err)
	}
	var out ReviewsSummary
	for _, r := range reviews {
		switch strings.ToUpper(strings.TrimSpace(r.State)) {
		case "APPROVED":
			out.Approved++
		case "CHANGES_REQUESTED":
			out.ChangesRequested++
		case "COMMENTED":
			out.Commented++
		case "PENDING":
			out.Pending++
		case "DISMISSED":
			out.Dismissed++
		}
	}
	return out, nil
}
