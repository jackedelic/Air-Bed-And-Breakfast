package repository

import (
	"time"

	"github.com/jackedelic/bookings/internal/models"
)

type DatabaseRepo interface {
	AllUsers() bool
	InsertReservation(res models.Reservation) (int, error)
	InsertRoomRestriction(rr models.RoomRestriction) error
	SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error)
	SearchAvailableRoomsByDates(start, end time.Time) ([]models.Room, error)
	GetRoomById(id int) (models.Room, error)
}
