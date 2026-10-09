package interfaces

// ICRUDModel is the persistence surface MService needs from a metadata model.
type ICRUDModel interface {
	CreateValue_(value any) (any, error)
	Update_(value any, keyFields ...string) error
	Delete_(value any, keyFields ...string) error
	FindByID_(destination any, keyFields ...string) error
	FindByFields(destination any, filterFields ...string) error
	FindListOrdered(destination any, filters map[string]any, search string, searchFields []string, orderBy, orderDirection string, limit int) error
}

type IService interface {
	Create_(model ICRUDModel, value any) (any, error)
	Update_(model ICRUDModel, value any, keyFields ...string) error
	Delete_(model ICRUDModel, value any, keyFields ...string) error
	FindByID_(model ICRUDModel, destination any, keyFields ...string) error
	FindByFields_(model ICRUDModel, destination any, filterFields ...string) error
	FindList_(model ICRUDModel, destination any, filters map[string]any, search string, searchFields []string, orderBy, orderDirection string, limit int) error
}
