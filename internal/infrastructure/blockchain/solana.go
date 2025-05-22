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
