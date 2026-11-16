package service

import (
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"

	"manisora/model"
	"manisora/repository"
)

func GetAllLocations() ([]model.Location, error) {
	return repository.GetAllLocations()
}

func CreateLocation(loc model.Location) (model.Location, error) {
	loc.Code = strings.TrimSpace(loc.Code)
	loc.Name = strings.TrimSpace(loc.Name)
	loc.Type = strings.ToUpper(strings.TrimSpace(loc.Type))

	if loc.Code == "" || loc.Name == "" || loc.Type == "" {
		return loc, errors.New("code, name, dan type wajib diisi")
	}

	if loc.Type != "TOKO" && loc.Type != "GUDANG" && loc.Type != "UNIT" {
		return loc, errors.New("type harus TOKO, GUDANG, atau UNIT")
	}

	id, err := repository.CreateLocation(loc)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return loc, errors.New("kode lokasi sudah dipakai")
		}
		return loc, err
	}
	loc.ID = id
	return loc, nil
}
