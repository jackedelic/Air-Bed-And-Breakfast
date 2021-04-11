package dbrepo

import (
	"errors"
	"time"

	"github.com/jackedelic/bookings/internal/models"
)

func (m *testingDBRepo) AllUsers() bool {
	return true
}

// InsertReservation returns 0 and error if rr.RoomID == 2
func (m *testingDBRepo) InsertReservation(res models.Reservation) (int, error) {
	if res.RoomID == 2 {
		return 0, errors.New("error inserting reservation with room id of 2")
	}
	return 1, nil
}

// InsertRoomRestriction returns error if rr.RoomID == 1000
func (m *testingDBRepo) InsertRoomRestriction(rr models.RoomRestriction) error {
	if rr.RoomID == 1000 {
		return errors.New("error inserting room restriction with room id of 1000")
	}
	return nil
}

// SearchAvailabilityByDatesByRoomID returns false and error for room id of 1 and dates 01-01-2050 to 01-01-2050
func (m *testingDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	if (start.Format("02-01-2006") == "01-01-2050" || end.Format("02-01-2006") == "01-01-2050") &&
		roomID == 1 {
		return false, errors.New("rooms not available")
	}
	return false, nil
}

func (m *testingDBRepo) SearchAvailableRoomsByDates(start, end time.Time) ([]models.Room, error) {
	var rooms []models.Room
	return rooms, nil
}

// GetRoomById returns empty room with an error for room id > 2, otherwise nil.
func (m *testingDBRepo) GetRoomById(id int) (models.Room, error) {
	var room models.Room
	if room.ID > 2 {
		return room, errors.New("no room with id > 2")
	}
	return room, nil
}

// GetUserByID returns the user by id
func (m *testingDBRepo) GetUserById(id int) (models.User, error) {
	var u models.User
	return u, nil
}

// UpdateUser updates the user
func (m *testingDBRepo) UpdateUser(u models.User) error {
	return nil
}

// Authenticate authenticates user by user-given email and password
func (m *testingDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	return 0, "", nil
}

// GetAllReservations returns a slice of all reservations
func (m *testingDBRepo) GetAllReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation
	return reservations, nil
}

// GetAllNewReservations returns a slice of all new reservations (processed = 0)
func (m *testingDBRepo) GetAllNewReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation
	return reservations, nil
}

// GetReservationByID returns a reservation by id
func (m *testingDBRepo) GetReservationByID(id int) (models.Reservation, error) {
	var res models.Reservation
	return res, nil
}

// UpdateReservation updates the given reservation
func (m *testingDBRepo) UpdateReservation(r models.Reservation) error {
	return nil
}

// Delete Reservation deletes the reservation of given id from reservations table
func (m *testingDBRepo) DeleteReservation(id int) error {
	return nil
}

// UpdateProcessedForReservation updates the processed column of the reservation of the given id
func (m *testingDBRepo) UpdateProcessedForReservation(id, processed int) error {
	return nil
}
