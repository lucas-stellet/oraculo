// Package claudekit provides the embedded skill files for Oraculo.
package claudekit

import "embed"

//go:embed skills
var SkillsFS embed.FS
