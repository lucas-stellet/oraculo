package spa

import "strings"

// WithPlaceholders replaces dynamic route segments with "__placeholder__".
// Known dynamic positions: epics/{id} at index 1, approvals/{approvalId} at index 3,
// and stories/{storyId} at index 3 when under /stories/.
func WithPlaceholders(fsPath string) string {
	segs := strings.Split(strings.Trim(fsPath, "/"), "/")
	if len(segs) < 2 || segs[0] != "epics" {
		return fsPath
	}
	result := make([]string, len(segs))
	copy(result, segs)
	changed := false
	if result[1] != "__placeholder__" {
		result[1] = "__placeholder__"
		changed = true
	}
	if len(result) >= 4 && result[2] == "approvals" && result[3] != "__placeholder__" {
		result[3] = "__placeholder__"
		changed = true
	}
	if len(result) >= 4 && result[2] == "stories" && result[3] != "__placeholder__" {
		result[3] = "__placeholder__"
		changed = true
	}
	if !changed {
		return fsPath
	}
	return strings.Join(result, "/")
}

// Shell maps a URL path to the most appropriate pre-rendered shell file.
func Shell(urlPath string, isRSC bool) string {
	ext := ".html"
	if isRSC {
		ext = ".txt"
	}
	segs := splitPath(urlPath)
	n := len(segs)

	// /epics/{id}/approvals/{approvalId}/review
	if n >= 5 && segs[0] == "epics" && segs[2] == "approvals" && segs[4] == "review" {
		return "/epics/__placeholder__/approvals/__placeholder__/review" + ext
	}
	// /epics/{id}/approvals
	if n >= 3 && segs[0] == "epics" && segs[2] == "approvals" {
		return "/epics/__placeholder__/approvals" + ext
	}
	// /epics/{id}/stories/{storyId}
	if n >= 4 && segs[0] == "epics" && segs[2] == "stories" {
		return "/epics/__placeholder__/stories/__placeholder__" + ext
	}
	// /epics/{id}
	if n >= 2 && segs[0] == "epics" {
		return "/epics/__placeholder__" + ext
	}
	return "/"
}

func splitPath(p string) []string {
	var segs []string
	for _, s := range strings.Split(strings.Trim(p, "/"), "/") {
		if s != "" {
			segs = append(segs, s)
		}
	}
	return segs
}
