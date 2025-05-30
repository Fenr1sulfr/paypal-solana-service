package http_client

type InitializeStorageRequest struct {
	PublicKey string `json:"publicKey"`
}

type InitializeStorageResponse struct {
	Success     bool   `json:"success"`
	Transaction string `json:"transaction"`
}

type AddPaymentRequest struct {
	PublicKey        string   `json:"publicKey"`
	PaymentId        string   `json:"paymentId"`
	PaymentOwner     string   `json:"paymentOwner"`
	PaymentSender    string   `json:"paymentSender"`
	PaymentTimestamp int64    `json:"paymentTimestamp"`
	PaymentAmount    int64    `json:"paymentAmount"`
	PaymentCurrency  []string `json:"paymentCurrency"`
	PaymentUrl       string   `json:"paymentUrl"`
}

type AddPaymentResponse struct {
	Success     bool   `json:"success"`
	Transaction string `json:"transaction"`
}
