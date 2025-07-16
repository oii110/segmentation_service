package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"segment_service/internal/config"
	"segment_service/internal/repository"
	"segment_service/internal/service"

	"github.com/gorilla/mux"
)

func main() {
	ctx := context.Background()
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	postgresRepo, err := repository.NewPostgresRepository(ctx, 3, cfg.Storage)
	if err != nil {
		log.Fatal(err)
	}
	defer postgresRepo.Conn().Close()

	segmentService := service.NewSegmentService(postgresRepo.Conn())

	r := mux.NewRouter()

	r.HandleFunc("/segments", segmentService.CreateSegment).Methods("POST")
	r.HandleFunc("/segments/{id}", segmentService.DeleteSegment).Methods("DELETE")
	r.HandleFunc("/segments/{id}", segmentService.UpdateSegment).Methods("PUT")
	r.HandleFunc("/segments/assign", segmentService.AssignSegmentToUsers).Methods("POST")
	r.HandleFunc("/users/{id}/segments", segmentService.GetUserSegments).Methods("GET")

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
