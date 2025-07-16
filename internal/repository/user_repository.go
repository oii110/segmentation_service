package repository

import (
	"context"
	"segment_service/internal/entities"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	AddUserSegments(db *pgxpool.Pool, user entities.User, newSegments []entities.Segment) error
}

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{pool: pool}
}

func (r *userRepository) AddUserSegments(db *pgxpool.Pool, user entities.User, newSegments []entities.Segment) error {
	var newSegmentsNames []string
	for _, segment := range newSegments {
		newSegmentsNames = append(newSegmentsNames, segment.Name)
	}
	query := `
        INSERT INTO users (user_id, segments)
        VALUES ($1, $2)
        ON CONFLICT (user_id) DO UPDATE
        SET segments = (
            SELECT array_agg(DISTINCT s)
            FROM unnest(users.segments || excluded.segments) AS s
        )
    `
	_, err := db.Exec(context.Background(), query, user.ID, newSegmentsNames)
	return err
}
