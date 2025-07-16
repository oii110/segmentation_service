package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"segment_service/internal/entities"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/exp/rand"
)

type SegmentService struct {
	db *pgxpool.Pool
}

func NewSegmentService(db *pgxpool.Pool) *SegmentService {
	return &SegmentService{db: db}
}

func (s *SegmentService) CreateSegment(w http.ResponseWriter, r *http.Request) {
	var segment entities.Segment
	if err := json.NewDecoder(r.Body).Decode(&segment); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var id int
	err := s.db.QueryRow(context.Background(),
		"INSERT INTO segments (name) VALUES ($1) RETURNING id",
		segment.Name).Scan(&id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	segment.ID = id
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(segment)
}

func (s *SegmentService) DeleteSegment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid segment ID", http.StatusBadRequest)
		return
	}

	tx, err := s.db.Begin(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), "DELETE FROM user_segments WHERE segment_id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(context.Background(), "DELETE FROM segments WHERE id = $1", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = tx.Commit(context.Background()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Segment %d deleted", id)
}

func (s *SegmentService) UpdateSegment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid segment ID", http.StatusBadRequest)
		return
	}

	var segment entities.Segment
	if err := json.NewDecoder(r.Body).Decode(&segment); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = s.db.Exec(context.Background(),
		"UPDATE segments SET name = $1 WHERE id = $2",
		segment.Name, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	segment.ID = id
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(segment)
}

func (s *SegmentService) AssignSegmentToUsers(w http.ResponseWriter, r *http.Request) {
	type AssignmentRequest struct {
		SegmentID int     `json:"segment_id"`
		Percent   float64 `json:"percent"`
	}

	var req AssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	rows, err := s.db.Query(context.Background(), "SELECT user_id FROM users")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var userIDs []int
	for rows.Next() {
		var userID int
		if err := rows.Scan(&userID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		userIDs = append(userIDs, userID)
	}

	rand.Seed(uint64(time.Now().UnixNano()))
	rand.Shuffle(len(userIDs), func(i, j int) {
		userIDs[i], userIDs[j] = userIDs[j], userIDs[i]
	})

	targetCount := int(float64(len(userIDs)) * req.Percent / 100)
	selectedUsers := userIDs[:min(targetCount, len(userIDs))]

	tx, err := s.db.Begin(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(context.Background())

	for _, userID := range selectedUsers {
		_, err = tx.Exec(context.Background(),
			"INSERT INTO user_segments (user_id, segment_id) VALUES ($1, $2) ON CONFLICT DO NOTHING",
			userID, req.SegmentID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err = tx.Commit(context.Background()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Segment %d assigned to %d users", req.SegmentID, len(selectedUsers))
}

func (s *SegmentService) GetUserSegments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	rows, err := s.db.Query(context.Background(),
		`SELECT s.id, s.name 
         FROM segments s 
         JOIN user_segments us ON s.id = us.segment_id 
         WHERE us.user_id = $1`, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var segments []entities.Segment
	for rows.Next() {
		var segment entities.Segment
		if err := rows.Scan(&segment.ID, &segment.Name); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		segments = append(segments, segment)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(segments)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
