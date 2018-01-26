package main

import (
	"log"
	"github.com/go-chi/chi"
	"net/http"
	"github.com/go-chi/render"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"forlayos/rpc"
	"net"
	"github.com/soheilhy/cmux"
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
	Number int32
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

type ForlayosImpl struct {
}

func (f ForlayosImpl) ListForlayos(empty *rpc.Empty, stream rpc.Forlayos_ListForlayosServer) error {
	log.Println("List Forlayos")
	for _, forlayo := range forlayosMap {
		forlayoRPC := &rpc.Forlayo{
			Id:     forlayo.Id,
			Name:   forlayo.Name,
			Number: forlayo.Number,
			Price:  forlayo.Price,
		}
		log.Println(forlayo)
		stream.Send(forlayoRPC)
	}
	return nil
}

func main() {
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	mux := cmux.New(lis)

	restListener := mux.Match(cmux.HTTP1HeaderField("content-type", "application/json"))
	grpcListener := mux.Match(cmux.Any())

	//GRPC Server
	grpcServer := grpc.NewServer()
	rpc.RegisterForlayosServer(grpcServer, ForlayosImpl{})

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
	chiServer := &http.Server{Handler: r}

	go grpcServer.Serve(grpcListener)
	go chiServer.Serve(restListener)

	log.Print("Starting server at port 8080")

	mux.Serve()
}
