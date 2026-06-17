package development

import (
	"regexp"
	"strings"
	"unicode"
)

var cavemanHeaderReplacements = map[string]string{
	"ENTITY BRIEFINGS":        "BRIEF",
	"Entity Briefings:":       "BRIEF:",
	"CODE CONTEXT":            "CODE",
	"REPOSITORY CONTEXT":      "REPO",
	"REPOSITORY-CODE LINKS":   "LINKS",
	"Repository-Code Links:":  "LINKS:",
	"Evidence:":               "EV:",
	"Source Services:":        "SRC:",
	"Suggested Workflow:":     "FLOW:",
	"Suggested Steps:":        "STEPS:",
	"Related Entities:":       "REL:",
	"Answer Type:":            "TYPE:",
	"Question:":               "Q:",
	"Topic:":                    "TOPIC:",
	"Task:":                     "TASK:",
	"Summary:":                  "SUM:",
	"Impacted Areas:":           "IMPACT:",
	"Relevant Decisions:":       "DEC:",
	"Relevant Facts:":           "FACTS:",
	"Recommended Reviewers:":    "REV:",
	"Required Expertise:":       "EXP:",
	"Affected Areas:":           "AFF:",
	"Related Knowledge:":        "KNOW:",
	"Call Graph:":               "CALL:",
	"Dependencies:":             "DEP:",
	"Orientation:":              "ORIENT:",
	"Assignment Plan:":          "PLAN:",
	"Key Decisions:":            "KEY:",
	"Impacted Decisions:":       "IDEC:",
	"Impacted Facts:":           "IFACT:",
	"Impacted Events:":          "IEVT:",
	"Code Dependencies:":        "CDEP:",
	"Dependent Areas:":          "DEPND:",
	"Starting Points:":          "START:",
	"Owners:":                   "OWN:",
	"Reviewers:":                "REV:",
	"Modules:":                  "MOD:",
	"Files:":                    "FILES:",
	"Packages:":                 "PKG:",
	"Structs:":                  "S:",
	"Interfaces:":               "IF:",
	"Type Aliases:":             "T:",
	"Functions:":                "FN:",
	"Methods:":                  "M:",
	"Endpoints:":                "EP:",
	"Defined in:":               "DEF:",
	"Signature:":                "SIG:",
	"Fields:":                   "FLD:",
	"Called by":                 "BY",
	"Calls/uses":                "USES",
	"Related decisions":         "RDEC",
	"Layer:":                    "L:",
	"Role:":                     "R:",
}

var rxEvidenceSource = regexp.MustCompile(`^·\s+source:\s+(\S+)`)
var rxEvidenceType = regexp.MustCompile(`^type:\s+(\S+)`)

// ToCaveman compresses prose DE output for token-efficient agent consumption.
func ToCaveman(text string) string {
	if text == "" {
		return ""
	}
	out := text
	for from, to := range cavemanHeaderReplacements {
		out = strings.ReplaceAll(out, from, to)
	}

	lines := strings.Split(out, "\n")
	compact := make([]string, 0, len(lines))
	evidenceKinds := make(map[string]int)
	inEvidence := false
	pendingSource := ""
	inRelated := false
	relatedCount := 0
	relatedCap := 8
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			inEvidence = false
			continue
		}
		trimmed = collapseSpaces(trimmed)
		if strings.HasPrefix(trimmed, "- ") {
			trimmed = "· " + strings.TrimPrefix(trimmed, "- ")
		}
		if strings.HasPrefix(trimmed, "REL:") {
			inRelated = true
			compact = append(compact, trimmed)
			continue
		}
		if strings.HasPrefix(trimmed, "EV:") {
			inRelated = false
			inEvidence = true
			continue
		}
		if strings.HasPrefix(trimmed, "SRC:") {
			inEvidence = false
			inRelated = false
			pendingSource = ""
		}
		if inEvidence {
			if m := rxEvidenceSource.FindStringSubmatch(trimmed); len(m) == 2 {
				pendingSource = m[1]
				continue
			}
			if pendingSource != "" {
				if m := rxEvidenceType.FindStringSubmatch(trimmed); len(m) == 2 {
					evidenceKinds[pendingSource+"/"+m[1]]++
					pendingSource = ""
					continue
				}
			}
			if strings.HasPrefix(trimmed, "· source:") || strings.HasPrefix(trimmed, "- source:") {
				continue
			}
		}
		if inRelated && strings.HasPrefix(trimmed, "· ") {
			relatedCount++
			if relatedCount > relatedCap {
				continue
			}
		}
		trimmed = strings.ReplaceAll(trimmed, "[FILE]", "[F]")
		trimmed = strings.ReplaceAll(trimmed, "[FUNCTION]", "[fn]")
		trimmed = strings.ReplaceAll(trimmed, "[METHOD]", "[m]")
		trimmed = strings.ReplaceAll(trimmed, "[STRUCT]", "[S]")
		trimmed = strings.ReplaceAll(trimmed, "[INTERFACE]", "[I]")
		trimmed = strings.ReplaceAll(trimmed, "[DECISION]", "[D]")
		trimmed = strings.ReplaceAll(trimmed, "[FACT]", "[f]")
		trimmed = strings.ReplaceAll(trimmed, "[EVENT]", "[e]")
		trimmed = strings.ReplaceAll(trimmed, "[CONTRIBUTOR]", "[c]")
		trimmed = strings.ReplaceAll(trimmed, "internal/", "i/")
		trimmed = strings.ReplaceAll(trimmed, "github.com/reponerve/reponerve/", "rn/")
		trimmed = strings.ReplaceAll(trimmed, "repository_", "r_")
		trimmed = strings.ReplaceAll(trimmed, "code_intelligence", "code")
		trimmed = strings.ReplaceAll(trimmed, "repository_search", "search")
		compact = append(compact, trimmed)
	}
	if len(evidenceKinds) > 0 {
		parts := make([]string, 0, len(evidenceKinds))
		for k, n := range evidenceKinds {
			parts = append(parts, k+"×"+itoa(n))
		}
		compact = append(compact, "EV:"+strings.Join(parts, ","))
	}
	if relatedCount > relatedCap {
		compact = append(compact, "REL:+"+itoa(relatedCount-relatedCap)+" more")
	}
	return strings.Join(compact, "\n")
}

func itoa(n int) string {
	if n <= 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

// TruncateToTokenBudget limits text to an approximate token budget (4 chars ≈ 1 token).
func TruncateToTokenBudget(text string, budget int) string {
	if budget <= 0 || text == "" {
		return text
	}
	maxChars := budget * 4
	if len(text) <= maxChars {
		return text
	}
	if maxChars <= 3 {
		return text[:maxChars]
	}
	return text[:maxChars-3] + "..."
}

func collapseSpaces(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	prevSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !prevSpace {
				b.WriteByte(' ')
				prevSpace = true
			}
			continue
		}
		b.WriteRune(r)
		prevSpace = false
	}
	return strings.TrimSpace(b.String())
}
