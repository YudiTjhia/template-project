package models_test

import (
	"testing"

	"backend/shared/models"
)

func newRealSupplierDB(t *testing.T) (*models.Supplier, func()) {
	t.Helper()

	if accountDB == nil {
		t.Skip("supplier db is not initialized")
	}

	auth := models.Auth{AccountID: 7, UserName: "alice", Email: "alice@example.com"}
	supplier := models.NewSupplier(accountDB, auth)
	_, _ = accountDB.Exec(`TRUNCATE TABLE suppliers RESTART IDENTITY CASCADE`)

	cleanup := func() {
		_, _ = accountDB.Exec(`TRUNCATE TABLE suppliers RESTART IDENTITY CASCADE`)
	}

	return &supplier, cleanup
}

func TestSupplierTableName(t *testing.T) {
	var supplier models.Supplier
	if got := supplier.TableName(); got != "suppliers" {
		t.Fatalf("expected table suppliers, got %q", got)
	}
}

func TestSupplierValidateCreate(t *testing.T) {
	tests := []struct {
		name     string
		supplier models.Supplier
		wantErr  string
	}{
		{
			name:     "missing account",
			supplier: models.Supplier{Code: "SUP-001", Name: "Supplier A"},
			wantErr:  "id akun wajib diisi",
		},
		{
			name:     "missing code",
			supplier: models.Supplier{BaseModel: models.BaseModel{Auth: models.Auth{AccountID: 1}}, Name: "Supplier A"},
			wantErr:  "code supplier wajib diisi",
		},
		{
			name:     "missing name",
			supplier: models.Supplier{BaseModel: models.BaseModel{Auth: models.Auth{AccountID: 1}}, Code: "SUP-001"},
			wantErr:  "nama supplier wajib diisi",
		},
		{
			name: "valid",
			supplier: models.Supplier{
				BaseModel: models.BaseModel{Auth: models.Auth{AccountID: 1}},
				Code:      "SUP-001",
				Name:      "Supplier A",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.supplier.ValidateCreate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}

			if err == nil || err.Error() != tt.wantErr {
				t.Fatalf("expected %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestSupplierCreateUpdateDelete(t *testing.T) {
	supplier, cleanup := newRealSupplierDB(t)
	defer cleanup()

	supplier.Code = "SUP-001"
	supplier.Name = "Supplier A"
	supplier.Email = "supplier@example.com"
	supplier.PhoneNumber = "08123456789"
	supplier.Address = "Jl. Supplier 1"
	supplier.ContactPerson = "Budi"
	supplier.NPWP = "01.234.567.8-901.000"
	supplier.BankName = "BCA"
	supplier.BankAccountNumber = "1234567890"
	supplier.BankAccountOwner = "PT Supplier A"
	supplier.DebitLimit = 25000000

	id, err := supplier.Create()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}

	supplier.ID = id
	found, err := supplier.FindByID()
	if err != nil {
		t.Fatalf("expected no error reading back, got %v", err)
	}
	if found == nil || found.Code != "SUP-001" || found.Name != "Supplier A" {
		t.Fatalf("unexpected supplier: %+v", found)
	}
	if found.Embedding != "" {
		t.Fatalf("expected empty embedding string for null embedding, got %q", found.Embedding)
	}
	if found.NPWP != "01.234.567.8-901.000" || found.BankName != "BCA" || found.BankAccountNumber != "1234567890" || found.BankAccountOwner != "PT Supplier A" || found.DebitLimit != 25000000 {
		t.Fatalf("unexpected supplier banking fields: %+v", found)
	}
	if !found.Active {
		t.Fatalf("expected supplier to be active by default")
	}

	updatedBy := "bob@example.com"
	supplier.Code = "SUP-002"
	supplier.Name = "Supplier B"
	supplier.Email = "supplierb@example.com"
	supplier.PhoneNumber = "08123456780"
	supplier.Address = "Jl. Supplier 2"
	supplier.ContactPerson = "Andi"
	supplier.NPWP = "02.345.678.9-012.000"
	supplier.BankName = "Mandiri"
	supplier.BankAccountNumber = "9876543210"
	supplier.BankAccountOwner = "PT Supplier B"
	supplier.DebitLimit = 50000000
	supplier.Active = false
	supplier.UpdatedBy = &updatedBy
	if err := supplier.Update(); err != nil {
		t.Fatalf("expected no error updating, got %v", err)
	}

	found, err = supplier.FindByID()
	if err != nil {
		t.Fatalf("expected no error reading back after update, got %v", err)
	}
	if found.Code != "SUP-002" || found.Name != "Supplier B" || found.Active {
		t.Fatalf("unexpected updated supplier: %+v", found)
	}
	if found.NPWP != "02.345.678.9-012.000" || found.BankName != "Mandiri" || found.BankAccountNumber != "9876543210" || found.BankAccountOwner != "PT Supplier B" || found.DebitLimit != 50000000 {
		t.Fatalf("unexpected updated banking fields: %+v", found)
	}

	if err := supplier.Delete(); err != nil {
		t.Fatalf("expected no error deleting, got %v", err)
	}
	if _, err := supplier.FindByID(); err == nil {
		t.Fatalf("expected error after delete")
	}
}

func TestSupplierFindListAndLookups(t *testing.T) {
	supplier, cleanup := newRealSupplierDB(t)
	defer cleanup()

	rows := []models.Supplier{
		{
			BaseModel:         models.BaseModel{DB: accountDB, Auth: models.Auth{AccountID: 1}},
			Code:              "SUP-001",
			Name:              "Alpha Supplier",
			Email:             "alpha@example.com",
			PhoneNumber:       "0811111111",
			Address:           "Jl. Alpha",
			ContactPerson:     "Alpha Contact",
			NPWP:              "01.111.111.1-111.000",
			BankName:          "BCA",
			BankAccountNumber: "1111111111",
			BankAccountOwner:  "Alpha Supplier",
			DebitLimit:        10000000,
		},
		{
			BaseModel:         models.BaseModel{DB: accountDB, Auth: models.Auth{AccountID: 1}},
			Code:              "SUP-002",
			Name:              "Beta Supplier",
			Email:             "beta@example.com",
			PhoneNumber:       "0822222222",
			Address:           "Jl. Beta",
			ContactPerson:     "Beta Contact",
			NPWP:              "02.222.222.2-222.000",
			BankName:          "Mandiri",
			BankAccountNumber: "2222222222",
			BankAccountOwner:  "Beta Supplier",
			DebitLimit:        20000000,
		},
	}

	for _, row := range rows {
		row := row
		if _, err := row.Create(); err != nil {
			t.Fatalf("expected no error creating supplier, got %v", err)
		}
	}

	foundByCode, err := supplier.FindByCode("sup-001")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if foundByCode == nil || foundByCode.Name != "Alpha Supplier" {
		t.Fatalf("unexpected supplier by code: %+v", foundByCode)
	}

	foundByEmail, err := supplier.FindByEmail("beta@example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if foundByEmail == nil || foundByEmail.Code != "SUP-002" {
		t.Fatalf("unexpected supplier by email: %+v", foundByEmail)
	}

	foundByPhone, err := supplier.FindByPhoneNumber("0822222222")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if foundByPhone == nil || foundByPhone.Code != "SUP-002" {
		t.Fatalf("unexpected supplier by phone: %+v", foundByPhone)
	}

	items, err := supplier.FindList("Supplier")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}
