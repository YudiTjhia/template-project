package services_test

import (
	"testing"

	"backend/shared/models"
	"backend/shared/services"
)

func newRealSupplierService(t *testing.T, auth models.Auth) (*services.SupplierService, func()) {
	t.Helper()

	if authServiceDB == nil {
		t.Skip("supplier service db is not initialized")
	}

	models.NewSupplier(authServiceDB, auth).Migrate()
	service := services.NewSupplierService(authServiceDB, auth)
	_, _ = authServiceDB.Exec(`TRUNCATE TABLE suppliers RESTART IDENTITY CASCADE`)

	cleanup := func() {
		_, _ = authServiceDB.Exec(`TRUNCATE TABLE suppliers RESTART IDENTITY CASCADE`)
	}

	return service, cleanup
}

func TestSupplierServiceCRUD(t *testing.T) {
	service, cleanup := newRealSupplierService(t, models.Auth{AccountID: 7, UserName: "alice", Email: "alice@example.com"})
	defer cleanup()

	id, err := service.Create(models.Supplier{
		Code:              "SUP-001",
		Name:              "Supplier A",
		Email:             "supplier@example.com",
		PhoneNumber:       "0811111111",
		Address:           "Jl. Supplier 1",
		ContactPerson:     "Budi",
		NPWP:              "01.234.567.8-901.000",
		BankName:          "BCA",
		BankAccountNumber: "1234567890",
		BankAccountOwner:  "PT Supplier A",
		DebitLimit:        25000000,
		Active:            true,
	})
	if err != nil {
		t.Fatalf("expected no error creating supplier, got %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}

	found, err := service.FindByID(id)
	if err != nil {
		t.Fatalf("expected no error finding supplier, got %v", err)
	}
	if found == nil || found.Code != "SUP-001" || found.Name != "Supplier A" {
		t.Fatalf("unexpected supplier: %+v", found)
	}
	if found.NPWP != "01.234.567.8-901.000" || found.BankName != "BCA" || found.BankAccountNumber != "1234567890" || found.BankAccountOwner != "PT Supplier A" || found.DebitLimit != 25000000 {
		t.Fatalf("unexpected supplier banking fields: %+v", found)
	}

	updatedBy := "bob@example.com"
	if err := service.Update(models.Supplier{
		ID:                id,
		Code:              "SUP-002",
		Name:              "Supplier B",
		Email:             "supplierb@example.com",
		PhoneNumber:       "0811111112",
		Address:           "Jl. Supplier 2",
		ContactPerson:     "Andi",
		NPWP:              "02.345.678.9-012.000",
		BankName:          "Mandiri",
		BankAccountNumber: "9876543210",
		BankAccountOwner:  "PT Supplier B",
		DebitLimit:        50000000,
		Active:            false,
		BaseModel: models.BaseModel{
			UpdatedBy: &updatedBy,
		},
	}); err != nil {
		t.Fatalf("expected no error updating supplier, got %v", err)
	}

	found, err = service.FindByCode("sup-002")
	if err != nil {
		t.Fatalf("expected no error finding by code, got %v", err)
	}
	if found == nil || found.Email != "supplierb@example.com" || found.Active {
		t.Fatalf("unexpected updated supplier: %+v", found)
	}
	if found.NPWP != "02.345.678.9-012.000" || found.BankName != "Mandiri" || found.BankAccountNumber != "9876543210" || found.BankAccountOwner != "PT Supplier B" || found.DebitLimit != 50000000 {
		t.Fatalf("unexpected updated banking fields: %+v", found)
	}

	found, err = service.FindByEmail("SUPPLIERB@example.com")
	if err != nil {
		t.Fatalf("expected no error finding by email, got %v", err)
	}
	if found == nil || found.ID != id {
		t.Fatalf("unexpected supplier by email: %+v", found)
	}

	found, err = service.FindByPhoneNumber("0811111112")
	if err != nil {
		t.Fatalf("expected no error finding by phone number, got %v", err)
	}
	if found == nil || found.ID != id {
		t.Fatalf("unexpected supplier by phone number: %+v", found)
	}

	items, err := service.FindList("Supplier")
	if err != nil {
		t.Fatalf("expected no error listing suppliers, got %v", err)
	}
	if len(items) != 1 || items[0].ID != id {
		t.Fatalf("unexpected list result: %+v", items)
	}

	recent, err := service.FindRecentList(10, "Supplier")
	if err != nil {
		t.Fatalf("expected no error finding recent suppliers, got %v", err)
	}
	if len(recent) != 1 || recent[0].ID != id {
		t.Fatalf("unexpected recent result: %+v", recent)
	}

	if err := service.Delete(id); err != nil {
		t.Fatalf("expected no error deleting supplier, got %v", err)
	}
	if _, err := service.FindByID(id); err == nil {
		t.Fatalf("expected error after delete")
	}
}

func TestSupplierServiceEnforcesAccountScopeAndDuplicates(t *testing.T) {
	service, cleanup := newRealSupplierService(t, models.Auth{AccountID: 7, UserName: "alice", Email: "alice@example.com"})
	defer cleanup()

	if _, err := service.Create(models.Supplier{
		Code:              "SUP-001",
		Name:              "Supplier A",
		Email:             "supplier@example.com",
		PhoneNumber:       "0811111111",
		Address:           "Jl. Supplier 1",
		ContactPerson:     "Budi",
		NPWP:              "01.234.567.8-901.000",
		BankName:          "BCA",
		BankAccountNumber: "1234567890",
		BankAccountOwner:  "PT Supplier A",
		DebitLimit:        25000000,
		Active:            true,
	}); err != nil {
		t.Fatalf("expected no error creating first supplier, got %v", err)
	}

	otherService := services.NewSupplierService(authServiceDB, models.Auth{AccountID: 8, UserName: "bob", Email: "bob@example.com"})
	if _, err := otherService.Create(models.Supplier{
		Code:              "SUP-001",
		Name:              "Supplier A",
		Email:             "supplier@example.com",
		PhoneNumber:       "0811111111",
		Address:           "Jl. Supplier 1",
		ContactPerson:     "Budi",
		NPWP:              "01.234.567.8-901.000",
		BankName:          "BCA",
		BankAccountNumber: "1234567890",
		BankAccountOwner:  "PT Supplier A",
		DebitLimit:        25000000,
		Active:            true,
	}); err != nil {
		t.Fatalf("expected same code to be allowed for different account, got %v", err)
	}

	if _, err := service.Create(models.Supplier{
		Code:              "SUP-001",
		Name:              "Supplier B",
		Email:             "supplierb@example.com",
		PhoneNumber:       "0811111112",
		Address:           "Jl. Supplier 2",
		ContactPerson:     "Andi",
		NPWP:              "02.345.678.9-012.000",
		BankName:          "Mandiri",
		BankAccountNumber: "9876543210",
		BankAccountOwner:  "PT Supplier B",
		DebitLimit:        50000000,
		Active:            true,
	}); err == nil || err.Error() != "code supplier sudah digunakan" {
		t.Fatalf("expected duplicate code error, got %v", err)
	}

	if _, err := service.Create(models.Supplier{
		Code:              "SUP-002",
		Name:              "Supplier B",
		Email:             "SUPPLIER@example.com",
		PhoneNumber:       "0811111112",
		Address:           "Jl. Supplier 2",
		ContactPerson:     "Andi",
		NPWP:              "02.345.678.9-012.000",
		BankName:          "Mandiri",
		BankAccountNumber: "9876543210",
		BankAccountOwner:  "PT Supplier B",
		DebitLimit:        50000000,
		Active:            true,
	}); err == nil || err.Error() != "email supplier sudah digunakan" {
		t.Fatalf("expected duplicate email error, got %v", err)
	}

	if _, err := service.Create(models.Supplier{
		Code:              "SUP-003",
		Name:              "Supplier C",
		Email:             "supplierc@example.com",
		PhoneNumber:       "0811111111",
		Address:           "Jl. Supplier 3",
		ContactPerson:     "Cici",
		NPWP:              "03.456.789.0-123.000",
		BankName:          "BNI",
		BankAccountNumber: "1111222233",
		BankAccountOwner:  "PT Supplier C",
		DebitLimit:        75000000,
		Active:            true,
	}); err == nil || err.Error() != "phone number supplier sudah digunakan" {
		t.Fatalf("expected duplicate phone number error, got %v", err)
	}

	otherScope := services.NewSupplierService(authServiceDB, models.Auth{AccountID: 8, UserName: "bob", Email: "bob@example.com"})
	if _, err := otherScope.FindByCode("SUP-001"); err != nil {
		t.Fatalf("expected other account to find its own supplier, got %v", err)
	}
	if _, err := otherScope.FindByCode("missing"); err == nil {
		t.Fatalf("expected error for missing supplier")
	}
}
