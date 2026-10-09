package impl

// MField describes one database column. A primary key without identity is
// treated as a UUID primary key and defaults to gen_random_uuid().
type MField struct {
	Type    string
	DbType  string
	JsonTag string
	DbTag   string

	IsNotSelectField bool
	IsNotCreateField bool
	IsNotUpdateField bool
	SelectField      string

	IsPrimaryKey bool
	IsIdentity   bool
	IsNullable   bool
	Default      string
	MaxLength    int
	MinLength    int
	Pattern      string
	Required     bool
	IsPassword   bool
	ExactLength  int

	IsEmail bool
	IsPhone bool
}

func (f MField) GetType() string {
	if f.IsPrimaryKey && !f.IsIdentity {
		return "string"
	}
	return f.Type
}
func (f MField) GetDbType() string {
	if f.IsPrimaryKey && !f.IsIdentity {
		return "UUID"
	}
	return f.DbType
}
func (f MField) GetJsonTag() string        { return f.JsonTag }
func (f MField) GetDbTag() string          { return f.DbTag }
func (f MField) GetIsSelectField() bool    { return !f.IsNotSelectField }
func (f MField) GetIsNotSelectField() bool { return f.IsNotSelectField }
func (f MField) GetSelectField() string    { return f.SelectField }
func (f MField) GetIsNotCreateField() bool { return f.IsNotCreateField }
func (f MField) GetIsNotUpdateField() bool { return f.IsNotUpdateField }
func (f MField) GetIsPrimaryKey() bool     { return f.IsPrimaryKey }
func (f MField) GetIsIdentity() bool       { return f.IsIdentity }
func (f MField) GetIsNullable() bool       { return f.IsNullable }
func (f MField) GetDefault() string {
	if f.IsPrimaryKey && !f.IsIdentity && f.Default == "" {
		return "gen_random_uuid()"
	}
	return f.Default
}
func (f MField) GetMaxLength() int   { return f.MaxLength }
func (f MField) GetRequired() bool   { return f.Required }
func (f MField) GetIsPassword() bool { return f.IsPassword }
func (f MField) GetExactLength() int { return f.ExactLength }
func (f MField) GetMinLength() int   { return f.MinLength }
func (f MField) GetPattern() string  { return f.Pattern }
func (f MField) GetIsEmail() bool    { return f.IsEmail }
func (f MField) GetIsPhone() bool    { return f.IsPhone }
