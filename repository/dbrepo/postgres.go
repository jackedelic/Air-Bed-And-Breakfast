package dbrepo

import (
	"context"
	"errors"
	"time"

	"github.com/jackedelic/bookings/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (m *postgresDBRepo) AllUsers() bool {
	return true
}

// InsertReservation inserts a reservation into the database and returns the id for the reservation.
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
		m.App.ErrorLog.Println(err)
		return err
	}

	return nil
}

// SearchAvailabilityByDatesByRoomID search whether the given room is available.
// A room of a given date range is available if no existing RoomRestriction overlaps with it.
func (m *postgresDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
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

// SearchAvailableRoomsByDates returns a slice of available rooms if any for given date range.
func (m *postgresDBRepo) SearchAvailableRoomsByDates(start, end time.Time) ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
		select
			r.id, r.room_name
		from
			rooms r
		where
			r.id not in
		(select rr.room_id from room_restrictions rr where $1 < rr.end_date and rr.start_date <= $2)

	`
	rows, err := m.DB.QueryContext(ctx, stmt, start, end)
	if err != nil {
		return []models.Room{}, err
	}
	defer rows.Close()

	var rooms []models.Room

	for rows.Next() {
		room := models.Room{}
		err := rows.Scan(&room.ID, &room.RoomName)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, room)
	}

	if err := rows.Err(); err != nil {
		return rooms, err
	}

	return rooms, nil
}

// GetRoomById returns a models.Room by id
func (m *postgresDBRepo) GetRoomById(id int) (models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var room models.Room
	query := `select id, room_name, created_at, updated_at from rooms where id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(&room.ID, &room.RoomName, &room.CreatedAt, &room.UpdatedAt)
	if err != nil {
		return room, err
	}

	return room, nil
}

// GetAllRooms returns a slice of models.Room containing all existing rooms
func (m *postgresDBRepo) GetAllRooms() ([]models.Room, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var rooms []models.Room
	query := `select id, room_name, created_at, updated_at from rooms order by room_name`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return rooms, err
	}
	defer rows.Close()

	for rows.Next() {
		room := models.Room{}
		err = rows.Scan(&room.ID, &room.RoomName, &room.CreatedAt, &room.UpdatedAt)
		if err != nil {
			return rooms, err
		}
		rooms = append(rooms, room)
	}
	if err = rows.Err(); err != nil {
		return rooms, err
	}
	return rooms, nil
}

// GetUserById returns a user by id
func (m *postgresDBRepo) GetUserById(id int) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `select id, first_name, last_name, email, password,
		 access_level, created_at, updated_at
		from users where id = $1`

	row := m.DB.QueryRowContext(ctx, query, id)
	var u models.User
	err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.Password, &u.AccessLevel, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return u, err
	}

	return u, nil
}

// UpdateUserById updates the user by id
func (m *postgresDBRepo) UpdateUser(u models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update users set first_name = $1, last_name = $2, email = $3, access_level = $4
		updated_at = $5 where id = $6`

	_, err := m.DB.ExecContext(ctx, query, u.FirstName, u.LastName, u.Email, u.AccessLevel, time.Now(), u.ID)
	if err != nil {
		return err
	}
	return nil
}

// Authenticate authenticates a user by user-given email and password.
// Returns user_id from users table, hashed password and error
func (m *postgresDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var (
		userID         int
		hashedPassword string
	)

	query := `select id, password from users where email = $1`
	row := m.DB.QueryRowContext(ctx, query, email)
	err := row.Scan(&userID, &hashedPassword)
	if err != nil {
		return userID, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", err
	}

	return userID, hashedPassword, nil
}

// GetAllReservations returns a slice of all reservations
func (m *postgresDBRepo) GetAllReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
		select r.id, r.first_name, r.last_name, r.email, r.start_date, r.end_date, r.room_id,
		r.created_at, r.updated_at, r.processed,
		rm.id, rm.room_name
		from reservations r
		left join rooms rm on (r.room_id = rm.id)
		order by r.start_date asc
	`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.Reservation
		err := rows.Scan(&r.ID, &r.FirstName, &r.LastName, &r.Email,
			&r.StartDate, &r.EndDate, &r.RoomID, &r.CreatedAt, &r.UpdatedAt, &r.Processed, &r.Room.ID, &r.Room.RoomName)

		if err != nil {
			return reservations, err
		}
		reservations = append(reservations, r)
	}
	if err = rows.Err(); err != nil {
		return reservations, err
	}
	return reservations, nil
}

// GetAllNewReservations returns a slice of all the new reservations (processed = 0)
func (m *postgresDBRepo) GetAllNewReservations() ([]models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservations []models.Reservation

	query := `
		select r.id, r.first_name, r.last_name, r.email, r.start_date, r.end_date, r.room_id,
		r.created_at, r.updated_at, r.processed,
		rm.id, rm.room_name
		from reservations r
		left join rooms rm on (r.room_id = rm.id)
		where r.processed = 0
		order by r.start_date asc
	`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return reservations, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.Reservation
		err := rows.Scan(&r.ID, &r.FirstName, &r.LastName, &r.Email,
			&r.StartDate, &r.EndDate, &r.RoomID, &r.CreatedAt, &r.UpdatedAt,
			&r.Processed, &r.Room.ID, &r.Room.RoomName)

		if err != nil {
			return reservations, err
		}
		reservations = append(reservations, r)
	}
	if err = rows.Err(); err != nil {
		return reservations, err
	}
	return reservations, nil
}

// GetReservationByID returns a reservation by id
func (m *postgresDBRepo) GetReservationByID(id int) (models.Reservation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var res models.Reservation

	query := `select r.id, r.first_name, r.last_name, r.email, r.start_date, r.end_date, r.room_id,
		r.created_at, r.updated_at, r.processed,
		rm.id, rm.room_name
		from reservations r
		left join rooms rm on (r.room_id = rm.id)
		where r.id = $1
		order by r.start_date asc
	`
	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(&res.ID, &res.FirstName, &res.LastName, &res.Email, &res.StartDate, &res.EndDate, &res.RoomID,
		&res.CreatedAt, &res.UpdatedAt, &res.Processed, &res.Room.ID, &res.Room.RoomName)

	if err != nil {
		return res, err
	}
	return res, nil
}

// UpdateReservation updates the given reservation
func (m *postgresDBRepo) UpdateReservation(r models.Reservation) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update reservations set first_name = $1, last_name = $2, email = $3, 
	start_date = $4, end_date = $5, updated_at = $6 where id = $7`

	_, err := m.DB.ExecContext(ctx, query, r.FirstName, r.LastName, r.Email, r.StartDate, r.EndDate, r.UpdatedAt, r.ID)
	if err != nil {
		return err
	}

	return nil
}

// Delete Reservation deletes the reservation of given id from reservations table
func (m *postgresDBRepo) DeleteReservation(id int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `delete from reservations where id = $1`

	_, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

// UpdateProcessedForReservation updates the processed column of the reservation of the given id
func (m *postgresDBRepo) UpdateProcessedForReservation(id, processed int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `update reservations set processed = $1 where id = $2`

	_, err := m.DB.ExecContext(ctx, query, processed, id)

	if err != nil {
		return err
	}

	return nil
}

// GetRoomRestrictionsForRoomByDate returns a slice of RoomRestriction for the given roomID and date range
func (m *postgresDBRepo) GetRoomRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var roomRestrictions []models.RoomRestriction

	query := `
		select id, start_date, end_date, room_id, coalesce(reservation_id, 0), restriction_id
		from room_restrictions
		where room_id = $1 and not ($2 < start_date or $3 >= end_date)
	`
	rows, err := m.DB.QueryContext(ctx, query, roomID, end, start)
	if err != nil {
		return roomRestrictions, err
	}
	defer rows.Close()

	for rows.Next() {
		var r models.RoomRestriction
		err = rows.Scan(&r.ID, &r.StartDate, &r.EndDate, &r.RoomID, &r.ReservationID, &r.RestrictionID)
		if err != nil {
			return roomRestrictions, err
		}

		roomRestrictions = append(roomRestrictions, r)
	}

	if err = rows.Err(); err != nil {
		return roomRestrictions, err
	}

	return roomRestrictions, nil
}

// InsertBlockForRoom inserts a room restriction marking a room as blocked for one day
func (m *postgresDBRepo) InsertBlockForRoom(roomID int, start time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `insert into room_restrictions (start_date, end_date, room_id, restriction_id, created_at, updated_at) values ($1, $2, $3, $4, $5, $6)`

	_, err := m.DB.ExecContext(ctx, query, start, start.AddDate(0, 0, 1), roomID, 2, time.Now(), time.Now())
	if err != nil {
		return err
	}

	return nil
}

// DeleteBlockByRoomRestrictionID removes the room restriction for a room which was blocked
func (m *postgresDBRepo) DeleteBlockByRoomRestrictionID(ID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	query := `delete from room_restrictions where id = $1`
	_, err := m.DB.ExecContext(ctx, query, ID)
	if err != nil {
		return err
	}

	return nil
}
