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
	GetAllRooms() ([]models.Room, error)

	GetUserById(id int) (models.User, error)
	UpdateUser(u models.User) error
	Authenticate(email, testPassword string) (int, string, error)

	GetAllReservations() ([]models.Reservation, error)
	GetAllNewReservations() ([]models.Reservation, error)
	GetReservationByID(id int) (models.Reservation, error)
	UpdateReservation(r models.Reservation) error
	DeleteReservation(id int) error
	UpdateProcessedForReservation(id, processed int) error

	GetRoomRestrictionsForRoomByDate(roomID int, start, end time.Time) ([]models.RoomRestriction, error)
	InsertBlockForRoom(roomID int, start time.Time) error
	DeleteBlockByRoomRestrictionID(roomID int) error
}
