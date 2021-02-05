package kafka

import (
	"fmt"
	"os"

	"imersao_fullstack_fullcycle/codepix-go/application/factory"
	appmodel "imersao_fullstack_fullcycle/codepix-go/application/model"
	"imersao_fullstack_fullcycle/codepix-go/application/usecase"
	"imersao_fullstack_fullcycle/codepix-go/domain/model"

	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/jinzhu/gorm"
)

// KafkaProcessor is ..
type KafkaProcessor struct {
	Database     *gorm.DB
	Producer     *ckafka.Producer
	DeliveryChan chan ckafka.Event
}

// NewKafkaProcessor is ..
func NewKafkaProcessor(database *gorm.DB, producer *ckafka.Producer, deliveryChan chan ckafka.Event) *KafkaProcessor {
	return &KafkaProcessor{
		Database:     database,
		Producer:     producer,
		DeliveryChan: deliveryChan,
	}
}

// Consume is ..
func (k *KafkaProcessor) Consume() {
	configMap := &ckafka.ConfigMap{
		"bootstrap.servers": os.Getenv("kafkaBootstrapServers"),
		"group.id":          os.Getenv("kafkaConsumerGroupId"),
		"auto.offset.reset": "earliest",
	}
	c, err := ckafka.NewConsumer(configMap)

	if err != nil {
		panic(err)
	}

	topics := []string{os.Getenv("kafkaTransactionTopic"), os.Getenv("kafkaTransactionConfirmationTopic")}
	c.SubscribeTopics(topics, nil)

	fmt.Println("kafka consumer has been started")
	for {
		msg, err := c.ReadMessage(-1)
		if err == nil {
			fmt.Println(string(msg.Value))
			k.processMessage(msg)
		}
	}
}

// processMessage is ..
func (k *KafkaProcessor) processMessage(msg *ckafka.Message) {
	transactionsTopic := "transactions"
	transactionConfirmationTopic := "transaction_confirmation"

	switch topic := *msg.TopicPartition.Topic; topic {
	case transactionsTopic:
		k.processTransaction(msg)
	case transactionConfirmationTopic:
		k.processTransactionConfirmation(msg)
	default:
		fmt.Println("not a valid topic", string(msg.Value))
	}
}

// processTransaction is ..
func (k *KafkaProcessor) processTransaction(msg *ckafka.Message) error {
	transaction := appmodel.NewTransaction()
	err := transaction.ParseJSON(msg.Value)
	if err != nil {
		return err
	}

	transactionUseCase := factory.TransactionUseCaseFactory(k.Database)

	createdTransaction, err := transactionUseCase.Register(
		transaction.AccountID,
		transaction.Amount,
		transaction.PixKeyTo,
		transaction.PixKeyKindTo,
		transaction.Description,
		//transaction.ID,
	)
	if err != nil {
		fmt.Println("error registering transaction", err)
		return err
	}

	topic := "bank" + createdTransaction.PixKeyTo.Account.Bank.Code
	transaction.ID = createdTransaction.ID
	transaction.Status = model.TransactionPending
	transactionJSON, err := transaction.ToJSON()

	if err != nil {
		return err
	}

	err = Publish(string(transactionJSON), topic, k.Producer, k.DeliveryChan)
	if err != nil {
		return err
	}
	return nil
}

// processTransactionConfirmation is ..
func (k *KafkaProcessor) processTransactionConfirmation(msg *ckafka.Message) error {
	transaction := appmodel.NewTransaction()
	err := transaction.ParseJSON(msg.Value)
	if err != nil {
		return err
	}

	transactionUseCase := factory.TransactionUseCaseFactory(k.Database)

	if transaction.Status == model.TransactionConfirmed {
		err = k.confirmTransaction(transaction, transactionUseCase)
		if err != nil {
			return err
		}
	} else if transaction.Status == model.TransactionCompleted {
		_, err := transactionUseCase.Complete(transaction.ID)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

// confirmTransaction is ..
func (k *KafkaProcessor) confirmTransaction(transaction *appmodel.Transaction, transactionUseCase usecase.TransactionUseCase) error {
	confirmedTransaction, err := transactionUseCase.Confirm(transaction.ID)
	if err != nil {
		return err
	}

	topic := "bank" + confirmedTransaction.AccountFrom.Bank.Code
	transactionJSON, err := transaction.ToJSON()
	if err != nil {
		return err
	}

	err = Publish(string(transactionJSON), topic, k.Producer, k.DeliveryChan)
	if err != nil {
		return err
	}
	return nil
}
