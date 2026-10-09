package services

import (
	"context"
	"fmt"
	"strings"

	"backend/shared/models"

	"github.com/jmoiron/sqlx"
)

type SupplierService struct {
	DB   *sqlx.DB
	Auth models.Auth
}

func NewSupplierService(db *sqlx.DB, auth models.Auth) *SupplierService {
	return &SupplierService{DB: db, Auth: auth}
}

func (s *SupplierService) NewSupplier() models.Supplier {
	return models.NewSupplier(s.DB, s.Auth)
}

func (s *SupplierService) Migrate() {
	s.NewSupplier().Migrate()
}

func (s *SupplierService) Create(supplier models.Supplier) (int64, error) {
	item := s.NewSupplier()
	item.AccountID = supplier.AccountID
	item.Code = supplier.Code
	item.Name = supplier.Name
	item.Email = supplier.Email
	item.PhoneNumber = supplier.PhoneNumber
	item.Address = supplier.Address
	item.ContactPerson = supplier.ContactPerson
	item.NPWP = supplier.NPWP
	item.BankName = supplier.BankName
	item.BankAccountNumber = supplier.BankAccountNumber
	item.BankAccountOwner = supplier.BankAccountOwner
	item.DebitLimit = supplier.DebitLimit
	item.Active = supplier.Active
	item.CreatedBy = supplier.CreatedBy
	if err := s.assignSupplierEmbedding(&item); err != nil {
		fmt.Println("[WARN] supplier embedding build failed:", err)
	}
	id, err := item.Create()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *SupplierService) Update(supplier models.Supplier) error {
	item := s.NewSupplier()
	item.ID = supplier.ID
	item.AccountID = supplier.AccountID
	item.Code = supplier.Code
	item.Name = supplier.Name
	item.Email = supplier.Email
	item.PhoneNumber = supplier.PhoneNumber
	item.Address = supplier.Address
	item.ContactPerson = supplier.ContactPerson
	item.NPWP = supplier.NPWP
	item.BankName = supplier.BankName
	item.BankAccountNumber = supplier.BankAccountNumber
	item.BankAccountOwner = supplier.BankAccountOwner
	item.DebitLimit = supplier.DebitLimit
	item.Active = supplier.Active
	item.UpdatedBy = supplier.UpdatedBy
	if err := s.assignSupplierEmbedding(&item); err != nil {
		fmt.Println("[WARN] supplier embedding build failed:", err)
	}
	if err := item.Update(); err != nil {
		return err
	}

	return nil
}

func (s *SupplierService) Delete(id int64) error {
	if err := s.deleteSupplierEmbedding(id); err != nil {
		fmt.Println("[WARN] supplier embedding delete failed:", err)
	}

	item := s.NewSupplier()
	item.ID = id
	return item.Delete()
}

func (s *SupplierService) FindByID(id int64) (*models.Supplier, error) {
	item := s.NewSupplier()
	item.ID = id
	return item.FindByID()
}

func (s *SupplierService) FindByCode(code string) (*models.Supplier, error) {
	item := s.NewSupplier()
	return item.FindByCode(code)
}

func (s *SupplierService) FindByEmail(email string) (*models.Supplier, error) {
	item := s.NewSupplier()
	return item.FindByEmail(email)
}

func (s *SupplierService) FindByPhoneNumber(phoneNumber string) (*models.Supplier, error) {
	item := s.NewSupplier()
	return item.FindByPhoneNumber(phoneNumber)
}

func (s *SupplierService) FindList(search string) ([]models.Supplier, error) {
	return s.NewSupplier().FindList(search)
}

func (s *SupplierService) FindAll() ([]models.Supplier, error) {
	return s.NewSupplier().FindAll()
}

func (s *SupplierService) FindRecentList(limit int, search string) ([]models.Supplier, error) {
	return s.NewSupplier().FindRecentList(limit, search)
}

func (s *SupplierService) CountByActive(active *bool) (int64, error) {
	return s.NewSupplier().CountByActive(active)
}

func (s *SupplierService) FindSemanticList(search string, limit int) ([]models.Supplier, error) {
	search = strings.TrimSpace(search)
	if search == "" {
		return s.FindRecentList(limit, "")
	}

	accountID := s.Auth.AccountID
	if accountID <= 0 {
		return nil, fmt.Errorf("id akun wajib diisi")
	}
	if limit <= 0 {
		limit = 5
	}

	embedding, err := GenerateOpenAIEmbedding(context.Background(), search)
	if err != nil {
		return s.FindList(search)
	}
	vectorLiteral := EmbeddingToVectorLiteral(embedding)
	if vectorLiteral == "" {
		return s.FindList(search)
	}

	sql := `SELECT ` + models.SupplierSelectColumns() + `
	FROM suppliers s
	WHERE s.account_id = $1 AND s.embedding IS NOT NULL
	ORDER BY s.embedding <=> $2::vector ASC
	LIMIT $3`

	var suppliers []models.Supplier
	if err := s.DB.Select(&suppliers, sql, accountID, vectorLiteral, limit); err != nil {
		return s.FindList(search)
	}
	return suppliers, nil
}

func (s *SupplierService) refreshSupplierEmbedding(id int64, supplier models.Supplier) error {
	if id <= 0 {
		return nil
	}

	vectorLiteral, err := s.supplierEmbeddingLiteral(supplier)
	if err != nil {
		return err
	}

	_, err = s.DB.Exec(`UPDATE suppliers SET embedding = $1::vector WHERE id = $2 AND account_id = $3`, vectorLiteral, id, s.Auth.AccountID)
	return err
}

func (s *SupplierService) deleteSupplierEmbedding(id int64) error {
	if id <= 0 || s.Auth.AccountID <= 0 {
		return nil
	}

	_, err := s.DB.Exec(`UPDATE suppliers SET embedding = NULL WHERE id = $1 AND account_id = $2`, id, s.Auth.AccountID)
	return err
}

func (s *SupplierService) assignSupplierEmbedding(supplier *models.Supplier) error {
	if supplier == nil {
		return nil
	}

	vectorLiteral, err := s.supplierEmbeddingLiteral(*supplier)
	if err != nil {
		return err
	}
	supplier.Embedding = vectorLiteral
	return nil
}

func (s *SupplierService) supplierEmbeddingLiteral(supplier models.Supplier) (string, error) {
	input := BuildEmbeddingText(
		supplier.Code,
		supplier.Name,
		supplier.Email,
		supplier.PhoneNumber,
		supplier.Address,
		supplier.ContactPerson,
		supplier.NPWP,
		supplier.BankName,
		supplier.BankAccountNumber,
		supplier.BankAccountOwner,
		fmt.Sprintf("%.2f", supplier.DebitLimit),
	)
	embedding, err := GenerateOpenAIEmbedding(context.Background(), input)
	if err != nil {
		return "", err
	}
	vectorLiteral := EmbeddingToVectorLiteral(embedding)
	if vectorLiteral == "" {
		return "", fmt.Errorf("embedding supplier kosong")
	}

	return vectorLiteral, nil
}
