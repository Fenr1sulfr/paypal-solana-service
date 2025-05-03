package blockchain

// type SolanaClient struct {
// 	client     *rpc.Client
// 	privateKey solana.PrivateKey
// }

// func ComputeDiscriminator(method string) []byte {
// 	hash := sha256.Sum256([]byte("global:" + method))
// 	return hash[:8]
// }

// func EncodeInstructionData(instruction interface{}) ([]byte, error) {
// 	var buf bytes.Buffer
// 	encoder := borsh.NewEncoder(&buf)
// 	err := encoder.Encode(instruction)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return buf.Bytes(), nil
// }

// func NewSolanaClient(endpoint string, privateKey string) payment.SolanaClient {
// 	decodedKey, err := hex.DecodeString(privateKey)
// 	if err != nil {
// 		log.Fatalf("failed to decode private key: %v", err)
// 	}
// 	privKey, err := solana.PrivateKeyFromSolanaKeygenFileBytes(decodedKey)
// 	if err != nil {
// 		log.Fatalf("failed to create private key from bytes")
// 	}

// 	return &SolanaClient{
// 		client:     rpc.New(endpoint),
// 		privateKey: privKey,
// 	}
// }

// func (c *SolanaClient) StorePayment(ctx context.Context, payment *payment.Payment) (string, error) {
// 	// Define the program ID. Make SURE this is the correct program ID.

// 	client := rpc.New("https://api.devnet.solana.com") // Or your desired RPC endpoint

// 	// // Replace with your actual keypair. DO NOT hardcode in production!
// 	// payer, err := solana.WalletFromPrivateKeyBase58("5Uq5PxYJorBYVz5WFxz2u7PFvX4jrnd8D5FRHgcYYKpY2p7LcJrPK6HNwhrUEdq7wueGDY5cduQhsG2xLvNjo7Jf")
// 	// if err != nil {
// 	// 	log.Fatal("Problem with secret key")
// 	// 	return err
// 	// }

// 	ownerPayer, err := solana.WalletFromPrivateKeyBase58("3Bzq2Vw1GPb9SZma5NgX2o9j6hVbCqoafQSQS6zDoeDoWMZWujeXMgdwcoBKsVFTumvX8cqnCHq1nZUpVn8pqnsj")
// 	if err != nil {
// 		log.Fatal("Problem with secret key")
// 		return err

// 	}

// 	// Initialize Payment Storage
// 	initializeInstruction := InitializePaymentStorageInstruction{}
// 	initializeData, err := EncodeInstructionData(initializeInstruction)
// 	if err != nil {
// 		log.Fatalf("Error encoding instruction data: %v", err)
// 	}

// 	discriminator := ComputeDiscriminator("initialize_payments_storage")
// 	encodedInitializeData := append(discriminator, initializeData...)

// 	// Derive the payment storage account address
// 	paymentStorageSeed := "transactionBank"
// 	paymentStorage, _, err := solana.FindProgramAddress([][]byte{[]byte(paymentStorageSeed), ownerPayer.PublicKey().Bytes()}, programID)
// 	if err != nil {
// 		log.Fatalf("Error deriving payment storage address: %v", err)
// 		return err

// 	}
// 	firstMetaAccount := solana.AccountMeta{PublicKey: paymentStorage, IsWritable: true, IsSigner: false}
// 	secondMetaAccount := solana.AccountMeta{PublicKey: ownerPayer.PublicKey(), IsWritable: true, IsSigner: true}
// 	// thirdMetaAccount := solana.AccountMeta{PublicKey: payer.PublicKey(), IsWritable: true, IsSigner: true}
// 	systemMetaAccount := solana.AccountMeta{PublicKey: solana.SystemProgramID, IsWritable: false, IsSigner: false}
// 	metaSlice := solana.AccountMetaSlice{}
// 	metaSlice.Append(&firstMetaAccount)
// 	metaSlice.Append(&secondMetaAccount)
// 	// metaSlice.Append(&thirdMetaAccount)
// 	metaSlice.Append(&systemMetaAccount)
// 	// Create the instruction for initialize_payments_storage
// 	initializeIx := solana.NewInstruction(
// 		programID,
// 		metaSlice,
// 		encodedInitializeData,
// 	)

// 	// Add Payment
// 	addPaymentInstruction := AddPaymentInstruction{
// 		PaymentID:        id,
// 		PaymentOwner:     ownerPayer.PublicKey().String(),
// 		PaymentSender:    ownerPayer.PublicKey().String(),
// 		PaymentTimestamp: 1633029132,
// 		PaymentAmount:    1000,
// 		PaymentCurrency:  [3]byte{'U', 'S', 'D'},
// 		PaymentURL:       link,
// 	}

// 	addPaymentData, err := EncodeInstructionData(addPaymentInstruction)
// 	if err != nil {
// 		log.Fatalf("Error encoding add payment data: %v", err)
// 		return "", err

// 	}

// 	discriminator = ComputeDiscriminator("add_payment")
// 	encodedAddPaymentData := append(discriminator, addPaymentData...)

// 	// Derive the payment account address
// 	paymentSeed := "123456789012"
// 	paymentAccount, _, err := solana.FindProgramAddress([][]byte{[]byte(paymentSeed), ownerPayer.PublicKey().Bytes()}, programID)
// 	if err != nil {
// 		log.Fatalf("Error deriving payment account address: %v", err)
// 		return "", err

// 	}
// 	paymentSlice := solana.AccountMetaSlice{}
// 	newPublicKey := solana.NewWallet()

// 	newMetaKey := solana.AccountMeta{PublicKey: newPublicKey.PublicKey(), IsWritable: true, IsSigner: false}
// 	newMetaAccount := solana.AccountMeta{PublicKey: paymentAccount, IsWritable: true, IsSigner: true}

// 	paymentSlice.Append(&newMetaKey)
// 	paymentSlice.Append(&newMetaAccount)
// 	// paymentSlice.Append(&thirdMetaAccount)
// 	paymentSlice.Append(&systemMetaAccount)
// 	// Create the instruction for add_payment
// 	addPaymentIx := solana.NewInstruction(
// 		programID,
// 		metaSlice,
// 		encodedAddPaymentData,
// 	)

// 	// Get the latest blockhash
// 	hash, err := client.GetLatestBlockhash(context.TODO(), rpc.CommitmentFinalized)
// 	if err != nil {
// 		log.Fatalf("Error getting latest blockhash: %v", err)
// 		return "", err

// 	}

// 	// Create and send the transaction
// 	tx, err := solana.NewTransaction(
// 		[]solana.Instruction{initializeIx, addPaymentIx},
// 		hash.Value.Blockhash,
// 		solana.TransactionPayer(ownerPayer.PublicKey()),
// 	)
// 	if err != nil {
// 		log.Fatalf("Error creating transaction: %v", err)
// 		return "", err

// 	}

// 	// Sign the transaction
// 	_, err = tx.Sign(
// 		func(key solana.PublicKey) *solana.PrivateKey {
// 			if ownerPayer.PublicKey().Equals(key) {
// 				return &ownerPayer.PrivateKey
// 			}
// 			if ownerPayer.PublicKey().Equals(key) {
// 				return &ownerPayer.PrivateKey
// 			}
// 			return nil
// 		},
// 	)
// 	if err != nil {
// 		log.Fatalf("Error signing transaction: %v", err)
// 		return "", err

// 	}

// 	// Send the transaction
// 	sig, err := client.SendTransaction(context.TODO(), tx)
// 	if err != nil {
// 		log.Fatalf("Error sending transaction: %v", err)
// 		return "", err

// 	}

// 	spew.Dump(sig)
// 	return "", err
// }

// func (c *SolanaClient) VerifyTransaction(ctx context.Context, signature string) (bool, time.Time, error) {
// 	if signature == "" {
// 		return false, errors.New("signature is empty")
// 	}

// 	// Get the transaction status using the signature
// 	txStatus, err := c.client.GetConfirmedTransaction(ctx, solana.SignatureFromBase58(signature))
// 	if err != nil {
// 		if jsonrpc.IsErrNotFound(err) {
// 			return false, nil
// 		}
// 		return false, err
// 	}

// 	// Check if the transaction is confirmed
// 	if txStatus == nil || txStatus.Meta == nil || txStatus.Meta.Err != nil {
// 		return false, nil
// 	}

// 	return true, nil
// }

// // Add other methods as needed...
