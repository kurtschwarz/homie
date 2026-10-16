// Package assets embeds the built display and manager SPAs into the
// binary. go:embed paths are relative to this file's directory, which is
// why this package lives at the module root next to web/ rather than
// alongside main.go.
package assets

import "embed"

//go:embed web/display
var DisplayFS embed.FS

//go:embed web/manager
var ManagerFS embed.FS
