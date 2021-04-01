package dbrepo

import (
	"time"

	"github.com/jackedelic/bookings/internal/models"
)

func (m *testingDBRepo) AllUsers() bool {
	return true
}

func (m *testingDBRepo) InsertReservation(res models.Reservation) (int, error) {
	return 1, nil
}

func (m *testingDBRepo) InsertRoomRestriction(rr models.RoomRestriction) error {
	return nil
}

func (m *testingDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	return false, nil
}

func (m *testingDBRepo) SearchAvailableRoomsByDates(start, end time.Time) ([]models.Room, error) {
	var rooms []models.Room
	return rooms, nil
}

func (m *testingDBRepo) GetRoomById(id int) (models.Room, error) {
	var room models.Room
	return room, nil
}
