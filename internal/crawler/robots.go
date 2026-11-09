package crawler

import "strings"

// robotRules answers one host question: may we fetch this path? It is a
// deliberately small robots.txt subset — exactly what an honest single-agent
// crawler needs (AGENTS.md rule 8). A full parser needs a dependency, and
// adding one requires an ADR first (rule 4).
//
// Understood lines, case-insensitive, "#" comments trimmed:
//
//	User-Agent: <token>
//	Disallow: <prefix>   (empty prefix = disallow everything)
//	Allow: <prefix>
//
// Semantics (RFC 9309 as much as the subset allows):
//   - user-agent lines that come back-to-back belong to one group; a rule
//     line ends the agent block, so the next user-agent line starts a new
//     group;
//   - the group whose user-agent equals our token (case-insensitive) wins;
//     if none match, the "*" group is used; if even that is missing,
//     everything is allowed;
//   - inside the winning group the longest matching prefix wins; a tie
//     between Allow and Disallow favours Allow;
//   - no matching rule at all -> allowed.
type robotRules struct {
	agentToken string   // our token, lowercased, for tests/debug
	agents     []string // the group's user-agent tokens, lowercased
	disallow   []string
	allow      []string
}

// parseRobotRules parses robots.txt bytes for the given agent token, e.g.
// "jobcrawler".
func parseRobotRules(body []byte, agentToken string) robotRules {
	r := robotRules{agentToken: strings.ToLower(strings.TrimSpace(agentToken))}

	var (
		groups       []*robotRules
		current      = &robotRules{}
		lastWasAgent = false
	)
	flush := func() {
		if len(current.agents) > 0 {
			groups = append(groups, current)
		}
		current = &robotRules{}
	}
	appendAgent := func(tok string) {
		if !lastWasAgent {
			flush()
		}
		lastWasAgent = true
		current.agents = append(current.agents, strings.ToLower(tok))
	}

	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}
		key, value, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		switch key {
		case "user-agent":
			appendAgent(value)
		case "disallow":
			if lastWasAgent {
				lastWasAgent = false
			}
			if len(current.agents) > 0 {
				current.disallow = append(current.disallow, value)
			}
		case "allow":
			if lastWasAgent {
				lastWasAgent = false
			}
			if len(current.agents) > 0 {
				current.allow = append(current.allow, value)
			}
		}
	}
	flush()

	best := -1
	for gi, g := range groups {
		for _, a := range g.agents {
			// RFC 9309: the token must match ours exactly (case-insensitive);
			// the "*" wildcard group is the fallback below.
			if a != "" && a == r.agentToken {
				best = gi
				break
			}
		}
		if best >= 0 {
			break
		}
	}
	if best < 0 {
		best = indexOfStar(groups)
	}
	if best >= 0 {
		r.agents, r.disallow, r.allow = groups[best].agents, groups[best].disallow, groups[best].allow
	}
	return r
}

func indexOfStar(groups []*robotRules) int {
	for gi, g := range groups {
		for _, a := range g.agents {
			if a == "*" {
				return gi
			}
		}
	}
	return -1
}

// allowed reports whether path (like "/v1/boards/acme/jobs") may be fetched.
// A robots.txt with no rules for us — or none at all — means allowed.
func (r robotRules) allowed(path string) bool {
	if len(r.disallow) == 0 && len(r.allow) == 0 {
		return true
	}
	path = strings.ToLower(path)
	longest, verdict := -1, true
	for _, d := range r.disallow {
		if prefixCovers(d, path) && len(d) > longest {
			longest, verdict = len(d), false
		}
	}
	for _, a := range r.allow {
		if prefixCovers(a, path) && len(a) >= longest && len(a) > 0 {
			longest, verdict = len(a), true
		}
	}
	return verdict
}

// prefixCovers: the rule matches when it is a lowercased prefix of the path.
// A rule of exactly "" covers everything (legacy robots semantics:
// "Disallow: " with empty value = disallow all).
func prefixCovers(rule, path string) bool {
	if rule == "" {
		return true
	}
	return strings.HasPrefix(path, rule)
}
