//go:build embedui

package webui

import (
	"embed"
	"io/fs"
)

// dist is filled with the built frontend (web/dist) before a combined build;
// see the Dockerfile's combined target.
//
//go:embed all:dist
var dist embed.FS

func files() fs.FS {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}
