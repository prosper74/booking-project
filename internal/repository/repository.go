package repository

import (
	"time"

	"github.com/atuprosper/booking-project/internal/models"
)

type DatabaseRepo interface {
	AllUsers() bool

	InsertReservation(res models.Reservation) (int, error)
	InsertRoomRestriction(res models.RoomRestriction) error
	SearchAvailabilityByDatesByRoomID(start, end time.Time, roomID int) (bool, error)
	SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error)
	GetRoomByID(id int) (models.Room, error)

	GetUserByID(id int) (models.User, error)
	UpdateUser(user models.User) error
	Authenticate(email, testPassword string) (int, string, error)

	AllReservations() ([]models.Reservation, error)
	AllNewReservations() ([]models.Reservation, error)

	GetReservationByID(id int) (models.Reservation, error)
	UpdateReservation(u models.Reservation) error
	DeleteReservation(id int) error
	UpdateProcessedForReservation(id, processed int) error
	InsertBlockForRoom(id int, startDate time.Time) error
	DeleteBlockByID(id int) error

	AllRooms() ([]models.Room, error)
	UpdateRoom(room models.Room) error
	InsertRoom(room models.Room) error
	DeleteRoom(id int) error

	InsertTodoList(todo models.TodoList) error
	GetTodoListByUserID(id int) ([]models.TodoList, error)
	DeleteTodo(id int) error

	GetRestrictionsForCurrentRoom(roomID int, start, end time.Time) ([]models.RoomRestriction, error)
}
