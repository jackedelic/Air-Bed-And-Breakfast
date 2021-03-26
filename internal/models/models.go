package models

import "time"

// Users represents the users table
type User struct {
	ID          int
	FirstName   string
	LastName    string
	Email       string
	Password    string
	AccessLevel string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Rooms represents the rooms table
type Room struct {
	ID        int
	RoomName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Restriction represents the restrictions table. It is either a Reservation or Owner Block type.
type Restriction struct {
	ID              int
	RestrictionName string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Reservation represents the reservations table
type Reservation struct {
	ID        int
	FirstName string
	LastName  string
	Email     string
	StartDate time.Time
	EndDate   time.Time
	RoomID    int
	CreatedAt time.Time
	UpdatedAt time.Time
	Room      Room // Optional
}

// RoomRestriction represents room_restrictions table
type RoomRestriction struct {
	ID            int
	StartDate     time.Time
	EndDate       time.Time
	RoomID        int
	ReservationID int
	RestrictionID int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Room          Room        // Optional
	Reservation   Reservation // Optional
	Restriction   Restriction // Optional
}
