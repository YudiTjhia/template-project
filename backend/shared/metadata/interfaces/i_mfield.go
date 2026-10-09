package interfaces

type IMField interface {
	GetType() string
	GetDbType() string
	GetJsonTag() string
	GetDbTag() string
	GetIsNotSelectField() bool
	GetSelectField() string
	GetIsNotCreateField() bool
	GetIsNotUpdateField() bool
	GetIsPrimaryKey() bool
	GetIsIdentity() bool
	GetIsNullable() bool
	GetDefault() string
	GetMinLength() int
	GetMaxLength() int
	GetPattern() string
	GetExactLength() int
	GetRequired() bool
	GetIsPassword() bool
	GetIsEmail() bool
	GetIsPhone() bool
}
