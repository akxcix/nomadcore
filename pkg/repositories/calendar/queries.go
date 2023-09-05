package calendar

import (
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func (db *Database) CreateCalendar(userId uuid.UUID, name, visibility string) error {
	query := `
		INSERT 
			INTO 
				public.calendars (user_id, name, visibility) 
			VALUES 
				($1, $2, $3)
	`

	tx := db.db.MustBegin()
	_, err := tx.Exec(query, userId, name, visibility)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (db *Database) GetCalendars(userID uuid.UUID, visibility Visibility) ([]Calendar, error) {
	calendars := []Calendar{}
	query := `
		SELECT 
			* 
		FROM 
			public.calendars 
		WHERE 
			user_id = $1 
			AND visibility = $2
	`

	// QueryRow still works, but now we're scanning into multiple variables
	err := db.db.Select(&calendars, query, userID, visibility)
	if err != nil {
		return nil, err
	}

	return calendars, nil
}

func (db *Database) GetDates(calendarIDs []uuid.UUID) ([]Date, error) {
	dates := []Date{}
	query := `
		SELECT 
			*
		FROM
			public.dates
		WHERE
			calendar_id in (?)
		LIMIT
			50
	`

	q, vs, err := sqlx.In(query, calendarIDs)
	if err != nil {
		return nil, err
	}

	q = db.db.Rebind(q)

	err = db.db.Select(&dates, q, vs...)
	if err != nil {
		return nil, err
	}

	return dates, nil
}

func (db *Database) AddDatesToCalendar(userID, calendarID uuid.UUID, from, to time.Time) error {
	query := `
	INSERT INTO 
		dates (from_date, to_date, calendar_id)
	SELECT 
		$1, $2, $3
	WHERE 
		EXISTS (
			SELECT 
				1
			FROM 
				calendars
			WHERE 
				id = $3 
				AND user_id = $4
		);
	`

	tx := db.db.MustBegin()
	_, err := tx.Exec(query, from, to, calendarID, userID)
	if err != nil {
		return err
	}
	return tx.Commit()
}
