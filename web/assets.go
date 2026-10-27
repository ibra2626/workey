// Package web embeds the built static assets (CSS, htmx, fonts) so the server
// ships as a single binary. The Tailwind source lives in web/styles and is not
// embedded; only the compiled web/static tree is.
package web

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io/fs"
	"net/http"
)

//go:embed static
var staticFS embed.FS

// Version hashes all embedded assets so script-only changes also invalidate
// the immutable browser cache. WalkDir visits paths in lexical order.
var Version = computeVersion()

func computeVersion() string {
	h := sha256.New()
	err := fs.WalkDir(staticFS, "static", func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		b, err := staticFS.ReadFile(path)
		if err != nil {
			return err
		}
		h.Write([]byte(path + "\x00"))
		h.Write(b)
		return nil
	})
	if err != nil {
		return "dev"
	}
	return hex.EncodeToString(h.Sum(nil))[:12]
}

// StaticHandler serves the embedded assets. Register it under "GET /static/".
// Assets are immutable for a year; the ?v= query busts the cache on rebuild.
func StaticHandler() http.Handler {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic("web: embedded static subtree missing: " + err.Error())
	}
	// StripPrefix without the trailing slash leaves a rooted path ("/app.css")
	// which http.FileServerFS expects.
	fileServer := http.StripPrefix("/static", http.FileServerFS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		fileServer.ServeHTTP(w, r)
	})
}
