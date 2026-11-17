package service

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-sql-driver/mysql"

	"manisora/database"
	"manisora/model"
	"manisora/repository"
)

var ErrInsufficientStock = errors.New("stok tidak cukup")

func friendlySaleError(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1452 {
		return errors.New("lokasi atau barang tidak ditemukan")
	}
	return err
}

func CreateSale(s model.Sale) (model.Sale, error) {
	if s.LocationID <= 0 {
		return s, errors.New("location_id wajib diisi")
	}
	if len(s.Items) == 0 {
		return s, errors.New("minimal satu barang harus diisi")
	}
	for _, it := range s.Items {
		if it.ItemID <= 0 || it.Qty <= 0 {
			return s, errors.New("setiap barang harus punya item_id dan qty lebih dari 0")
		}
	}

	if s.DocDate == "" {
		s.DocDate = time.Now().Format("2006-01-02")
	} else if _, err := time.Parse("2006-01-02", s.DocDate); err != nil {
		return s, errors.New("doc_date harus berformat YYYY-MM-DD")
	}
	s.Status = "DRAFT"

	tx, err := database.DB.Begin()
	if err != nil {
		return s, err
	}
	defer tx.Rollback()

	id, err := repository.CreateSale(tx, s)
	if err != nil {
		return s, friendlySaleError(err)
	}
	s.ID = id
	s.DocNo = fmt.Sprintf("SL-%s-%04d", s.DocDate[:4], id)

	if err := repository.SetSaleDocNo(tx, id, s.DocNo); err != nil {
		return s, err
	}

	for i := range s.Items {
		itemID, err := repository.CreateSaleItem(tx, id, s.Items[i])
		if err != nil {
			return s, friendlySaleError(err)
		}
		s.Items[i].ID = itemID
		s.Items[i].SaleID = id
	}

	if err := tx.Commit(); err != nil {
		return s, err
	}
	return s, nil
}

func GetSaleByID(id int64) (model.Sale, error) {
	s, err := repository.GetSaleByID(id)
	if errors.Is(err, sql.ErrNoRows) {
		return s, ErrNotFound
	}
	return s, err
}

func PostSale(id int64) (model.Sale, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return model.Sale{}, err
	}
	defer tx.Rollback()

	status, locationID, err := repository.LockSale(tx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Sale{}, ErrNotFound
	}
	if err != nil {
		return model.Sale{}, err
	}
	if status != "DRAFT" {
		return model.Sale{}, ErrNotDraft
	}

	items, err := repository.GetSaleItemsTx(tx, id)
	if err != nil {
		return model.Sale{}, err
	}
	if len(items) == 0 {
		return model.Sale{}, errors.New("dokumen tidak punya barang")
	}

	for _, it := range items {
		available, err := repository.GetStockForUpdate(tx, it.ItemID, locationID)
		if err != nil {
			return model.Sale{}, err
		}
		if available < it.Qty {
			return model.Sale{}, fmt.Errorf("%w: barang id %d, tersedia %.3f, diminta %.3f",
				ErrInsufficientStock, it.ItemID, available, it.Qty)
		}

		if err := repository.InsertLedger(tx, it.ItemID, locationID, 0, it.Qty,
			"SALE", "SALE", id); err != nil {
			return model.Sale{}, err
		}
		if err := repository.AddStockBalance(tx, it.ItemID, locationID, -it.Qty); err != nil {
			return model.Sale{}, err
		}
	}

	if err := repository.SetSaleStatus(tx, id, "POSTED"); err != nil {
		return model.Sale{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.Sale{}, err
	}
	return GetSaleByID(id)
}
