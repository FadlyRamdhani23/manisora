package repository

import (
	"manisora/database"
	"manisora/model"
)

func GetAllLocations() ([]model.Location, error) {
	rows, err := database.DB.Query("SELECT id, code, name, type FROM locations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	locations := []model.Location{}
	for rows.Next() {
		var loc model.Location
		if err := rows.Scan(&loc.ID, &loc.Code, &loc.Name, &loc.Type); err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}
	return locations, rows.Err()
}

func CreateLocation(loc model.Location) (int64, error) {
	result, err := database.DB.Exec(
		"INSERT INTO locations (code, name, type) VALUES (?, ?, ?)",
		loc.Code, loc.Name, loc.Type,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
