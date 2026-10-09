package interfaces

type IMMigrator interface {
	Migrate() error
	SetModel(model IMModel)
}
