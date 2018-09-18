package main

import (
	"github.com/go-chi/chi"
	"net/http"
	"golang.org/x/net/context"
	"github.com/go-chi/render"
)

var (
	forlayosMap = map[string]*Forlayo{
		"1": {Id: "1", Name: "forlayo1", Number: 3, Price: 3.30},
		"2": {Id: "2", Name: "forlayo2", Number: 10, Price: 5.50},
	}
)

type Forlayo struct {
	Id     string
	Name   string
	Number int
	Price  float32
}

func listForlayos(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, forlayosMap)
}
func createForlayo(w http.ResponseWriter, r *http.Request) {
	f := &Forlayo{}
	if err := render.DecodeJSON(r.Body, f); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, r, err)
		return
	}
	forlayosMap[f.Id] = f
	render.JSON(w, r, f)
}

func getForlayo(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, r.Context().Value("forlayo").(*Forlayo))
}

func updateForlayo(w http.ResponseWriter, r *http.Request) {
	f := r.Context().Value("forlayo").(*Forlayo)
	forlayosMap[f.Id] = f
	render.JSON(w, r, f)
}

func deleteForlayo(w http.ResponseWriter, r *http.Request) {
	f := r.Context().Value("forlayo").(*Forlayo)
	delete(forlayosMap, f.Id)
	w.WriteHeader(http.StatusAccepted)
}

func ForlayoCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var forlayo *Forlayo
		var found bool

		if forlayoID := chi.URLParam(r, "forlayoID"); forlayoID != "" {
			forlayo, found = forlayosMap[forlayoID]
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if !found {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		ctx := context.WithValue(r.Context(), "forlayo", forlayo)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func main() {
	r := chi.NewRouter()

	r.Route("/forlayos", func(r chi.Router) {
		r.Get("/", listForlayos)
		r.Post("/", createForlayo)
		r.Route("/{forlayoID}", func(r chi.Router) {
			r.Use(ForlayoCtx)
			r.Get("/", getForlayo)       // GET /forlayos/123
			r.Put("/", updateForlayo)    // PUT /forlayos/123
			r.Delete("/", deleteForlayo) // DELETE /forlayos/123
		})
	})
	println("Listen on 8080")
	http.ListenAndServe(":8080", r)
}
