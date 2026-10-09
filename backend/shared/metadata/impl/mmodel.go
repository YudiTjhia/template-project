package impl

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"backend/shared/metadata/interfaces"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type MModel struct {
	Context   context.Context `json:"-" db:"-"`
	Name      string
	TableName string

	Fields []interfaces.IMField

	DB *sqlx.DB `json:"-" db:"-"`
}

// MWhere is a structured custom condition for FindList. Field names and
// operators are validated before the query is built; Value is parameterized.
type MWhere struct {
	Field    string
	Operator string
	Value    any
}

// MOrderField describes one metadata field and its sort direction.
type MOrderField struct {
	Field     string
	Direction string
}

// MListOptions controls filtering, search, sorting, and pagination for FindList.
// Page is one-based. A zero Page selects the first page; a zero Limit uses 50.
// Limits above 500 are capped at 500.
type MListOptions struct {
	Filters      map[string]any
	Where        []MWhere
	Search       string
	SearchFields []string
	OrderFields  []MOrderField
	Page         int
	Limit        int
}

var uuidPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func NewMModel(ctx context.Context, db *sqlx.DB) *MModel {

	// validate here
	if db == nil {
		return nil
	}

	mm := &MModel{Context: contextOrBackground(ctx), DB: db}
	err := validateModelMetadata(mm)
	if err != nil {
		return nil
	}

	return mm
}

func contextOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func (m *MModel) ContextOrBackground() context.Context {
	if m == nil {
		return context.Background()
	}
	return contextOrBackground(m.Context)
}

// ResolveAccountID prefers the authenticated account ID and falls back to the
// account ID on the model value.
func (m *MModel) ResolveAccountID(authAccountID, accountID int64) (int64, error) {
	if m == nil {
		return 0, errors.New("model is nil")
	}
	if authAccountID > 0 {
		return authAccountID, nil
	}
	if accountID > 0 {
		return accountID, nil
	}
	return 0, errors.New("id akun wajib diisi")
}

// ValidateAccountID_ verifies the model has an account_id metadata field and
// returns the authenticated account ID, falling back to the model value.
func (m *MModel) ValidateAccountID_(authAccountID, accountID int64) (int64, error) {
	if m == nil {
		return 0, errors.New("model is nil")
	}
	for _, field := range m.GetFields() {
		if strings.EqualFold(fieldColumnName(field), "account_id") {
			return m.ResolveAccountID(authAccountID, accountID)
		}
	}
	return 0, fmt.Errorf("model %q tidak memiliki field account_id", m.Name)
}

// ResolveCreatedBy_ validates the created_by metadata field and resolves its
// value from the model input or the authenticated user name.
func (m *MModel) ResolveCreatedBy_(createdBy, authenticatedUser string) (string, error) {
	if m == nil {
		return "", errors.New("model is nil")
	}
	for _, field := range m.GetFields() {
		if strings.EqualFold(fieldColumnName(field), "created_by") {
			createdBy = strings.TrimSpace(createdBy)
			if createdBy == "" {
				createdBy = strings.TrimSpace(authenticatedUser)
			}
			if createdBy == "" {
				return "", errors.New("created_by wajib diisi")
			}
			return createdBy, nil
		}
	}
	return "", fmt.Errorf("model %q tidak memiliki field created_by", m.Name)
}

// ResolveUpdatedBy_ checks the updated_by metadata field and uses the
// authenticated user when no explicit updater was provided. The field is
// optional, so an empty result remains SQL NULL.
func (m *MModel) ResolveUpdatedBy_(updatedBy sql.NullString, authenticatedUser string) (sql.NullString, error) {
	if m == nil {
		return sql.NullString{}, errors.New("model is nil")
	}
	for _, field := range m.GetFields() {
		if strings.EqualFold(fieldColumnName(field), "updated_by") {
			if updatedBy.Valid {
				updatedBy.String = strings.TrimSpace(updatedBy.String)
				return updatedBy, nil
			}
			authenticatedUser = strings.TrimSpace(authenticatedUser)
			return sql.NullString{String: authenticatedUser, Valid: authenticatedUser != ""}, nil
		}
	}
	return sql.NullString{}, fmt.Errorf("model %q tidak memiliki field updated_by", m.Name)
}

// ValidateID_ validates the positive integer identity used by a model.
func (m *MModel) ValidateID_(id int64, entityName string) error {
	if m == nil {
		return errors.New("model is nil")
	}
	entityName = strings.ToLower(strings.TrimSpace(entityName))
	if entityName == "" {
		entityName = strings.ToLower(strings.TrimSpace(m.Name))
	}
	if entityName == "" {
		entityName = "data"
	}
	if id <= 0 {
		return fmt.Errorf("id %s tidak valid", entityName)
	}
	return nil
}

// ValidateCreate_ checks a concrete model value against the metadata.
// Values are indexed by their db tag so metadata and Go field names can differ.
func (m *MModel) ValidateCreate_(value any) error {
	return m.validateValues(value, true)
}

// ValidateUpdate_ validates fields enabled for update on the concrete value.
func (m *MModel) ValidateUpdate_(value any) error {
	return m.validateValues(value, false)
}

// Create_ inserts a model with an integer identity and returns that key.
// UUID primary keys should use CreateGUID_ or CreateValue_.
func (m *MModel) Create_(value any) (int64, error) {
	if m.hasGUIDPrimaryKey() {
		return 0, errors.New("model uses a UUID primary key; use CreateGUID_ or CreateValue_")
	}
	created, err := m.CreateValue_(value)
	if err != nil {
		return 0, err
	}
	id, ok := created.(int64)
	if !ok {
		return 0, fmt.Errorf("generated primary key has type %T, expected int64", created)
	}
	return id, nil
}

// CreateGUID_ inserts a model with a UUID primary key and returns its UUID.
func (m *MModel) CreateGUID_(value any) (string, error) {
	if !m.hasGUIDPrimaryKey() {
		return "", errors.New("model does not use a UUID primary key")
	}
	created, err := m.CreateValue_(value)
	if err != nil {
		return "", err
	}
	id, ok := created.(string)
	if !ok {
		return "", fmt.Errorf("generated primary key has type %T, expected UUID string", created)
	}
	return id, nil
}

// CreateValue_ inserts a model and returns the primary key in its database type.
func (m *MModel) CreateValue_(value any) (any, error) {
	if m == nil {
		return nil, errors.New("model is nil")
	}
	if m.DB == nil {
		return nil, errors.New("model database is nil")
	}
	if err := m.ValidateCreate_(value); err != nil {
		return nil, err
	}
	values, err := modelValuesByDBTag(value)
	if err != nil {
		return nil, err
	}

	var primaryKey string
	columns := make([]string, 0, len(m.Fields))
	placeholders := make([]string, 0, len(m.Fields))
	args := make([]any, 0, len(m.Fields))
	for _, field := range m.GetFields() {
		name := fieldColumnName(field)
		if field.GetIsPrimaryKey() {
			primaryKey = name
		}
		if field.GetIsIdentity() || field.GetIsNotCreateField() {
			continue
		}
		fieldValue, exists := values[name]
		if !exists {
			continue
		}
		if text, ok := fieldValue.(string); ok {
			fieldValue = strings.TrimSpace(text)
		}
		if field.GetIsPrimaryKey() && !field.GetIsIdentity() && isZeroValue(fieldValue) {
			continue // Let the database generate the UUID primary key.
		}
		if isZeroValue(fieldValue) && !isExplicitBooleanFalse(fieldValue) && defaultForField(field) != "" {
			continue // Let the database apply the declared default.
		}
		columns = append(columns, fmt.Sprintf("%q", name))
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(args)+1))
		if stringValues, ok := fieldValue.([]string); ok {
			fieldValue = pq.Array(stringValues)
		}
		args = append(args, fieldValue)
	}
	if primaryKey == "" {
		return nil, errors.New("model has no primary key field")
	}
	tableName := strings.TrimSpace(m.TableName)
	query := ""
	if len(columns) == 0 {
		query = fmt.Sprintf("INSERT INTO %q DEFAULT VALUES RETURNING %q", tableName, primaryKey)
	} else {
		query = fmt.Sprintf(
			"INSERT INTO %q (%s) VALUES (%s) RETURNING %q",
			tableName, strings.Join(columns, ", "), strings.Join(placeholders, ", "), primaryKey,
		)
	}
	var id any
	if err := m.DB.QueryRowxContext(m.ContextOrBackground(), query, args...).Scan(&id); err != nil {
		return nil, fmt.Errorf("insert into %q: %w", m.TableName, err)
	}
	if m.hasGUIDPrimaryKey() {
		if rawID, ok := id.([]byte); ok {
			id = string(rawID)
		}
	}
	return id, nil
}

func (m *MModel) hasGUIDPrimaryKey() bool {
	if m == nil {
		return false
	}
	for _, field := range m.GetFields() {
		if field.GetIsPrimaryKey() {
			return !field.GetIsIdentity()
		}
	}
	return false
}

// Update_ updates fields enabled for update and scopes the statement by the
// model primary key. Additional key columns can be passed for tenant scoping.
func (m *MModel) Update_(value any, keyFields ...string) error {
	if m == nil {
		return errors.New("model is nil")
	}
	if m.DB == nil {
		return errors.New("model database is nil")
	}
	if err := m.ValidateUpdate_(value); err != nil {
		return err
	}
	values, err := modelValuesByDBTag(value)
	if err != nil {
		return err
	}

	fieldsByColumn := make(map[string]interfaces.IMField, len(m.Fields))
	if len(keyFields) == 0 {
		for _, field := range m.GetFields() {
			if field.GetIsPrimaryKey() {
				keyFields = append(keyFields, fieldColumnName(field))
				break
			}
		}
	}
	if len(keyFields) == 0 {
		return errors.New("model has no primary key field")
	}
	for _, field := range m.GetFields() {
		fieldsByColumn[fieldColumnName(field)] = field
	}

	keys := make(map[string]struct{}, len(keyFields))
	where := make([]string, 0, len(keyFields))
	args := make([]any, 0, len(m.Fields))
	for _, key := range keyFields {
		key = strings.TrimSpace(key)
		_, exists := fieldsByColumn[key]
		if !exists {
			return fmt.Errorf("unknown update key field %q", key)
		}
		if _, duplicate := keys[key]; duplicate {
			return fmt.Errorf("duplicate update key field %q", key)
		}
		keys[key] = struct{}{}
		value, exists := values[key]
		if !exists || isZeroValue(value) {
			return fmt.Errorf("update key %q wajib diisi", key)
		}
		where = append(where, fmt.Sprintf("%q = $%d", key, len(args)+1))
		args = append(args, value)
	}

	sets := make([]string, 0, len(m.Fields))
	for _, field := range m.GetFields() {
		name := fieldColumnName(field)
		if !fieldEnabledForOperation(field, false) || field.GetIsPrimaryKey() {
			continue
		}
		if _, isKey := keys[name]; isKey {
			continue
		}
		fieldValue, exists := values[name]
		if !exists || (field.GetIsPassword() && isZeroValue(fieldValue)) {
			continue
		}
		if text, ok := fieldValue.(string); ok {
			fieldValue = strings.TrimSpace(text)
		}
		sets = append(sets, fmt.Sprintf("%q = $%d", name, len(args)+1))
		if stringValues, ok := fieldValue.([]string); ok {
			fieldValue = pq.Array(stringValues)
		}
		args = append(args, fieldValue)
	}
	if len(sets) == 0 {
		return errors.New("model has no values to update")
	}

	query := fmt.Sprintf("UPDATE %q SET %s WHERE %s", strings.TrimSpace(m.TableName), strings.Join(sets, ", "), strings.Join(where, " AND "))
	result, err := m.DB.ExecContext(m.ContextOrBackground(), query, args...)
	if err != nil {
		return fmt.Errorf("update %q: %w", m.TableName, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read rows affected for update %q: %w", m.TableName, err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Delete_ deletes a record by its primary key. Additional key columns can be
// supplied to scope the deletion, for example by account_id in a tenant model.
func (m *MModel) Delete_(value any, keyFields ...string) error {
	if m == nil {
		return errors.New("model is nil")
	}
	if m.DB == nil {
		return errors.New("model database is nil")
	}
	if err := validateModelMetadata(m); err != nil {
		return err
	}
	values, err := modelValuesByDBTag(value)
	if err != nil {
		return err
	}

	fieldsByColumn := make(map[string]interfaces.IMField, len(m.Fields))
	if len(keyFields) == 0 {
		for _, field := range m.GetFields() {
			if field.GetIsPrimaryKey() {
				keyFields = append(keyFields, fieldColumnName(field))
				break
			}
		}
	}
	if len(keyFields) == 0 {
		return errors.New("model has no primary key field")
	}
	for _, field := range m.GetFields() {
		fieldsByColumn[fieldColumnName(field)] = field
	}

	keys := make(map[string]struct{}, len(keyFields))
	where := make([]string, 0, len(keyFields))
	args := make([]any, 0, len(keyFields))
	for _, key := range keyFields {
		key = strings.TrimSpace(key)
		if _, exists := fieldsByColumn[key]; !exists {
			return fmt.Errorf("unknown delete key field %q", key)
		}
		if _, duplicate := keys[key]; duplicate {
			return fmt.Errorf("duplicate delete key field %q", key)
		}
		keys[key] = struct{}{}
		value, exists := values[key]
		if !exists || isZeroValue(value) {
			return fmt.Errorf("delete key %q wajib diisi", key)
		}
		where = append(where, fmt.Sprintf("%q = $%d", key, len(args)+1))
		args = append(args, value)
	}

	query := fmt.Sprintf("DELETE FROM %q WHERE %s", strings.TrimSpace(m.TableName), strings.Join(where, " AND "))
	result, err := m.DB.ExecContext(m.ContextOrBackground(), query, args...)
	if err != nil {
		return fmt.Errorf("delete from %q: %w", m.TableName, err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read rows affected for delete %q: %w", m.TableName, err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// FindByID_ loads one row into destination using the primary key by default.
// Additional key columns can be passed for tenant scoping.
func (m *MModel) FindByID_(destination any, keyFields ...string) error {
	if m == nil {
		return errors.New("model is nil")
	}
	if m.DB == nil {
		return errors.New("model database is nil")
	}
	if err := validateModelMetadata(m); err != nil {
		return err
	}
	values, err := modelValuesByDBTag(destination)
	if err != nil {
		return fmt.Errorf("find destination: %w", err)
	}

	fieldsByColumn := make(map[string]interfaces.IMField, len(m.Fields))
	if len(keyFields) == 0 {
		for _, field := range m.GetFields() {
			if field.GetIsPrimaryKey() {
				keyFields = append(keyFields, fieldColumnName(field))
				break
			}
		}
	}
	if len(keyFields) == 0 {
		return errors.New("model has no primary key field")
	}
	for _, field := range m.GetFields() {
		fieldsByColumn[fieldColumnName(field)] = field
	}

	keys := make(map[string]struct{}, len(keyFields))
	where := make([]string, 0, len(keyFields))
	args := make([]any, 0, len(keyFields))
	for _, key := range keyFields {
		key = strings.TrimSpace(key)
		if _, exists := fieldsByColumn[key]; !exists {
			return fmt.Errorf("unknown find key field %q", key)
		}
		if _, duplicate := keys[key]; duplicate {
			return fmt.Errorf("duplicate find key field %q", key)
		}
		keys[key] = struct{}{}
		value, exists := values[key]
		if !exists || isZeroValue(value) {
			return fmt.Errorf("find key %q wajib diisi", key)
		}
		where = append(where, fmt.Sprintf("%q = $%d", key, len(args)+1))
		args = append(args, value)
	}

	selectFields := m.selectFields()
	if len(selectFields) == 0 {
		return errors.New("model has no selectable fields")
	}

	query := fmt.Sprintf("SELECT %s FROM %q WHERE %s", strings.Join(selectFields, ", "), strings.TrimSpace(m.TableName), strings.Join(where, " AND "))
	if err := m.DB.Unsafe().GetContext(m.ContextOrBackground(), destination, query, args...); err != nil {
		return fmt.Errorf("find row in %q: %w", m.TableName, err)
	}
	return nil
}

// FindByFields loads one row matching all explicitly named filter fields.
// Filter values and the destination are read from the same struct value.
func (m *MModel) FindByFields(destination any, filterFields ...string) error {
	if m == nil {
		return errors.New("model is nil")
	}
	if m.DB == nil {
		return errors.New("model database is nil")
	}
	if err := validateModelMetadata(m); err != nil {
		return err
	}
	if len(filterFields) == 0 {
		return errors.New("at least one find field is required")
	}
	values, err := modelValuesByDBTag(destination)
	if err != nil {
		return fmt.Errorf("find destination: %w", err)
	}

	fieldsByColumn := make(map[string]interfaces.IMField, len(m.Fields))
	for _, field := range m.GetFields() {
		fieldsByColumn[fieldColumnName(field)] = field
	}

	filters := make(map[string]struct{}, len(filterFields))
	where := make([]string, 0, len(filterFields))
	args := make([]any, 0, len(filterFields))
	for _, name := range filterFields {
		name = strings.TrimSpace(name)
		if _, exists := fieldsByColumn[name]; !exists {
			return fmt.Errorf("unknown find field %q", name)
		}
		if _, duplicate := filters[name]; duplicate {
			return fmt.Errorf("duplicate find field %q", name)
		}
		filters[name] = struct{}{}
		value, exists := values[name]
		if !exists || isZeroValue(value) {
			return fmt.Errorf("find field %q wajib diisi", name)
		}
		if text, ok := value.(string); ok {
			value = strings.TrimSpace(text)
		}
		where = append(where, fmt.Sprintf("%q = $%d", name, len(args)+1))
		if stringValues, ok := value.([]string); ok {
			value = pq.Array(stringValues)
		}
		args = append(args, value)
	}

	selectFields := m.selectFields()
	if len(selectFields) == 0 {
		return errors.New("model has no selectable fields")
	}
	query := fmt.Sprintf("SELECT %s FROM %q WHERE %s", strings.Join(selectFields, ", "), strings.TrimSpace(m.TableName), strings.Join(where, " AND "))
	if err := m.DB.Unsafe().GetContext(m.ContextOrBackground(), destination, query, args...); err != nil {
		return fmt.Errorf("find row in %q: %w", m.TableName, err)
	}
	return nil
}

// FindByAnyField loads one row where any of the supplied field/value pairs
// match. Field names are checked against metadata; values remain parameters.
func (m *MModel) FindByAnyField(destination any, fieldValues ...any) error {
	if m == nil {
		return errors.New("model is nil")
	}
	if m.DB == nil {
		return errors.New("model database is nil")
	}
	if err := validateModelMetadata(m); err != nil {
		return err
	}
	if len(fieldValues) == 0 || len(fieldValues)%2 != 0 {
		return errors.New("field/value pairs are required")
	}
	fieldsByColumn := make(map[string]interfaces.IMField, len(m.Fields))
	for _, field := range m.GetFields() {
		fieldsByColumn[fieldColumnName(field)] = field
	}
	conditions := make([]string, 0, len(fieldValues)/2)
	args := make([]any, 0, len(fieldValues)/2)
	seen := make(map[string]struct{}, len(fieldValues)/2)
	for i := 0; i < len(fieldValues); i += 2 {
		name, ok := fieldValues[i].(string)
		if !ok {
			return errors.New("find field name must be a string")
		}
		name = strings.TrimSpace(name)
		if _, exists := fieldsByColumn[name]; !exists {
			return fmt.Errorf("unknown find field %q", name)
		}
		if _, duplicate := seen[name]; duplicate {
			return fmt.Errorf("duplicate find field %q", name)
		}
		seen[name] = struct{}{}
		value := fieldValues[i+1]
		if value == nil {
			return fmt.Errorf("find field %q wajib diisi", name)
		}
		args = append(args, value)
		if isTextField(fieldsByColumn[name]) {
			conditions = append(conditions, fmt.Sprintf("LOWER(%q) = LOWER($%d)", name, len(args)))
		} else {
			conditions = append(conditions, fmt.Sprintf("%q = $%d", name, len(args)))
		}
	}
	selectFields := m.selectFields()
	if len(selectFields) == 0 {
		return errors.New("model has no selectable fields")
	}
	query := fmt.Sprintf("SELECT %s FROM %q WHERE (%s)", strings.Join(selectFields, ", "), strings.TrimSpace(m.TableName), strings.Join(conditions, " OR "))
	if err := m.DB.Unsafe().GetContext(m.ContextOrBackground(), destination, query, args...); err != nil {
		return fmt.Errorf("find row in %q: %w", m.TableName, err)
	}
	return nil
}

// ExistsByField reports whether a value already exists, optionally excluding
// the record identified by excludeField/excludeValue (for update validation).
func (m *MModel) ExistsByField(fieldName string, value any, excludeField string, excludeValue any) (bool, error) {
	if m == nil || m.DB == nil {
		return false, errors.New("model database is nil")
	}
	if err := validateModelMetadata(m); err != nil {
		return false, err
	}
	fields := make(map[string]interfaces.IMField, len(m.Fields))
	for _, field := range m.GetFields() {
		fields[fieldColumnName(field)] = field
	}
	fieldName = strings.TrimSpace(fieldName)
	field, exists := fields[fieldName]
	if !exists {
		return false, fmt.Errorf("unknown unique field %q", fieldName)
	}
	if !isTextField(field) {
		return false, fmt.Errorf("unique field %q must be text", fieldName)
	}
	if value == nil || isZeroValue(value) {
		return false, fmt.Errorf("unique field %q wajib diisi", fieldName)
	}
	conditions := []string{fmt.Sprintf("LOWER(%q) = LOWER($1)", fieldName)}
	args := []any{value}
	if strings.TrimSpace(excludeField) != "" {
		excludeField = strings.TrimSpace(excludeField)
		if _, exists := fields[excludeField]; !exists {
			return false, fmt.Errorf("unknown exclusion field %q", excludeField)
		}
		args = append(args, excludeValue)
		conditions = append(conditions, fmt.Sprintf("%q <> $%d", excludeField, len(args)))
	}
	query := fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM %q WHERE %s)", strings.TrimSpace(m.TableName), strings.Join(conditions, " AND "))
	var found bool
	if err := m.DB.GetContext(m.ContextOrBackground(), &found, query, args...); err != nil {
		return false, fmt.Errorf("check unique %q in %q: %w", fieldName, m.TableName, err)
	}
	return found, nil
}

// FindList loads rows into destination using structured filters, custom where
// conditions, full-text-style substring search, sorting, and one-based paging.
func (m *MModel) FindList(destination any, options MListOptions) error {
	return m.findList(destination, options, false)
}

func (m *MModel) findList(destination any, options MListOptions, allowUnbounded bool) error {
	if m == nil {
		return errors.New("model is nil")
	}
	if m.DB == nil {
		return errors.New("model database is nil")
	}
	if err := validateModelMetadata(m); err != nil {
		return err
	}
	if options.Page < 0 {
		return errors.New("list page cannot be negative")
	}
	page := options.Page
	if page == 0 {
		page = 1
	}
	limit := options.Limit
	if limit <= 0 && allowUnbounded {
		limit = 0
	} else if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	if limit > 0 && page-1 > int(^uint(0)>>1)/limit {
		return errors.New("list page is too large")
	}
	offset := 0
	if limit > 0 {
		offset = (page - 1) * limit
	}

	fieldsByColumn := make(map[string]interfaces.IMField, len(m.Fields))
	for _, field := range m.GetFields() {
		fieldsByColumn[fieldColumnName(field)] = field
	}
	args := make([]any, 0, len(options.Filters)+len(options.Where)+len(options.SearchFields)+2)
	where := make([]string, 0, len(options.Filters)+len(options.Where)+1)
	for name, value := range options.Filters {
		name = strings.TrimSpace(name)
		if _, exists := fieldsByColumn[name]; !exists {
			return fmt.Errorf("unknown list filter field %q", name)
		}
		if text, ok := value.(string); ok {
			value = strings.TrimSpace(text)
		}
		if value == nil {
			where = append(where, fmt.Sprintf("%q IS NULL", name))
			continue
		}
		args = append(args, normalizeListValue(value))
		where = append(where, fmt.Sprintf("%q = $%d", name, len(args)))
	}
	for _, condition := range options.Where {
		expression, values, err := m.buildListCondition(condition, fieldsByColumn)
		if err != nil {
			return err
		}
		for _, value := range values {
			args = append(args, normalizeListValue(value))
		}
		// Conditions use local argument positions; shift placeholders to account
		// for filters and earlier custom conditions already in args.
		where = append(where, shiftSQLPlaceholders(expression, len(args)-len(values)))
	}

	search := strings.TrimSpace(options.Search)
	if search != "" {
		if len(options.SearchFields) == 0 {
			return errors.New("at least one search field is required")
		}
		searchConditions := make([]string, 0, len(options.SearchFields))
		seen := make(map[string]struct{}, len(options.SearchFields))
		for _, name := range options.SearchFields {
			name = strings.TrimSpace(name)
			field, exists := fieldsByColumn[name]
			if !exists {
				return fmt.Errorf("unknown list search field %q", name)
			}
			if _, duplicate := seen[name]; duplicate {
				return fmt.Errorf("duplicate list search field %q", name)
			}
			seen[name] = struct{}{}
			if !isTextField(field) {
				return fmt.Errorf("list search field %q must be a text field", name)
			}
			args = append(args, search)
			searchConditions = append(searchConditions, fmt.Sprintf("%q ILIKE '%%' || $%d || '%%'", name, len(args)))
		}
		where = append(where, "("+strings.Join(searchConditions, " OR ")+")")
	}

	selectFields := m.selectFields()
	if len(selectFields) == 0 {
		return errors.New("model has no selectable fields")
	}
	query := fmt.Sprintf("SELECT %s FROM %q", strings.Join(selectFields, ", "), strings.TrimSpace(m.TableName))
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	if len(options.OrderFields) > 0 {
		orderParts := make([]string, 0, len(options.OrderFields))
		seen := make(map[string]struct{}, len(options.OrderFields))
		for _, order := range options.OrderFields {
			name := strings.TrimSpace(order.Field)
			if _, exists := fieldsByColumn[name]; !exists {
				return fmt.Errorf("unknown list order field %q", name)
			}
			if _, duplicate := seen[name]; duplicate {
				return fmt.Errorf("duplicate list order field %q", name)
			}
			seen[name] = struct{}{}
			direction := strings.ToUpper(strings.TrimSpace(order.Direction))
			if direction == "" {
				direction = "ASC"
			}
			if direction != "ASC" && direction != "DESC" {
				return fmt.Errorf("invalid list order direction %q", order.Direction)
			}
			orderParts = append(orderParts, fmt.Sprintf("%q %s", name, direction))
		}
		query += " ORDER BY " + strings.Join(orderParts, ", ")
	}
	if limit > 0 {
		args = append(args, limit, offset)
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	}
	if err := m.DB.Unsafe().SelectContext(m.ContextOrBackground(), destination, query, args...); err != nil {
		return fmt.Errorf("list rows from %q: %w", m.TableName, err)
	}
	return nil
}

// FindListOrdered is retained as a convenience for callers using the previous
// filter/search/sort arguments. A zero limit leaves the result unbounded.
func (m *MModel) FindListOrdered(destination any, filters map[string]any, search string, searchFields []string, orderBy, orderDirection string, limit int) error {
	if limit < 0 {
		return errors.New("list limit cannot be negative")
	}
	orders := make([]MOrderField, 0)
	if strings.TrimSpace(orderBy) != "" {
		for _, field := range strings.Split(orderBy, ",") {
			orders = append(orders, MOrderField{Field: strings.TrimSpace(field), Direction: orderDirection})
		}
	}
	if limit == 0 {
		// Preserve FindListOrdered's historical unbounded behavior.
		return m.findList(destination, MListOptions{Filters: filters, Search: search, SearchFields: searchFields, OrderFields: orders}, true)
	}
	return m.FindList(destination, MListOptions{Filters: filters, Search: search, SearchFields: searchFields, OrderFields: orders, Page: 1, Limit: limit})
}

func normalizeListValue(value any) any {
	if stringValues, ok := value.([]string); ok {
		return pq.Array(stringValues)
	}
	return value
}

func (m *MModel) buildListCondition(condition MWhere, fields map[string]interfaces.IMField) (string, []any, error) {
	name := strings.TrimSpace(condition.Field)
	if _, exists := fields[name]; !exists {
		return "", nil, fmt.Errorf("unknown custom where field %q", name)
	}
	operator := strings.ToUpper(strings.TrimSpace(condition.Operator))
	switch operator {
	case "IS NULL", "IS NOT NULL":
		return fmt.Sprintf("%q %s", name, operator), nil, nil
	case "=", "<>", ">", ">=", "<", "<=", "LIKE", "ILIKE":
		if condition.Value == nil {
			return "", nil, fmt.Errorf("custom where %s requires a value", operator)
		}
		if (operator == "LIKE" || operator == "ILIKE") && !isTextField(fields[name]) {
			return "", nil, fmt.Errorf("custom where %s requires a text field", operator)
		}
		return fmt.Sprintf("%q %s $1", name, operator), []any{condition.Value}, nil
	case "IN", "NOT IN":
		values := reflect.ValueOf(condition.Value)
		if !values.IsValid() || (values.Kind() != reflect.Array && values.Kind() != reflect.Slice) || values.Len() == 0 {
			return "", nil, errors.New("custom where IN requires a non-empty slice")
		}
		placeholders := make([]string, 0, values.Len())
		args := make([]any, 0, values.Len())
		for i := 0; i < values.Len(); i++ {
			args = append(args, values.Index(i).Interface())
			placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		}
		return fmt.Sprintf("%q %s (%s)", name, operator, strings.Join(placeholders, ", ")), args, nil
	default:
		return "", nil, fmt.Errorf("unsupported custom where operator %q", condition.Operator)
	}
}

func shiftSQLPlaceholders(expression string, offset int) string {
	if offset == 0 {
		return expression
	}
	return regexp.MustCompile(`\$(\d+)`).ReplaceAllStringFunc(expression, func(match string) string {
		index, _ := strconv.Atoi(strings.TrimPrefix(match, "$"))
		return fmt.Sprintf("$%d", index+offset)
	})
}

func isTextField(field interfaces.IMField) bool {
	typeName := strings.ToUpper(strings.TrimSpace(databaseTypeForField(field)))
	if index := strings.IndexByte(typeName, '('); index >= 0 {
		typeName = typeName[:index]
	}
	switch typeName {
	case "TEXT", "VARCHAR", "CHARACTER VARYING", "CHAR", "CHARACTER", "STRING":
		return true
	default:
		return false
	}
}

func (m *MModel) selectFields() []string {
	fields := make([]string, 0, len(m.Fields))
	for _, field := range m.GetFields() {
		if field.GetIsNotSelectField() {
			continue
		}
		selectField := strings.TrimSpace(field.GetSelectField())
		if selectField == "" {
			selectField = fmt.Sprintf("%q", fieldColumnName(field))
		}
		fields = append(fields, selectField)
	}
	return fields
}

func (m *MModel) validateValues(value any, create bool) error {
	if m == nil {
		return errors.New("model is nil")
	}
	if err := validateModelMetadata(m); err != nil {
		return err
	}
	values, err := modelValuesByDBTag(value)
	if err != nil {
		return err
	}
	for _, field := range m.GetFields() {
		if !fieldEnabledForOperation(field, create) {
			continue
		}
		key := fieldColumnName(field)
		fieldValue, ok := values[key]
		if text, isString := fieldValue.(string); isString {
			fieldValue = strings.TrimSpace(text)
			values[key] = fieldValue
		}
		missing := isMissingRequiredValue(fieldValue, ok)
		optionalPassword := field.GetIsPassword() && !create && missing
		if ok && isNullValue(fieldValue) && !field.GetIsNullable() && !optionalPassword {
			return fmt.Errorf("%s tidak boleh null", key)
		}
		if create && !field.GetIsNullable() && !field.GetIsIdentity() && defaultForField(field) == "" && (!ok || isNullValue(fieldValue)) {
			return fmt.Errorf("%s tidak boleh null", key)
		}
		if create && field.GetRequired() && !field.GetIsIdentity() && defaultForField(field) == "" && missing {
			return fmt.Errorf("%s wajib diisi", key)
		}
		if !ok || isNullValue(fieldValue) {
			continue
		}
		fieldValue = unwrapNullableValue(fieldValue)
		if text, isString := fieldValue.(string); isString {
			fieldValue = strings.TrimSpace(text)
		}
		if !matchesFieldType(fieldValue, field) {
			return fmt.Errorf("tipe nilai %s tidak sesuai dengan %s", key, field.GetType())
		}
		if optionalPassword || isZeroValue(fieldValue) {
			continue
		}
		if field.GetIsPrimaryKey() && !field.GetIsIdentity() {
			text, ok := fieldValue.(string)
			if !ok || !uuidPattern.MatchString(strings.TrimSpace(text)) {
				return fmt.Errorf("%s harus berupa UUID yang valid", key)
			}
		}
		text, isString := fieldValue.(string)
		if field.GetIsEmail() {
			if !isString || !isValidEmail(text) {
				return fmt.Errorf("format %s tidak valid", key)
			}
		}
		if field.GetIsPhone() {
			if !isString || !isValidPhone(text) {
				return fmt.Errorf("format %s tidak valid", key)
			}
		}
		if field.GetMinLength() > 0 || field.GetMaxLength() > 0 || field.GetExactLength() > 0 || strings.TrimSpace(field.GetPattern()) != "" {
			if !isString {
				return fmt.Errorf("%s harus berupa teks", key)
			}
			length := utf8.RuneCountInString(text)
			if field.GetMinLength() > 0 && length < field.GetMinLength() {
				return fmt.Errorf("%s minimal %d karakter", key, field.GetMinLength())
			}
			if field.GetMaxLength() > 0 && length > field.GetMaxLength() {
				return fmt.Errorf("%s maksimal %d karakter", key, field.GetMaxLength())
			}
			if field.GetExactLength() > 0 && length != field.GetExactLength() {
				return fmt.Errorf("%s harus tepat %d karakter", key, field.GetExactLength())
			}
			if pattern := strings.TrimSpace(field.GetPattern()); pattern != "" {
				matcher, err := regexp.Compile(pattern)
				if err != nil {
					return fmt.Errorf("pattern untuk %s tidak valid: %w", key, err)
				}
				if !matcher.MatchString(text) {
					return fmt.Errorf("format %s tidak valid", key)
				}
			}
		}
	}
	return nil
}

func isMissingRequiredValue(value any, exists bool) bool {
	if !exists || value == nil || isNullValue(value) {
		return true
	}
	reflected := reflect.ValueOf(value)
	for reflected.IsValid() && (reflected.Kind() == reflect.Pointer || reflected.Kind() == reflect.Interface) {
		if reflected.IsNil() {
			return true
		}
		reflected = reflected.Elem()
	}
	if !reflected.IsValid() {
		return true
	}
	return reflected.Kind() == reflect.String && strings.TrimSpace(reflected.String()) == ""
}

func isNullValue(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	for reflected.IsValid() && (reflected.Kind() == reflect.Pointer || reflected.Kind() == reflect.Interface) {
		if reflected.IsNil() {
			return true
		}
		reflected = reflected.Elem()
	}
	if !reflected.IsValid() {
		return true
	}
	if (reflected.Kind() == reflect.Slice || reflected.Kind() == reflect.Map) && reflected.IsNil() {
		return true
	}
	if reflected.Kind() == reflect.Struct && strings.HasPrefix(reflected.Type().Name(), "Null") {
		valid := reflected.FieldByName("Valid")
		return valid.IsValid() && valid.Kind() == reflect.Bool && !valid.Bool()
	}
	return false
}

func unwrapNullableValue(value any) any {
	reflected := reflect.ValueOf(value)
	for reflected.IsValid() && (reflected.Kind() == reflect.Pointer || reflected.Kind() == reflect.Interface) {
		if reflected.IsNil() {
			return nil
		}
		reflected = reflected.Elem()
	}
	if !reflected.IsValid() || reflected.Kind() != reflect.Struct || !strings.HasPrefix(reflected.Type().Name(), "Null") {
		return value
	}
	for _, name := range []string{"String", "Bool", "Int16", "Int32", "Int64", "Float64", "Time"} {
		field := reflected.FieldByName(name)
		if field.IsValid() && field.CanInterface() {
			return field.Interface()
		}
	}
	return value
}

func matchesFieldType(value any, field interfaces.IMField) bool {
	if isNullValue(value) {
		return true
	}
	reflected := reflect.ValueOf(value)
	for reflected.Kind() == reflect.Pointer || reflected.Kind() == reflect.Interface {
		if reflected.IsNil() {
			return true
		}
		reflected = reflected.Elem()
	}
	if reflected.Kind() == reflect.Struct && strings.HasPrefix(reflected.Type().Name(), "Null") {
		for _, valueField := range []string{"String", "Bool", "Int16", "Int32", "Int64", "Float64", "Time"} {
			underlying := reflected.FieldByName(valueField)
			if underlying.IsValid() {
				reflected = underlying
				break
			}
		}
	}
	kind := reflected.Kind()
	dbType := strings.ToUpper(strings.TrimSpace(databaseTypeForField(field)))
	isArray := strings.HasSuffix(dbType, "[]")
	if isArray {
		dbType = strings.TrimSuffix(dbType, "[]")
	}
	if index := strings.IndexByte(dbType, '('); index >= 0 {
		dbType = dbType[:index]
	}
	if isArray {
		return kind == reflect.Slice || kind == reflect.Array
	}
	switch dbType {
	case "BOOL", "BOOLEAN":
		return kind == reflect.Bool
	case "SMALLINT", "INT2", "INTEGER", "INT", "INT4", "BIGINT", "INT8", "SERIAL", "BIGSERIAL":
		return kind >= reflect.Int && kind <= reflect.Int64 || kind >= reflect.Uint && kind <= reflect.Uint64
	case "REAL", "FLOAT", "FLOAT4", "FLOAT8", "DOUBLE PRECISION", "NUMERIC", "DECIMAL":
		return kind >= reflect.Int && kind <= reflect.Int64 || kind >= reflect.Uint && kind <= reflect.Uint64 || kind == reflect.Float32 || kind == reflect.Float64
	case "TEXT", "VARCHAR", "CHARACTER VARYING", "CHAR", "CHARACTER":
		return kind == reflect.String
	case "UUID":
		return kind == reflect.String
	case "TIMESTAMP", "TIMESTAMPTZ", "DATE", "TIME":
		return strings.Contains(strings.ToLower(reflected.Type().String()), "time") || strings.Contains(strings.ToLower(reflected.Type().Name()), "nulltime")
	default:
		return true
	}
}

func isValidEmail(value string) bool {
	address, err := mail.ParseAddress(strings.TrimSpace(value))
	return err == nil && address.Address == strings.TrimSpace(value) && strings.Contains(address.Address, "@")
}

func isValidPhone(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	digits := 0
	for i, r := range value {
		switch {
		case r >= '0' && r <= '9':
			digits++
		case r == '+' && i == 0:
		case r == ' ' || r == '-' || r == '(' || r == ')' || r == '.':
		default:
			return false
		}
	}
	return digits >= 6 && digits <= 20
}

func fieldEnabledForOperation(field interfaces.IMField, create bool) bool {
	if create {
		return !field.GetIsNotCreateField()
	}
	return !field.GetIsNotUpdateField()
}

func fieldColumnName(field interfaces.IMField) string {
	if name := strings.TrimSpace(field.GetDbTag()); name != "" {
		return name
	}
	jsonTag := strings.TrimSpace(field.GetJsonTag())
	if jsonTag != "" && jsonTag != "-" {
		return jsonTag
	}
	return ""
}

func modelValuesByDBTag(value any) (map[string]any, error) {
	if value == nil {
		return nil, errors.New("model value is nil")
	}
	reflected := reflect.ValueOf(value)
	for reflected.Kind() == reflect.Pointer || reflected.Kind() == reflect.Interface {
		if reflected.IsNil() {
			return nil, errors.New("model value is nil")
		}
		reflected = reflected.Elem()
	}
	if reflected.Kind() != reflect.Struct {
		return nil, errors.New("model value must be a struct or pointer to struct")
	}

	values := make(map[string]any)
	modelType := reflected.Type()
	for i := 0; i < reflected.NumField(); i++ {
		structField := modelType.Field(i)
		fieldValue := reflected.Field(i)
		if structField.Anonymous {
			for fieldValue.Kind() == reflect.Pointer || fieldValue.Kind() == reflect.Interface {
				if fieldValue.IsNil() {
					break
				}
				fieldValue = fieldValue.Elem()
			}
			if fieldValue.IsValid() && fieldValue.Kind() == reflect.Struct && fieldValue.Type() != reflect.TypeOf(MModel{}) {
				nestedValues, err := modelValuesByDBTag(fieldValue.Interface())
				if err != nil {
					return nil, err
				}
				for key, nestedValue := range nestedValues {
					values[key] = nestedValue
				}
			}
			continue
		}
		dbTag := strings.SplitN(structField.Tag.Get("db"), ",", 2)[0]
		if dbTag == "-" {
			continue
		}
		if dbTag == "" {
			dbTag = structField.Name
		}
		if fieldValue.CanInterface() {
			values[dbTag] = fieldValue.Interface()
		}
	}
	return values, nil
}

func isZeroValue(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	return reflected.IsZero() || (reflected.Kind() == reflect.String && strings.TrimSpace(reflected.String()) == "")
}

func isExplicitBooleanFalse(value any) bool {
	reflected := reflect.ValueOf(value)
	return reflected.IsValid() && reflected.Kind() == reflect.Bool && !reflected.Bool()
}

// Migrate creates the model's table using its field metadata.
func (m *MModel) Migrate() error {
	if m == nil {
		return errors.New("model is nil")
	}
	if err := validateModelMetadata(m); err != nil {
		return err
	}
	migrator := NewMMigrator(m, m.DB)
	migrator.SetContext(m.ContextOrBackground())
	return migrator.Migrate()
}

func (m *MModel) GetName() string {
	return m.Name
}

func (m *MModel) GetTableName() string {
	return m.TableName
}

func (m *MModel) GetFields() []interfaces.IMField {
	fields := make([]interfaces.IMField, len(m.Fields))
	for i, f := range m.Fields {
		fields[i] = f
	}
	return fields
}
