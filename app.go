package app

import (
  "net/http"
  "github.com/go-martini/martini"
  "github.com/martini-contrib/render"
// fb "github.com/huandu/facebook"
)

func init() {
	m := martini.Classic()
	m.Use(render.Renderer(render.Options{
		Extensions: []string{".tmpl", ".html"}, // Specify extensions to load for templates.

	}))

	m.Get("/", func(r render.Render) {
		r.HTML(200, "hello", "jeremy")
	})

	http.Handle("/", m)
}
