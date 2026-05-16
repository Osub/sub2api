package service

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var (
	responseAdQqPattern = regexp.MustCompile(`(?i)(公益\s*Q+群|QQ群|Q群|群号)[：:\s]*[1-9][0-9]{5,11}`)
	responseAdLineHints = []string{
		"服务暂停",
		"公益qq群",
		"qq群通知",
		"群号",
	}
)

func sanitizeModelAdResponseBody(body []byte) []byte {
	if len(body) == 0 || !gjson.ValidBytes(body) {
		return body
	}

	updated := body
	changed := false

	applyText := func(path string) {
		original := gjson.GetBytes(updated, path)
		if !original.Exists() || original.Type != gjson.String {
			return
		}
		cleaned := sanitizeModelAdText(original.String())
		if cleaned == original.String() {
			return
		}
		if next, err := sjson.SetBytes(updated, path, cleaned); err == nil {
			updated = next
			changed = true
		}
	}

	choices := gjson.GetBytes(updated, "choices").Array()
	for i := range choices {
		base := "choices." + strconvInt(i)
		applyText(base + ".message.content")
		applyText(base + ".delta.content")
	}

	outputs := gjson.GetBytes(updated, "output").Array()
	for i := range outputs {
		outputBase := "output." + strconvInt(i)
		contents := outputs[i].Get("content").Array()
		for j := range contents {
			contentBase := outputBase + ".content." + strconvInt(j)
			applyText(contentBase + ".text")
		}
	}

	if !changed {
		return body
	}
	return updated
}

func sanitizeModelAdText(text string) string {
	if text == "" {
		return text
	}

	normalized := strings.ToLower(text)
	hasHint := false
	for _, hint := range responseAdLineHints {
		if strings.Contains(normalized, strings.ToLower(hint)) {
			hasHint = true
			break
		}
	}
	if !hasHint && !responseAdQqPattern.MatchString(text) {
		return text
	}

	lines := strings.Split(text, "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		if shouldDropAdLine(line) {
			continue
		}
		kept = append(kept, line)
	}

	cleaned := strings.Join(kept, "\n")
	cleaned = collapseExtraBlankLines(cleaned)
	return strings.TrimSpace(cleaned)
}

func shouldDropAdLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}
	if responseAdQqPattern.MatchString(trimmed) {
		return true
	}

	lower := strings.ToLower(trimmed)
	hasPause := strings.Contains(lower, "服务暂停")
	hasQqHint := strings.Contains(lower, "qq群") || strings.Contains(lower, "q群") || strings.Contains(lower, "群号") || strings.Contains(lower, "公益qq群")
	return hasPause && hasQqHint
}

func collapseExtraBlankLines(text string) string {
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))
	prevBlank := false
	for _, line := range lines {
		blank := strings.TrimSpace(line) == ""
		if blank && prevBlank {
			continue
		}
		out = append(out, line)
		prevBlank = blank
	}
	return strings.Join(out, "\n")
}

func strconvInt(i int) string {
	return strconv.Itoa(i)
}
