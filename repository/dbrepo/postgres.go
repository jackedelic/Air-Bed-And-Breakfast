package dbrepo

import (
	"context"
	"time"

	"github.com/jackedelic/bookings/internal/models"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

func (m *postgresDBRepo) InsertReservation(res models.Reservation) error {
	// Makes sure the db connection does not stay open longer than the maxDBConnLifeTime configured to be 5 min.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	stmt := `insert into reservations (first_name, last_name, email, start_date,
					end_data, room_id, created_at, updated_at) 
					values ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := m.DB.ExecContext(ctx, stmt, res.FirstName,
		res.LastName,
		res.Email,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now())

	if err != nil {
		return err
	}
	return nil
}
