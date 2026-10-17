//go:build !embedui

package webui

import "io/fs"

// Without the embedui build tag there's no UI to serve: the split api image,
// and local development, where Vite serves the UI.
func files() fs.FS { return nil }
