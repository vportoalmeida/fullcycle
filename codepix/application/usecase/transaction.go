package usecase

import (
	"errors"
	"imersao_fullstack_fullcycle/codepix-go/domain/model"
	"log"
)

// TransactionUseCase is ..
type TransactionUseCase struct {
	TransactionRepository model.TransactionRepositoryInterface
	PixRepository         model.PixKeyRepositoryInterface
}

// Register is..
func (t *TransactionUseCase) Register(accountID string, amount float64, pixKeyto string, pixKeyKindTo string, description string) (*model.Transaction, error) {

	account, err := t.PixRepository.FindAccount(accountID)
	if err != nil {
		return nil, err
	}

	pixKey, err := t.PixRepository.FindKeyByKind(pixKeyto, pixKeyKindTo)
	if err != nil {
		return nil, err
	}

	transaction, err := model.NewTransaction(account, amount, pixKey, description)
	if err != nil {
		return nil, err
	}

	t.TransactionRepository.Save(transaction)
	if transaction.ID != "" {
		return transaction, nil
	}

	return nil, errors.New("unable to process this transaction")

}

// Confirm is ..
func (t *TransactionUseCase) Confirm(transactionID string) (*model.Transaction, error) {
	transaction, err := t.TransactionRepository.Find(transactionID)
	if err != nil {
		log.Println("Transaction not found", transactionID)
		return nil, err
	}

	transaction.Status = model.TransactionConfirmed
	err = t.TransactionRepository.Save(transaction)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

// Complete is ..
func (t *TransactionUseCase) Complete(transactionID string) (*model.Transaction, error) {
	transaction, err := t.TransactionRepository.Find(transactionID)
	if err != nil {
		log.Println("Transaction not found", transactionID)
		return nil, err
	}

	transaction.Status = model.TransactionCompleted
	err = t.TransactionRepository.Save(transaction)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

// Error is ..
func (t *TransactionUseCase) Error(transactionID string, reason string) (*model.Transaction, error) {
	transaction, err := t.TransactionRepository.Find(transactionID)
	if err != nil {
		return nil, err
	}

	transaction.Status = model.TransactionError
	transaction.CancelDescription = reason

	err = t.TransactionRepository.Save(transaction)
	if err != nil {
		return nil, err
	}

	return transaction, nil

}
