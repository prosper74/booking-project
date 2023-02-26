package dbrepo

import (
	"context"
	"time"

	"github.com/atuprosper/booking-project/internal/models"
)

func (repo *postgresDBRepo) AllUsers() bool {
	return true
}

// Inserts a reservation into the database
func (repo *postgresDBRepo) InsertReservation(res models.Reservation) (int, error) {
	// Close this transaction if unable to run this statement within 3 seconds
	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var newID int

	insertStatement := `insert into reservations (first_name, last_name, email, phone, start_date, end_date, room_id, created_at, updated_at) values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`

	err := repo.DB.QueryRowContext(context, insertStatement, res.FirstName, res.LastName, res.Email, res.Phone, res.StartDate, res.EndDate, res.RoomID, time.Now(), time.Now()).Scan(newID)

	if err != nil {
		return 0, err
	}

	return newID, nil
}

func (repo *postgresDBRepo) InsertRoomRestriction(res models.RoomRestriction) error {
	context, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	insertStatement := `insert into room_restrictions (start_date, end_date, room_id, reservation_id, created_at, updated_at, restriction_id) values($1, $2, $3, $4, $5, $6, $7)`

	_, err := repo.DB.ExecContext(context, insertStatement, res.StartDate, res.EndDate, res.RoomID, res.ReservationID, time.Now(), time.Now(), res.RestrictionID)

	if err != nil {
		return err
	}

	return nil
}
