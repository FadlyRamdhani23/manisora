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

var (
	ErrInvalidInput     = errors.New("input tidak valid")
	ErrTransferNotDraft = errors.New("hanya transfer berstatus DRAFT yang bisa dikirim")
	ErrTransferNotSent  = errors.New("hanya transfer berstatus SENT yang bisa diterima")
)

func friendlyTransferError(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1452 {
		return errors.New("lokasi atau barang tidak ditemukan")
	}
	return err
}

func CreateTransfer(t model.Transfer) (model.Transfer, error) {
	if t.FromLocationID <= 0 || t.ToLocationID <= 0 {
		return t, errors.New("from_location_id dan to_location_id wajib diisi")
	}
	if t.FromLocationID == t.ToLocationID {
		return t, errors.New("lokasi asal dan tujuan tidak boleh sama")
	}
	if len(t.Items) == 0 {
		return t, errors.New("minimal satu barang harus diisi")
	}

	seen := map[int64]bool{}
	for _, it := range t.Items {
		if it.ItemID <= 0 || it.QtySent <= 0 {
			return t, errors.New("setiap barang harus punya item_id dan qty_sent lebih dari 0")
		}
		if seen[it.ItemID] {
			return t, errors.New("barang yang sama tidak boleh muncul dua kali dalam satu transfer")
		}
		seen[it.ItemID] = true
	}

	if t.DocDate == "" {
		t.DocDate = time.Now().Format("2006-01-02")
	} else if _, err := time.Parse("2006-01-02", t.DocDate); err != nil {
		return t, errors.New("doc_date harus berformat YYYY-MM-DD")
	}
	t.Status = "DRAFT"

	tx, err := database.DB.Begin()
	if err != nil {
		return t, err
	}
	defer tx.Rollback()

	id, err := repository.CreateTransfer(tx, t)
	if err != nil {
		return t, friendlyTransferError(err)
	}
	t.ID = id
	t.DocNo = fmt.Sprintf("TRF-%s-%04d", t.DocDate[:4], id)

	if err := repository.SetTransferDocNo(tx, id, t.DocNo); err != nil {
		return t, err
	}

	for i := range t.Items {
		itemID, err := repository.CreateTransferItem(tx, id, t.Items[i])
		if err != nil {
			return t, friendlyTransferError(err)
		}
		t.Items[i].ID = itemID
		t.Items[i].TransferID = id
		t.Items[i].QtyReceived = 0
		t.Items[i].QtyDiff = 0
	}

	if err := tx.Commit(); err != nil {
		return t, err
	}
	return t, nil
}

func GetTransferByID(id int64) (model.Transfer, error) {
	t, err := repository.GetTransferByID(id)
	if errors.Is(err, sql.ErrNoRows) {
		return t, ErrNotFound
	}
	return t, err
}

// Kirim: stok keluar dari lokasi asal
func SendTransfer(id int64) (model.Transfer, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return model.Transfer{}, err
	}
	defer tx.Rollback()

	status, fromID, _, err := repository.LockTransfer(tx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Transfer{}, ErrNotFound
	}
	if err != nil {
		return model.Transfer{}, err
	}
	if status != "DRAFT" {
		return model.Transfer{}, ErrTransferNotDraft
	}

	items, err := repository.GetTransferItemsTx(tx, id)
	if err != nil {
		return model.Transfer{}, err
	}
	if len(items) == 0 {
		return model.Transfer{}, errors.New("dokumen tidak punya barang")
	}

	for _, it := range items {
		available, err := repository.GetStockForUpdate(tx, it.ItemID, fromID)
		if err != nil {
			return model.Transfer{}, err
		}
		if available < it.QtySent {
			return model.Transfer{}, fmt.Errorf("%w: barang id %d, tersedia %.3f, diminta %.3f",
				ErrInsufficientStock, it.ItemID, available, it.QtySent)
		}

		if err := repository.InsertLedger(tx, it.ItemID, fromID, 0, it.QtySent,
			"TRANSFER_OUT", "TRANSFER", id); err != nil {
			return model.Transfer{}, err
		}
		if err := repository.AddStockBalance(tx, it.ItemID, fromID, -it.QtySent); err != nil {
			return model.Transfer{}, err
		}
	}

	if err := repository.SetTransferStatus(tx, id, "SENT"); err != nil {
		return model.Transfer{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.Transfer{}, err
	}
	return GetTransferByID(id)
}

// Terima: stok masuk ke lokasi tujuan sebesar jumlah yang benar-benar diterima
func ReceiveTransfer(id int64, input model.ReceiveInput) (model.Transfer, error) {
	tx, err := database.DB.Begin()
	if err != nil {
		return model.Transfer{}, err
	}
	defer tx.Rollback()

	status, _, toID, err := repository.LockTransfer(tx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Transfer{}, ErrNotFound
	}
	if err != nil {
		return model.Transfer{}, err
	}
	if status != "SENT" {
		return model.Transfer{}, ErrTransferNotSent
	}

	items, err := repository.GetTransferItemsTx(tx, id)
	if err != nil {
		return model.Transfer{}, err
	}

	inTransfer := map[int64]bool{}
	for _, it := range items {
		inTransfer[it.ItemID] = true
	}

	received := map[int64]float64{}
	for _, r := range input.Items {
		if !inTransfer[r.ItemID] {
			return model.Transfer{}, fmt.Errorf("%w: barang id %d bukan bagian dari transfer ini", ErrInvalidInput, r.ItemID)
		}
		if _, dup := received[r.ItemID]; dup {
			return model.Transfer{}, fmt.Errorf("%w: barang id %d muncul dua kali", ErrInvalidInput, r.ItemID)
		}
		received[r.ItemID] = r.QtyReceived
	}

	for _, it := range items {
		qty := it.QtySent // barang yang tidak disebut dianggap diterima penuh
		if v, ok := received[it.ItemID]; ok {
			qty = v
		}
		if qty < 0 || qty > it.QtySent {
			return model.Transfer{}, fmt.Errorf("%w: qty_received barang id %d harus antara 0 dan %.3f",
				ErrInvalidInput, it.ItemID, it.QtySent)
		}

		diff := it.QtySent - qty
		if err := repository.SetTransferItemReceived(tx, it.ID, qty, diff); err != nil {
			return model.Transfer{}, err
		}

		if qty > 0 {
			if err := repository.InsertLedger(tx, it.ItemID, toID, qty, 0,
				"TRANSFER_IN", "TRANSFER", id); err != nil {
				return model.Transfer{}, err
			}
			if err := repository.AddStockBalance(tx, it.ItemID, toID, qty); err != nil {
				return model.Transfer{}, err
			}
		}
	}

	if err := repository.SetTransferStatus(tx, id, "RECEIVED"); err != nil {
		return model.Transfer{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.Transfer{}, err
	}
	return GetTransferByID(id)
}
