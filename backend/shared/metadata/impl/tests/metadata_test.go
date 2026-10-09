package tests

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"backend/shared/metadata/impl"
	"backend/shared/metadata/interfaces"
	"backend/shared/mmodels"

	"github.com/lib/pq"
)

type recordingExecutor struct {
	query string
	err   error
	calls int
}

func (e *recordingExecutor) Exec(query string, _ ...any) (sql.Result, error) {
	e.query = query
	e.calls++
	return nil, e.err
}

func testModel(table string, fields ...interfaces.IMField) *impl.MModel {
	return &impl.MModel{Name: "Example", TableName: table, Fields: fields}
}

func TestMFieldGetters(t *testing.T) {
	field := impl.MField{
		Type: "int64", DbType: "BIGINT", JsonTag: "id",
		DbTag: "id", IsNotSelectField: true, SelectField: "id AS ID",
		IsNotCreateField: true, IsNotUpdateField: true,
		IsPrimaryKey: true, IsIdentity: true, IsNullable: false, Default: "1",
		MinLength: 2, MaxLength: 64, Pattern: `^ID$`, ExactLength: 2, Required: true, IsPassword: true,
	}

	if field.GetType() != "int64" || field.GetDbType() != "BIGINT" ||
		field.GetJsonTag() != "id" || field.GetDbTag() != "id" || !field.GetIsNotSelectField() || !field.GetIsNotCreateField() || !field.GetIsNotUpdateField() ||
		field.GetSelectField() != "id AS ID" || !field.GetIsPrimaryKey() || !field.GetIsIdentity() ||
		field.GetIsNullable() || field.GetDefault() != "1" || field.GetMinLength() != 2 || field.GetMaxLength() != 64 || field.GetPattern() != `^ID$` || field.GetExactLength() != 2 || !field.GetRequired() || !field.GetIsPassword() {
		t.Fatal("MField getters did not return their configured values")
	}
}

func TestMModelGettersAndCopiesFields(t *testing.T) {
	original := &impl.MField{DbTag: "id"}
	model := testModel("users", original)

	if model.GetName() != "Example" || model.GetTableName() != "users" {
		t.Fatal("MModel getters returned unexpected values")
	}
	fields := model.GetFields()
	if len(fields) != 1 || fields[0] != original {
		t.Fatal("GetFields returned unexpected fields")
	}
	fields[0] = &impl.MField{DbTag: "changed"}
	if model.Fields[0] != original {
		t.Fatal("changing returned slice should not replace the model's field")
	}
}

func TestMModelValidateCreate(t *testing.T) {
	tests := []struct {
		name    string
		model   *impl.MModel
		wantErr bool
	}{
		{name: "valid model", model: testModel("users", impl.MField{DbTag: "id", DbType: "BIGINT", IsPrimaryKey: true})},
		{name: "invalid table", model: testModel("users;drop", impl.MField{DbTag: "id", DbType: "BIGINT"}), wantErr: true},
		{name: "invalid column", model: testModel("users", impl.MField{DbTag: "bad-name", DbType: "BIGINT"}), wantErr: true},
		{name: "invalid database type", model: testModel("users", impl.MField{DbTag: "id", DbType: "BIGINT; DROP TABLE users"}), wantErr: true},
		{name: "empty fields", model: testModel("users"), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.ValidateCreate_(struct{}{})
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateCreate_() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewUserMModel(t *testing.T) {
	model := mmodels.NewUserMModel(context.Background(), nil)
	if model.GetName() != "User" || model.GetTableName() != "users" {
		t.Fatalf("unexpected user model identity: %q / %q", model.GetName(), model.GetTableName())
	}
	passwordFields := 0
	for _, field := range model.GetFields() {
		if field.GetIsPassword() {
			passwordFields++
		}
	}
	if passwordFields != 2 {
		t.Fatalf("password field count = %d, want 2", passwordFields)
	}
	model.UserEmail = "user@example.com"
	model.MobilePhone = "08123456789"
	model.UserPassword = "secret"
	model.UserRoles = pq.StringArray{"tester"}
	model.Username = "user"
	model.CreatedBy = "tester"
	if err := model.MModel.ValidateCreate_(model); err != nil {
		t.Fatalf("valid user values rejected by metadata validation: %v", err)
	}
	model.UserEmail = "not-an-email"
	if err := model.MModel.ValidateCreate_(model); err == nil {
		t.Fatal("ValidateCreate_() accepted an invalid email format")
	}

	model.UserEmail = " "
	if err := model.MModel.ValidateCreate_(model); err == nil {
		t.Fatal("ValidateCreate_() accepted an empty email")
	}

	model.UserEmail = "user@example.com"
	model.CreatedBy = ""
	if err := model.MModel.ValidateCreate_(model); err == nil {
		t.Fatal("ValidateCreate_() accepted an empty required created_by field")
	}
	model.CreatedBy = "tester"
	model.Username = strings.Repeat("u", 256)
	if err := model.MModel.ValidateCreate_(model); err == nil {
		t.Fatal("ValidateCreate_() accepted username longer than MaxLength")
	}
}

func TestMModelValidateCreateMinLengthAndPattern(t *testing.T) {
	model := testModel("labels", impl.MField{DbTag: "code", Type: "string", DbType: "VARCHAR", MinLength: 3, MaxLength: 8, Pattern: `^[A-Z]+$`})
	value := struct {
		Code string `db:"code"`
	}{Code: "AB"}
	if err := model.ValidateCreate_(value); err == nil {
		t.Fatal("ValidateCreate_() accepted value below MinLength")
	}
	value.Code = "abc"
	if err := model.ValidateCreate_(value); err == nil {
		t.Fatal("ValidateCreate_() accepted a value that does not match Pattern")
	}
	value.Code = "ABCD"
	if err := model.ValidateCreate_(value); err != nil {
		t.Fatalf("ValidateCreate_() rejected valid value: %v", err)
	}
}

func TestMModelValidationUsesCreateAndUpdateFlags(t *testing.T) {
	model := testModel("records",
		impl.MField{DbTag: "create_value", Type: "string", DbType: "TEXT", Required: true, IsNotUpdateField: true},
		impl.MField{DbTag: "update_value", Type: "string", DbType: "TEXT", Required: true, IsNotCreateField: true},
	)
	type record struct {
		CreateValue string `db:"create_value"`
		UpdateValue string `db:"update_value"`
	}

	if err := model.ValidateCreate_(record{CreateValue: "created"}); err != nil {
		t.Fatalf("create validation rejected create fields: %v", err)
	}
	if err := model.ValidateUpdate_(record{UpdateValue: "updated"}); err != nil {
		t.Fatalf("update validation rejected update fields: %v", err)
	}
	if err := model.ValidateCreate_(record{}); err == nil {
		t.Fatal("create validation accepted a missing required create field")
	}
	if err := model.ValidateUpdate_(record{}); err != nil {
		t.Fatalf("update validation should not require omitted fields: %v", err)
	}
}

func TestMMigratorCreatesTableFromMetadata(t *testing.T) {
	executor := &recordingExecutor{}
	model := testModel("users",
		impl.MField{DbTag: "id", Type: "int64", DbType: "BIGINT", IsPrimaryKey: true, IsIdentity: true},
		impl.MField{DbTag: "email", Type: "string", DbType: "VARCHAR", MaxLength: 255, IsNullable: true},
		impl.MField{DbTag: "active", Type: "bool", DbType: "BOOLEAN", Default: "true"},
	)

	if err := impl.NewMMigrator(model, executor).Migrate(); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	want := `CREATE TABLE IF NOT EXISTS "users" ("id" BIGINT GENERATED BY DEFAULT AS IDENTITY PRIMARY KEY NOT NULL, "email" VARCHAR(255), "active" BOOLEAN NOT NULL DEFAULT true)`
	if executor.calls != 1 || executor.query != want {
		t.Fatalf("Exec called %d times with query %q; want once with %q", executor.calls, executor.query, want)
	}
}

func TestMMigratorAppliesRequiredAndMaxLength(t *testing.T) {
	executor := &recordingExecutor{}
	model := testModel("profiles",
		impl.MField{DbTag: "name", DbType: "VARCHAR", MaxLength: 80, Required: true, IsNullable: true},
	)

	if err := impl.NewMMigrator(model, executor).Migrate(); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	want := `CREATE TABLE IF NOT EXISTS "profiles" ("name" VARCHAR(80) NOT NULL)`
	if executor.query != want {
		t.Fatalf("query = %q, want %q", executor.query, want)
	}
}

func TestMMigratorUsesMaxLengthForVarcharWithExistingLength(t *testing.T) {
	executor := &recordingExecutor{}
	model := testModel("profiles", impl.MField{DbTag: "name", DbType: "VARCHAR(32)", MaxLength: 80})
	if err := impl.NewMMigrator(model, executor).Migrate(); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	want := `CREATE TABLE IF NOT EXISTS "profiles" ("name" VARCHAR(80) NOT NULL)`
	if executor.query != want {
		t.Fatalf("query = %q, want %q", executor.query, want)
	}
}

func TestMMigratorFallsBackToGoType(t *testing.T) {
	executor := &recordingExecutor{}
	model := testModel("users", impl.MField{DbTag: "id", Type: "INTEGER"})

	if err := impl.NewMMigrator(model, executor).Migrate(); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	if !strings.Contains(executor.query, `"id" INTEGER NOT NULL`) {
		t.Fatalf("query did not use Type fallback: %q", executor.query)
	}
}

func TestMMigratorRejectsInvalidMetadataWithoutExecuting(t *testing.T) {
	tests := []struct {
		name  string
		model *impl.MModel
	}{
		{name: "table name", model: testModel("users; DROP TABLE users", impl.MField{DbTag: "id", DbType: "INTEGER"})},
		{name: "column name", model: testModel("users", impl.MField{DbTag: "bad-name", DbType: "INTEGER"})},
		{name: "SQL type", model: testModel("users", impl.MField{DbTag: "id", DbType: "INTEGER); DROP TABLE users;--"})},
		{name: "duplicate column", model: testModel("users", impl.MField{DbTag: "id", DbType: "INTEGER"}, impl.MField{DbTag: "id", DbType: "TEXT"})},
		{name: "multiple primary keys", model: testModel("users", impl.MField{DbTag: "id", DbType: "INTEGER", IsPrimaryKey: true}, impl.MField{DbTag: "other", DbType: "INTEGER", IsPrimaryKey: true})},
		{name: "unsupported identity type", model: testModel("users", impl.MField{DbTag: "id", DbType: "TEXT", IsIdentity: true})},
		{name: "no fields", model: testModel("users")},
		{name: "nil field", model: testModel("users", nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &recordingExecutor{}
			if err := impl.NewMMigrator(tt.model, executor).Migrate(); err == nil {
				t.Fatal("Migrate() error = nil, want validation error")
			}
			if executor.calls != 0 {
				t.Fatalf("Exec called %d times for invalid metadata", executor.calls)
			}
		})
	}
}

func TestMMigratorReturnsExecutorError(t *testing.T) {
	wantErr := errors.New("database unavailable")
	executor := &recordingExecutor{err: wantErr}
	migrator := impl.NewMMigrator(testModel("users", impl.MField{DbTag: "id", DbType: "INTEGER"}), executor)

	err := migrator.Migrate()
	if !errors.Is(err, wantErr) {
		t.Fatalf("Migrate() error = %v, want wrapped %v", err, wantErr)
	}
}

func TestMMigratorRejectsMissingDependencies(t *testing.T) {
	tests := []struct {
		name     string
		migrator *impl.MMigrator
	}{
		{name: "nil migrator"},
		{name: "nil model", migrator: impl.NewMMigrator(nil, &recordingExecutor{})},
		{name: "nil executor", migrator: impl.NewMMigrator(testModel("users", impl.MField{DbTag: "id", DbType: "INTEGER"}), nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.migrator.Migrate(); err == nil {
				t.Fatal("Migrate() error = nil, want dependency error")
			}
		})
	}
}

func TestMMigratorSetters(t *testing.T) {
	firstExecutor := &recordingExecutor{}
	secondExecutor := &recordingExecutor{}
	model := testModel("users", impl.MField{DbTag: "id", DbType: "INTEGER"})
	migrator := impl.NewMMigrator(nil, nil)
	migrator.SetModel(model)
	migrator.SetExecutor(secondExecutor)

	if err := migrator.Migrate(); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	if secondExecutor.calls != 1 || firstExecutor.calls != 0 {
		t.Fatal("SetExecutor did not set the executor")
	}
}
