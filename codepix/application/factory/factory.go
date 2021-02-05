package factory

import (
	"imersao_fullstack_fullcycle/codepix-go/application/usecase"
	"imersao_fullstack_fullcycle/codepix-go/infrastructure/repository"

	"github.com/jinzhu/gorm"
)

// TransactionUseCaseFactory is ..
func TransactionUseCaseFactory(database *gorm.DB) usecase.TransactionUseCase {
	pixRepository := repository.PixKeyRepositoryDb{Db: database}
	transactionRepository := repository.TransactionRepositoryDb{Db: database}

	transactionUseCase := usecase.TransactionUseCase{
		TransactionRepository: &transactionRepository,
		PixRepository:         pixRepository,
	}

	return transactionUseCase
}
