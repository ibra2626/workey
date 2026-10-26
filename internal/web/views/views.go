// Package views holds the templ components for the server-rendered UI.
package views

import (
	"github.com/a-h/templ"

	assets "github.com/ibra2626/workey/web"
)

// asset returns a /static path with the build version query for cache busting.
func asset(path string) string {
	return path + "?v=" + assets.Version
}

// navItem is one primary navigation destination, shown in both the desktop
// rail and the mobile bottom tab bar.
type navItem struct {
	Href  string
	Label string
	Icon  string // sprite symbol id, e.g. "i-today"
	Key   string // matches the active key passed to Layout
}

// navItems are the three primary sections. Order is fixed.
var navItems = []navItem{
	{Href: "/", Label: "Bugün", Icon: "i-today", Key: "bugun"},
	{Href: "/hafta", Label: "Hafta", Icon: "i-week", Key: "hafta"},
	{Href: "/ilerleme", Label: "İlerleme", Icon: "i-prog", Key: "ilerleme"},
}

// hxHeaders attaches the CSRF token to htmx requests via the hx-headers
// attribute on <body>, so every non-GET htmx request carries X-CSRF-Token. It
// renders nothing when csrf is empty (e.g. an unauthenticated render).
func hxHeaders(csrf string) templ.Attributes {
	if csrf == "" {
		return nil
	}
	return templ.Attributes{"hx-headers": `{"X-CSRF-Token":"` + csrf + `"}`}
}

// ariaCurrent marks the active link with aria-current="page" and nothing on the
// others; the visual active state is driven from that attribute in the CSS.
func ariaCurrent(active bool) templ.Attributes {
	if active {
		return templ.Attributes{"aria-current": "page"}
	}
	return nil
}
