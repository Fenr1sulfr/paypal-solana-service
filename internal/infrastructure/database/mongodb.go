package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"payment.Diploma.service/internal/domain/payment"
)

type mongoPayment struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	PaymentID       string             `bson:"payment_id"`
	Amount          float64            `bson:"amount"`
	Currency        string             `bson:"currency"`
	Status          string             `bson:"status"`
	Description     string             `bson:"description"`
	PayerEmail      string             `bson:"payer_email,omitempty"`
	PayerID         string             `bson:"payer_id,omitempty"`
	MerchantID      string             `bson:"merchant_id,omitempty"`
	SolanaSignature string             `bson:"solana_signature,omitempty"`
	SolanaAddress   string             `bson:"solana_address,omitempty"`
	CreatedAt       time.Time          `bson:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at"`
	CompletedAt     time.Time          `bson:"completed_at,omitempty"`
	Metadata        map[string]string  `bson:"metadata,omitempty"`
}

type PaymentRepository struct {
	collection *mongo.Collection
}

func NewPaymentRepository(db *mongo.Database) payment.Repository {
	return &PaymentRepository{
		collection: db.Collection("payments"),
	}
}

func (r *PaymentRepository) Save(ctx context.Context, p *payment.Payment) error {
	mp := toMongoPayment(p)
	mp.UpdatedAt = time.Now()

	if mp.ID.IsZero() {
		mp.ID = primitive.NewObjectID()
		mp.CreatedAt = time.Now()
		_, err := r.collection.InsertOne(ctx, mp)
		if err != nil {
			return err
		}
		p.ID = mp.ID.Hex()
		return nil
	}

	id, _ := primitive.ObjectIDFromHex(p.ID)
	mp.ID = id
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": id}, mp)
	return err
}

func (r *PaymentRepository) FindByID(ctx context.Context, id string) (*payment.Payment, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var mp mongoPayment
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&mp)
	if err != nil {
		return nil, err
	}

	return toDomainPayment(&mp), nil
}

func (r *PaymentRepository) FindByPaymentID(ctx context.Context, paymentID string) (*payment.Payment, error) {
	var mp mongoPayment
	err := r.collection.FindOne(ctx, bson.M{"payment_id": paymentID}).Decode(&mp)
	if err != nil {
		return nil, err
	}

	return toDomainPayment(&mp), nil
}

func (r *PaymentRepository) UpdateStatus(ctx context.Context, id, status string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	if status == "completed" {
		update["$set"].(bson.M)["completed_at"] = time.Now()
	}

	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	return err
}

func (r *PaymentRepository) ListPayments(ctx context.Context, status string, startDate, endDate time.Time, page, pageSize int) ([]*payment.Payment, int, error) {
	filter := bson.M{}

	if status != "" {
		filter["status"] = status
	}

	if !startDate.IsZero() && !endDate.IsZero() {
		filter["created_at"] = bson.M{
			"$gte": startDate,
			"$lte": endDate,
		}
	} else if !startDate.IsZero() {
		filter["created_at"] = bson.M{"$gte": startDate}
	} else if !endDate.IsZero() {
		filter["created_at"] = bson.M{"$lte": endDate}
	}

	// Count total documents
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Configure pagination
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"created_at": -1}) // Sort by created_at desc
	findOptions.SetSkip(int64((page - 1) * pageSize))
	findOptions.SetLimit(int64(pageSize))

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var mongoPayments []mongoPayment
	if err = cursor.All(ctx, &mongoPayments); err != nil {
		return nil, 0, err
	}

	payments := make([]*payment.Payment, len(mongoPayments))
	for i, mp := range mongoPayments {
		payments[i] = toDomainPayment(&mp)
	}

	return payments, int(count), nil
}

// Helper functions to convert between domain and MongoDB models
func toMongoPayment(p *payment.Payment) *mongoPayment {
	return &mongoPayment{
		PaymentID:       p.PaymentID,
		Amount:          p.Amount,
		Currency:        p.Currency,
		Status:          p.Status,
		Description:     p.Description,
		PayerEmail:      p.PayerEmail,
		PayerID:         p.PayerID,
		MerchantID:      p.MerchantID,
		SolanaSignature: p.SolanaSignature,
		SolanaAddress:   p.SolanaAddress,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
		CompletedAt:     p.CompletedAt,
		Metadata:        p.Metadata,
	}
}

func toDomainPayment(mp *mongoPayment) *payment.Payment {
	return &payment.Payment{
		ID:              mp.ID.Hex(),
		PaymentID:       mp.PaymentID,
		Amount:          mp.Amount,
		Currency:        mp.Currency,
		Status:          mp.Status,
		Description:     mp.Description,
		PayerEmail:      mp.PayerEmail,
		PayerID:         mp.PayerID,
		MerchantID:      mp.MerchantID,
		SolanaSignature: mp.SolanaSignature,
		SolanaAddress:   mp.SolanaAddress,
		CreatedAt:       mp.CreatedAt,
		UpdatedAt:       mp.UpdatedAt,
		CompletedAt:     mp.CompletedAt,
		Metadata:        mp.Metadata,
	}
}
