// Copyright 2026 Ioannis Polyzos
//
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// BuildDryRunPreview returns a deterministic, machine-friendly preview of the mutation
// described by req. Call only after [MutationRequest.Validate] has succeeded so that
// GitArgs and Target reflect resolved tool inputs (e.g. stage_files paths).
func BuildDryRunPreview(req *MutationRequest) DryRunPreview {
	if req == nil {
		return DryRunPreview{}
	}
	wd := filepath.Clean(req.WorkDir)
	argv := append([]string(nil), req.GitArgs...)
	cmdLine := plannedPreviewCommandLine(req, wd, argv)
	return DryRunPreview{
		Action:        req.Action,
		Targets:       previewTargets(req),
		SecondaryRefs: previewSecondaryRefs(req),
		GitArgv:       argv,
		GitCommands:   []string{cmdLine},
		WorkDir:       wd,
		EffectSummary: previewEffectSummary(req),
	}
}

func previewTargets(req *MutationRequest) []string {
	switch req.Action {
	case "stage_files", "unstage_files":
		if req.Target == "" {
			return nil
		}
		parts := strings.Split(req.Target, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			if s := strings.TrimSpace(p); s != "" {
				out = append(out, s)
			}
		}
		return out
	case "tag_revision":
		if len(req.GitArgs) >= 3 && req.GitArgs[0] == "tag" && req.GitArgs[1] == "-a" {
			return []string{req.GitArgs[2]}
		}
		s := strings.TrimSpace(req.Target)
		if s == "" {
			return nil
		}
		return []string{s}
	case "commit_changes":
		s := strings.TrimSpace(req.Target)
		if s == "" {
			return nil
		}
		return []string{s}
	case "create_branch":
		s := strings.TrimSpace(req.Target)
		if s == "" {
			return nil
		}
		return []string{s}
	case "push_branch":
		s := strings.TrimSpace(req.Target)
		if s == "" {
			return nil
		}
		return []string{s}
	case "create_pull_request":
		s := strings.TrimSpace(req.PRSourceBranch)
		if s == "" {
			return nil
		}
		return []string{s}
	case "merge_pull_request":
		if req.PRMergeNumber > 0 {
			return []string{strconv.Itoa(req.PRMergeNumber)}
		}
		return nil
	case "create_release_from_main":
		s := strings.TrimSpace(req.ReleaseTagName)
		if s == "" {
			return nil
		}
		return []string{s}
	default:
		s := strings.TrimSpace(req.Target)
		if s == "" {
			return nil
		}
		return []string{s}
	}
}

func previewSecondaryRefs(req *MutationRequest) []string {
	switch req.Action {
	case "create_branch":
		if strings.TrimSpace(req.FromRef) != "" {
			return []string{req.FromRef}
		}
	case "tag_revision":
		if strings.TrimSpace(req.TargetRef) != "" {
			return []string{req.TargetRef}
		}
	case "push_branch":
		if strings.TrimSpace(req.PushRemote) != "" {
			return []string{req.PushRemote}
		}
	case "create_pull_request":
		if t := strings.TrimSpace(req.PRTargetBranch); t != "" {
			return []string{t}
		}
	case "merge_pull_request":
		if t := strings.TrimSpace(req.PRTargetBranch); t != "" {
			return []string{t}
		}
	case "create_release_from_main":
		if t := strings.TrimSpace(req.ReleaseTargetRef); t != "" {
			return []string{t}
		}
	}
	return nil
}

func previewEffectSummary(req *MutationRequest) string {
	switch req.Action {
	case "create_branch":
		name := strings.TrimSpace(req.Target)
		if req.FromRef != "" {
			return fmt.Sprintf(`Would create branch %q starting from ref %q.`, name, req.FromRef)
		}
		return fmt.Sprintf(`Would create branch %q from the current HEAD commit.`, name)
	case "stage_files":
		targets := previewTargets(req)
		if len(targets) == 0 {
			return `Would stage paths for the next commit.`
		}
		quoted := make([]string, len(targets))
		for i, t := range targets {
			quoted[i] = strconv.Quote(t)
		}
		return fmt.Sprintf(`Would stage %d path(s) for the next commit: %s.`, len(targets), strings.Join(quoted, ", "))
	case "unstage_files":
		targets := previewTargets(req)
		if len(targets) == 0 {
			return `Would remove paths from the Git index (working tree unchanged).`
		}
		quoted := make([]string, len(targets))
		for i, t := range targets {
			quoted[i] = strconv.Quote(t)
		}
		return fmt.Sprintf(`Would unstage %d path(s) from the index (working tree unchanged): %s.`, len(targets), strings.Join(quoted, ", "))
	case "commit_changes":
		return `Would create a new commit with the staged changes using the provided message.`
	case "tag_revision":
		tag := ""
		if len(req.GitArgs) >= 3 && req.GitArgs[0] == "tag" && req.GitArgs[1] == "-a" {
			tag = req.GitArgs[2]
		} else {
			tag = strings.TrimSpace(req.Target)
		}
		if req.TargetRef != "" {
			return fmt.Sprintf(`Would create annotated tag %q at ref %q.`, tag, req.TargetRef)
		}
		return fmt.Sprintf(`Would create annotated tag %q at HEAD.`, tag)
	case "push_branch":
		b := strings.TrimSpace(req.Target)
		r := strings.TrimSpace(req.PushRemote)
		return fmt.Sprintf(`Would push branch %q to remote %q (no force, no refspec).`, b, r)
	case "create_pull_request":
		ti := strings.TrimSpace(req.PRTitle)
		return fmt.Sprintf(`Would open pull request %q from %q into %q (draft=%v).`,
			ti, strings.TrimSpace(req.PRSourceBranch), strings.TrimSpace(req.PRTargetBranch), req.PRDraft)
	case "merge_pull_request":
		return fmt.Sprintf(`Would merge pull request #%d into %q using merge_method=%q after policy gates (checks, reviews, mergeability).`,
			req.PRMergeNumber, strings.TrimSpace(req.PRTargetBranch), strings.TrimSpace(req.PRMergeMethod))
	default:
		return `Would run git with the planned arguments (dry-run; no repository changes).`
	}
}

func plannedPreviewCommandLine(req *MutationRequest, wd string, argv []string) string {
	if req == nil {
		return plannedGitCommandLine(wd, argv)
	}
	o := strings.TrimSpace(req.GitHubOwner)
	r := strings.TrimSpace(req.GitHubRepo)
	switch req.Action {
	case "create_pull_request":
		return fmt.Sprintf("GitHub API POST /repos/%s/%s/pulls head=%q base=%q draft=%v",
			o, r, strings.TrimSpace(req.PRSourceBranch), strings.TrimSpace(req.PRTargetBranch), req.PRDraft)
	case "merge_pull_request":
		return fmt.Sprintf("GitHub API PUT /repos/%s/%s/pulls/%d/merge merge_method=%s",
			o, r, req.PRMergeNumber, strings.TrimSpace(req.PRMergeMethod))
	case "create_release_from_main":
		return fmt.Sprintf("GitHub API POST /repos/%s/%s/releases tag=%q name=%q target_commitish=%s",
			o, r, strings.TrimSpace(req.ReleaseTagName), strings.TrimSpace(req.ReleaseGitHubTitle), strings.TrimSpace(req.ReleaseTargetRef))
	default:
		return plannedGitCommandLine(wd, argv)
	}
}

func plannedGitCommandLine(workDir string, argv []string) string {
	wd := filepath.Clean(workDir)
	var b strings.Builder
	b.WriteString("git -C ")
	b.WriteString(strconv.Quote(wd))
	for _, a := range argv {
		b.WriteString(" ")
		b.WriteString(strconv.Quote(a))
	}
	return b.String()
}
