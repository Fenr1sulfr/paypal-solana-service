package solana_payment

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"payment.Diploma.service/internal/domain/payment"
	"payment.Diploma.service/internal/infrastructure/blockchain/receipt_diploma"
)

type SolanaPaymentSystem struct {
	Client    *rpc.Client
	programID solana.PublicKey
}
type PaymentStatus struct {
	Success   bool      `json:"success"`
	Signature string    `json:"signature,omitempty"`
	Amount    float64   `json:"amount,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
	Confirmed bool      `json:"confirmed"`
	Error     string    `json:"error,omitempty"`	
}

type PaymentResult struct {
	Success   bool
	Signature string
	Error     error
}

var (
	programID   = "YRoyDgtmHvKDpFdksFPdcCB16ymBspq2kUhgVz18JFQ"
	storageSeed = "transactionBank"
)

func StringToUint8Slice(s string) []uint8 {
	result := make([]uint8, len(s))
	for i := 0; i < 3; i++ {
		result[i] = s[i]
	}
	return result
}

func New() *SolanaPaymentSystem {
	endpoint := rpc.DevNet_RPC
	client := rpc.New(endpoint)
	programID := solana.MustPublicKeyFromBase58(programID)
	return &SolanaPaymentSystem{
		Client:    client,
		programID: programID,
	}

}

func containsReference(accountKeys []solana.PublicKey, referenceKey string) bool {
	refPubKey, err := solana.PublicKeyFromBase58(referenceKey)
	if err != nil {
		return false
	}

	for _, key := range accountKeys {
		if key.Equals(refPubKey) {
			return true
		}
	}
	return false
}

func isValidPayment(tx *rpc.GetTransactionResult, expectedAmount float64, recipient, referenceKey string) bool {
	if tx == nil || tx.Meta == nil || tx.Meta.Err != nil {
		return false
	}

	// Проверяем инструкции
	for _, instruction := range tx.Transaction.GetParsedTransaction().Message.Instructions {
		// Проверяем transfer инструкции
		if isTransferInstruction(instruction, expectedAmount, recipient) {
			// Проверяем наличие reference в аккаунтах
			if containsReference(tx.Transaction.GetParsedTransaction().Message.AccountKeys, referenceKey) {
				return true
			}
		}
	}

	return false
}

func isTransferInstruction(instruction rpc.ParsedInstruction, expectedAmount float64, recipient string) bool {
	// Проверяем программу (System Program для SOL или Token Program для SPL)
	// и параметры перевода
	// Это зависит от конкретной реализации
	return true // упрощенная проверка
}

func (s *SolanaPaymentSystem) GenerateSolanaLink(authority solana.PublicKey, amount float64, label, message string) (string, string, error) {
	// Generate a unique reference (for example, using current timestamp)
	referenceKey := solana.NewWallet().PublicKey()
	recipient := authority.String()
	memo := "DiplomaPayment"

	link := fmt.Sprintf(
		"solana:%s?amount=%.9f&reference=%s&label=%s&message=%s&memo=%s",
		recipient,
		amount,
		referenceKey.String(),
		label,
		message,
		memo,
	)
	return link, referenceKey.String(), nil
}

func (s *SolanaPaymentSystem) VerifyPayment(referenceKey string, expectedAmount float64, recipient string) (*PaymentStatus, error) {
	// Преобразуем reference ключ
	refPubKey, err := solana.PublicKeyFromBase58(referenceKey)
	if err != nil {
		return nil, fmt.Errorf("invalid reference key: %v", err)
	}
	limit := 10 // Получаем подписи транзакций для reference аккаунта
	signatures, err := s.Client.GetSignaturesForAddressWithOpts(
		context.Background(),
		refPubKey,
		&rpc.GetSignaturesForAddressOpts{
			Limit: &limit, // последние 10 транзакций
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get signatures: %v", err)
	}

	for _, sig := range signatures {
		// Получаем детали транзакции
		tx, err := s.Client.GetTransaction(
			context.Background(),
			sig.Signature,
			&rpc.GetTransactionOpts{
				Encoding:   solana.EncodingJSON,
				Commitment: rpc.CommitmentConfirmed,
			},
		)
		if err != nil {
			continue
		}

		// Проверяем транзакцию
		if isValidPayment(tx, expectedAmount, recipient, referenceKey) {
			return &PaymentStatus{
				Success:   true,
				Signature: sig.Signature.String(),
				Amount:    expectedAmount,
				Timestamp: time.Unix(int64(*sig.BlockTime), 0),
				Confirmed: true,
			}, nil
		}
	}

	return &PaymentStatus{Success: false}, nil
}

func (s *SolanaPaymentSystem) InitStorage(authority solana.PublicKey, system solana.PublicKey) string {
	pda, _, _ := solana.FindProgramAddress([][]byte{
		[]byte(storageSeed),
		authority.Bytes(),
	},
		s.programID)
	builder := receipt_diploma.NewInitializePaymentsStorageInstruction(pda, authority, solana.SystemProgramID)
	instruction := builder.Build()
	blockhash, err := s.Client.GetRecentBlockhash(context.Background(), rpc.CommitmentFinalized)
	if err != nil {
		panic(err)
	}
	tx, err := solana.NewTransaction([]solana.Instruction{instruction}, blockhash.Value.Blockhash, solana.TransactionPayer(authority))
	if err != nil {
		panic(err)
	}
	return tx.MustToBase64()
}

func (s *SolanaPaymentSystem) AddPayment(authority solana.PublicKey, receipt payment.Payment) string {

	pda, _, err := solana.FindProgramAddress(
		[][]byte{
			[]byte(receipt.ID),
			authority.Bytes(),
		},
		s.programID,
	)
	if err != nil {
		log.Fatal(err)
	}

	pd2, _, _ := solana.FindProgramAddress([][]byte{
		[]byte(storageSeed),
		authority.Bytes(),
	},
		s.programID)
	inst := receipt_diploma.NewAddPaymentInstructionBuilder().
		SetPaymentId(receipt.ID).
		SetOwnerAccount(authority).
		SetPaymentSender(authority).
		SetPaymentTimestamp(uint64(receipt.CreatedAt.Unix())).
		SetPaymentAmount(uint64(receipt.Amount)).
		SetPaymentCurrency([3]uint8(StringToUint8Slice(receipt.Currency))).
		SetPaymentUrl("localhost:8080").
		SetPaymentAccount(pda).
		SetPaymentStorageAccount(pd2).
		SetOwnerAccount(authority).
		SetSystemProgramAccount(solana.SystemProgramID)

	if err := inst.Validate(); err != nil {
		log.Printf("%s", err.Error())
	}
	ix := inst.Build()
	hash, err := s.Client.GetLatestBlockhash(context.TODO(), rpc.CommitmentFinalized)
	if err != nil {
		log.Printf("%s", err.Error())
	}
	tx, err := solana.NewTransaction(
		[]solana.Instruction{ix},
		hash.Value.Blockhash,
		solana.TransactionPayer(authority),
	)
	if err != nil {
		log.Printf("%s", err.Error())

	}
	return tx.MustToBase64()

}

func NewSolanaClient(endpoint, programID string) *SolanaPaymentSystem {
	return New()
}

func (s *SolanaPaymentSystem) StorePayment(ctx context.Context, payment *payment.Payment) (string, error) {
	clientPublicKey, err := solana.PublicKeyFromBase58(payment.SolanaAddress)
	if err != nil {
		return "", err
	}
	if payment.Metadata["initializedAccount"] == "No" {
		return s.InitStorage(clientPublicKey, solana.SystemProgramID), nil
	}
	tx := s.AddPayment(clientPublicKey, *payment)
	return tx, nil

}
func (s *SolanaPaymentSystem) VerifyTransaction(ctx context.Context, signature string) (bool, time.Time, error) {
	return false, time.Now(), fmt.Errorf("This is not realized now")
}
