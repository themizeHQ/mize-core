package repository

type Paginate struct {
	Limit int
	Page  int
}

type DBRepository interface {
	CreateOne(payload interface{}) interface{}
	FindOneById(id string) interface{}
	FindOneByFilter(filter interface{}) interface{}
	FindManyByFilter(filter interface{}, paginate Paginate) []interface{}
}
