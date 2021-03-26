package dbrepo

import (
	"context"
	"time"

	"github.com/jackedelic/bookings/internal/models"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

func (m *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	stmt := `insert into reservations (first_name, last_name, email, start_date,
					end_date, room_id, created_at, updated_at) 
					values ($1, $2, $3, $4, $5, $6, $7, $8) returning id`

	var newID int
	err := m.DB.QueryRowContext(ctx, stmt, res.FirstName,
		res.LastName,
		res.Email,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now()).Scan(&newID)

	if err != nil {
		return newID, err
	}

	return newID, nil
}

// InsertRoomRestriction inserts a room into the database
func (m *postgresDBRepo) InsertRoomRestriction(rr models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	stmt := `insert into room_restrictions 
	(start_date, end_date, room_id, reservation_id, restriction_id, created_at, updated_at) 
	values ($1, $2, $3, $4, $5, $6, $7)`
	_, err := m.DB.ExecContext(ctx, stmt,
		rr.StartDate,
		rr.EndDate,
		rr.RoomID,
		rr.ReservationID,
		rr.RestrictionID,
		time.Now(),
		time.Now())

	if err != nil {
		return err
	}

	return nil
}

// SearchAvailabilityByDatesByRoomID search whether the given room is available.
// A room of a given date range is available if no existing RoomRestriction overlaps with it.
func (m *postgresDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
	defer cancel()
	stmt := `
		select 
			count(id)
		from 
			room_restrictions
		where
			room_id = $1 and
			$2 < end_date and start_date < $3;`
	var numRows int
	row := m.DB.QueryRowContext(ctx, stmt, roomID, start, end)
	err := row.Scan(&numRows)
	if err != nil {
		return false, err
	}

	if numRows == 0 {
		return true, nil
	} else {
		return false, nil
	}
}
