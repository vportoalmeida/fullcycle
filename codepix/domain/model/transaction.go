package model

import (
	"errors"
	"time"

	"github.com/asaskevich/govalidator"
	uuid "github.com/satori/go.uuid"
)

const (
	// TransactionPending is ...
	TransactionPending string = "pending"
	// TransactionCompleted is ...
	TransactionCompleted string = "completed"
	// TransactionError is ...
	TransactionError string = "error"
	// TransactionConfirmed is ...
	TransactionConfirmed string = "confirmed"
)

// TransactionRepositoryInterface is ...
type TransactionRepositoryInterface interface {
	Register(transaction *Transaction) error
	Save(transaction *Transaction) error
	Find(id string) (*Transaction, error)
}

// Transactions is ...
type Transactions struct {
	Transaction []Transaction
}

// Transaction is ...
type Transaction struct {
	Base              `valid:"required"`
	AccountFrom       *Account `valid:"-"`
	AccountFromID     string   `gorm:"column:account_from_id;type:uuid;" valid:"notnull"`
	Amount            float64  `json:"amount" gorm:"type:float" valid:"notnull"`
	PixKeyTo          *PixKey  `valid:"-"`
	PixKeyIDTo        string   `gorm:"column:pix_key_id_to;type:uuid;" valid:"notnull"`
	Status            string   `json:"status" gorm:"type:varchar(20)" valid:"notnull"`
	Description       string   `json:"description" gorm:"type:varchar(255)" valid:"-"`
	CancelDescription string   `json:"cancel_description" gorm:"type:varchar(255)" valid:"-"`
}

func init() {
	govalidator.SetFieldsRequiredByDefault(true)
}

func (t *Transaction) isValid() error {
	_, err := govalidator.ValidateStruct(t)

	if t.Amount <= 0 {
		return errors.New("the amount must be greater than 0")
	}

	if t.Status != TransactionPending && t.Status != TransactionCompleted && t.Status != TransactionError {
		return errors.New("invalid status for the transaction")
	}

	if t.PixKeyTo.AccountID == t.AccountFromID {
		return errors.New("the source and destination account cannot be the same")
	}

	if err != nil {
		return err
	}
	return nil
}

// Complete is ...
func (t *Transaction) Complete() error {
	t.Status = TransactionCompleted
	t.UpdatedAt = time.Now()
	err := t.isValid()
	return err
}

// Cancel is ...
func (t *Transaction) Cancel(description string) error {
	t.Status = TransactionError
	t.CancelDescription = description
	t.UpdatedAt = time.Now()
	err := t.isValid()
	return err
}

// NewTransaction is ...
func NewTransaction(accountFrom *Account, amount float64, pixKeyTo *PixKey, description string) (*Transaction, error) {
	transaction := Transaction{
		AccountFrom:   accountFrom,
		AccountFromID: accountFrom.ID,
		Amount:        amount,
		PixKeyTo:      pixKeyTo,
		PixKeyIDTo:    pixKeyTo.ID,
		Status:        TransactionPending,
		Description:   description,
	}
	transaction.ID = uuid.NewV4().String()
	transaction.CreatedAt = time.Now()
	err := transaction.isValid()
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}
