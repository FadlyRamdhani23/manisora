package repository

import (
	"manisora/database"
	"manisora/model"
)

func GetAllItems() ([]model.Item, error) {
	rows, err := database.DB.Query("SELECT id, code, name, base_unit FROM items")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []model.Item{}
	for rows.Next() {
		var it model.Item
		if err := rows.Scan(&it.ID, &it.Code, &it.Name, &it.BaseUnit); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func CreateItem(item model.Item) (int64, error) {
	result, err := database.DB.Exec(
		"INSERT INTO items (code, name, base_unit) VALUES (?, ?, ?)",
		item.Code, item.Name, item.BaseUnit,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}
