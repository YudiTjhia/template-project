package interfaces

type IMModel interface {
	GetName() string
	GetFields() []IMField
	GetTableName() string
}
