package server

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := httprouter.New()

	// Wrap all routes with CORS middleware
	corsWrapper := s.corsMiddleware(r)

	r.GET("/", s.HelloWorldHandler)
	r.POST("/register", s.Register)
	r.POST("/login", s.Login)

	r.POST("/trainings/add", s.AddTraining)
	r.GET("/trainings/get_all/:username", s.GetUserTrainings)
	r.DELETE("/trainings/delete/:id", s.DeleteTraining)

	return corsWrapper
}

// CORS middleware
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Use "*" for all origins, or replace with specific origins
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Credentials", "false") // Set to "true" if credentials are needed
		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}
