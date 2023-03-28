package dbrepo

import (
	"errors"
	"time"

	"github.com/atuprosper/booking-project/internal/models"
)

func (repo *testDBRepo) AllUsers() bool {
	return true
}

// Inserts a reservation into the database
func (repo *testDBRepo) InsertReservation(res models.Reservation) (int, error) {
	// Fail test if the room_id == 2
	if res.RoomID == 2 {
		return 0, errors.New("failed to insert reservation")
	}
	return 1, nil
}

// InsertRoomRestriction inserts a room restriction into the database
func (repo *testDBRepo) InsertRoomRestriction(res models.RoomRestriction) error {
	return nil
}

// SearchAvailabilityByDatesByRoomID returns true if availability exists for roomID, and false if no availability
func (repo *testDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error) {
	return false, nil
}

// SearchAvailabilityForAllRooms returns a slice of available rooms, if any, for given date range
func (repo *testDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	var rooms []models.Room

	return rooms, nil
}

// GetRoomByID gets a room by id
func (repo *testDBRepo) GetRoomByID(id int) (models.Room, error) {
	var room models.Room

	return room, nil
}
