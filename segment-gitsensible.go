package main

import (
	"fmt"
	"strconv"
	"strings"
)

type repoStatsSensible struct {
	ahead      int
	behind     int
	untracked  int
	notStaged  int
	staged     int
	conflicted int
	stashed    int
}

func (r repoStatsSensible) dirtySensible() bool {
	return r.notStaged+r.staged+r.conflicted > 0
}

func addRepoStatsSegmentSensible(p *powerline, nChanges int, symbol string, foreground uint8, background uint8, printnChanges bool) {
	if nChanges > 0 {
		content := fmt.Sprintf("%s", symbol)
		if printnChanges {
			content = fmt.Sprintf("%d", nChanges) + content
		}
		p.appendSegment("git-status", segment{
			content:        content,
			foreground:     foreground,
			background:     background,
			hideSeparators: true,
		})
	}
}

func (r repoStatsSensible) addToPowerlineSensible(p *powerline, background uint8) {
	addRepoStatsSegmentSensible(p, r.ahead, p.symbolTemplates.RepoAhead, p.theme.GitAheadFg, p.theme.GitAheadBg, true)
	addRepoStatsSegmentSensible(p, r.behind, p.symbolTemplates.RepoBehind, p.theme.GitBehindFg, p.theme.GitBehindBg, true)
	addRepoStatsSegmentSensible(p, r.staged, p.symbolTemplates.RepoStaged, p.theme.GitStagedFg, p.theme.GitStagedBg, true)
	addRepoStatsSegmentSensible(p, r.notStaged, p.symbolTemplates.RepoNotStaged, p.theme.GitNotStagedFg, p.theme.GitNotStagedBg, true)
	addRepoStatsSegmentSensible(p, r.conflicted, p.symbolTemplates.RepoConflicted, p.theme.GitConflictedFg, p.theme.GitConflictedBg, true)
	addRepoStatsSegmentSensible(p, r.stashed, p.symbolTemplates.RepoStashed, p.theme.GitStashedFg, p.theme.GitStashedBg, false)
	addRepoStatsSegmentSensible(p, r.untracked, p.symbolTemplates.RepoUntracked, p.theme.GitUntrackedFg, p.theme.GitUntrackedBg, false)
	p.appendSegment("git-status", segment{
		content:    "",
		foreground: background,
		background: background,
	})
}

func parseGitStatsSensible(status []string) repoStatsSensible {
	stats := repoStatsSensible{}
	if len(status) > 1 {
		for _, line := range status[1:] {
			if len(line) > 2 {
				code := line[:2]
				switch code {
				case "??":
					stats.untracked++
				case "DD", "AU", "UD", "UA", "DU", "AA", "UU":
					stats.conflicted++
				default:
					if code[0] != ' ' {
						stats.staged++
					}

					if code[1] != ' ' {
						stats.notStaged++
					}
				}
			}
		}
	}
	return stats
}

func segmentGitSensible(p *powerline) {
	if len(p.ignoreRepos) > 0 {
		out, err := runGitCommand("git", "rev-parse", "--show-toplevel")
		if err != nil {
			return
		}
		out = strings.TrimSpace(out)
		if p.ignoreRepos[out] {
			return
		}
	}

	out, err := runGitCommand("git", "status", "--porcelain", "-b", "--ignore-submodules")
	if err != nil {
		return
	}

	status := strings.Split(out, "\n")
	stats := parseGitStatsSensible(status)
	branchInfo := parseGitBranchInfo(status)
	var branch string

	if branchInfo["local"] != "" {
		ahead, _ := strconv.ParseInt(branchInfo["ahead"], 10, 32)
		stats.ahead = int(ahead)

		behind, _ := strconv.ParseInt(branchInfo["behind"], 10, 32)
		stats.behind = int(behind)

		branch = branchInfo["local"]
	} else {
		branch = getGitDetachedBranch(p)
	}

	var foreground, background uint8
	if stats.dirtySensible() {
		foreground = p.theme.RepoDirtyFg
		background = p.theme.RepoDirtyBg
	} else {
		foreground = p.theme.RepoCleanFg
		background = p.theme.RepoCleanBg
	}

	out, err = runGitCommand("git", "stash", "list")
	if err != nil {
		return
	}
	if len(out) > 0 {
		stats.stashed = len(strings.Split(out, "\n")) - 1
	}

	p.appendSegment("git-branch", segment{
		content:    branch,
		foreground: foreground,
		background: background,
	})
	stats.addToPowerlineSensible(p, background)
}
