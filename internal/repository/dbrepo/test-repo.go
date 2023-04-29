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

// GetUserByID returns a user by id
func (repo *testDBRepo) GetUserByID(id int) (models.User, error) {
	var user models.User

	return user, nil
}

// UpdateUser updates a user in the database
func (repo *testDBRepo) UpdateUser(user models.User) error {
	return nil
}

// Authenticate authenticates a user
func (repo *testDBRepo) Authenticate(email, testPassword string) (int, string, error) {
	return 1, "", nil
}

// AllReservations returns a slice of all reservations
func (repo *testDBRepo) AllReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation

	return reservations, nil
}

// AllNewReservations returns a slice of all reservations
func (m *testDBRepo) AllNewReservations() ([]models.Reservation, error) {
	var reservations []models.Reservation

	return reservations, nil
}

// GetReservationByID returns one reservation by ID
func (m *testDBRepo) GetReservationByID(id int) (models.Reservation, error) {
	var reservation models.Reservation

	return reservation, nil
}

// UpdateReservation updates a reservation in the database
func (m *testDBRepo) UpdateReservation(u models.Reservation) error {
	return nil
}

// DeleteReservation deletes one reservation by id
func (m *testDBRepo) DeleteReservation(id int) error {
	return nil
}

// UpdateProcessedForReservation updates processed for a reservation by id
func (m *testDBRepo) UpdateProcessedForReservation(id, processed int) error {
	return nil
}

// Get all rooms
func (m *testDBRepo) AllRooms() ([]models.Room, error) {
	var rooms []models.Room

	return rooms, nil
}

// Get the restrictions for a room
func (m *testDBRepo) GetRestrictionsForCurrentRoom(roomID int, start, end time.Time) ([]models.RoomRestriction, error) {
	var restrictions []models.RoomRestriction

	return restrictions, nil
}

// InsertBlockForRoom inserts a room restriction
func (m *testDBRepo) InsertBlockForRoom(id int, startDate time.Time) error {
	return nil
}

// DeleteBlockByID deletes a room restriction
func (m *testDBRepo) DeleteBlockByID(id int) error {
	return nil
}

// UpdateRoom updates a room in the database
func (m *testDBRepo) UpdateRoom(room models.Room) error {
	return nil
}

// Inserts a room into the database
func (repo *testDBRepo) InsertRoom(room models.Room) error {
	return nil
}

// DeleteRoom deletes a room
func (m *testDBRepo) DeleteRoom(id int) error {
	return nil
}

// InsertTodoList inserts a new todo list into the database
func (repo *testDBRepo) InsertTodoList(todo models.TodoList) error {
	return nil
}

// GetTodoListByUserID gets all todo for a user by user_id
func (repo *testDBRepo) GetTodoListByUserID(id int) ([]models.TodoList, error) {
	var todoList []models.TodoList
	return todoList, nil
}

// DeleteTodo deletes a todo
func (m *testDBRepo) DeleteTodo(id int) error {
	return nil
}
