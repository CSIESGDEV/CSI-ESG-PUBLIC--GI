package bootstrap

import (
	"context"
	"sme-api/app/kit/validator"
	"sme-api/app/repository"

	storage "github.com/myussufz/cloud-storage"
	"go.mongodb.org/mongo-driver/mongo"
)

// Bootstrap :
type Bootstrap struct {
	Repository *repository.Repository
	Storage    *storage.Builder
	MongoDB    *mongo.Client
	Validator  *validator.CustomValidator
}

// New :
func New() *Bootstrap {
	bs := new(Bootstrap)

	bs.initMongoDB()
	bs.initValidator()

	repo := repository.New(context.Background(), bs.MongoDB)

	bs.Repository = repo

	return bs
}
