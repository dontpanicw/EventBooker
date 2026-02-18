package http

import (
	"context"
	"fmt"
	"github.com/dontpanicw/EventBooker/internal/port"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

type Server struct {
	handler *Handler
	server  *http.Server
}

func NewServer(port string, usecases port.Usecases) *Server {
	handler := NewHandler(usecases)

	router := mux.NewRouter()

	// CORS middleware
	router.Use(corsMiddleware)

	// Статические файлы
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static"))))
	
	// HTML страницы
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/user.html")
	})
	router.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/admin.html")
	})

	// API маршруты
	router.HandleFunc("/api/events", handler.GetAllEvents).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/events", handler.CreateEvent).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/events/{id}/book", handler.BookEvent).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/events/{id}", handler.GetEvent).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/bookings/{id}", handler.GetBooking).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/bookings/{id}/confirm", handler.ConfirmBooking).Methods("POST", "OPTIONS")

	server := &http.Server{
		Addr:         port,
		Handler:      router,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  300 * time.Second,
	}

	return &Server{
		handler: handler,
		server:  server,
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) Start() error {
	log.Printf("Starting HTTP server on %s\n", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) error {
	fmt.Println("Shutting down HTTP server...")
	return s.server.Shutdown(ctx)
}
