package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/tanguyRa/onvaou/internal/config"
	"github.com/tanguyRa/onvaou/internal/repository"

	"github.com/google/uuid"
)

// Webhook event types from Polar
const (
	// Subscription events
	EventSubscriptionCreated    = "subscription.created"
	EventSubscriptionUpdated    = "subscription.updated"
	EventSubscriptionActive     = "subscription.active"
	EventSubscriptionCanceled   = "subscription.canceled"
	EventSubscriptionUncanceled = "subscription.uncanceled"
	EventSubscriptionRevoked    = "subscription.revoked"

	// Order events
	EventOrderCreated  = "order.created"
	EventOrderPaid     = "order.paid"
	EventOrderUpdated  = "order.updated"
	EventOrderRefunded = "order.refunded"

	// Customer events
	EventCustomerCreated      = "customer.created"
	EventCustomerUpdated      = "customer.updated"
	EventCustomerDeleted      = "customer.deleted"
	EventCustomerStateChanged = "customer.state_changed"

	// Checkout events
	EventCheckoutCreated = "checkout.created"
	EventCheckoutUpdated = "checkout.updated"

	// Benefit events
	EventBenefitCreated      = "benefit.created"
	EventBenefitUpdated      = "benefit.updated"
	EventBenefitGrantCreated = "benefit_grant.created"
	EventBenefitGrantUpdated = "benefit_grant.updated"
	EventBenefitGrantRevoked = "benefit_grant.revoked"
)

// Subscription status values
const (
	SubscriptionStatusIncomplete        = "incomplete"
	SubscriptionStatusIncompleteExpired = "incomplete_expired"
	SubscriptionStatusTrialing          = "trialing"
	SubscriptionStatusActive            = "active"
	SubscriptionStatusPastDue           = "past_due"
	SubscriptionStatusCanceled          = "canceled"
	SubscriptionStatusUnpaid            = "unpaid"
)

// Recurring interval values
const (
	RecurringIntervalDay   = "day"
	RecurringIntervalWeek  = "week"
	RecurringIntervalMonth = "month"
	RecurringIntervalYear  = "year"
)

// Order billing reason values
const (
	BillingReasonPurchase           = "purchase"
	BillingReasonSubscriptionCreate = "subscription_create"
	BillingReasonSubscriptionCycle  = "subscription_cycle"
	BillingReasonSubscriptionUpdate = "subscription_update"
)

// WebhookEvent represents the base webhook event structure from Polar
// Following Standard Webhooks specification
type WebhookEvent struct {
	Type      string          `json:"type"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// PolarCustomer represents a customer in Polar
type PolarCustomer struct {
	ID             string            `json:"id"`
	CreatedAt      time.Time         `json:"created_at"`
	ModifiedAt     time.Time         `json:"modified_at"`
	Email          string            `json:"email"`
	EmailVerified  bool              `json:"email_verified"`
	Name           *string           `json:"name"`
	ExternalID     *string           `json:"external_id"`
	OrganizationID string            `json:"organization_id"`
	AvatarURL      *string           `json:"avatar_url"`
	Metadata       map[string]string `json:"metadata"`
	BillingAddress *BillingAddress   `json:"billing_address"`
	TaxID          []string          `json:"tax_id"`
	DeletedAt      *time.Time        `json:"deleted_at"`
}

// BillingAddress represents a customer's billing address
type BillingAddress struct {
	Line1      *string `json:"line1"`
	Line2      *string `json:"line2"`
	City       *string `json:"city"`
	State      *string `json:"state"`
	PostalCode *string `json:"postal_code"`
	Country    string  `json:"country"`
}

// PolarProduct represents a product in Polar
type PolarProduct struct {
	ID             string            `json:"id"`
	CreatedAt      time.Time         `json:"created_at"`
	ModifiedAt     time.Time         `json:"modified_at"`
	Name           string            `json:"name"`
	Description    *string           `json:"description"`
	IsRecurring    bool              `json:"is_recurring"`
	IsArchived     bool              `json:"is_archived"`
	OrganizationID string            `json:"organization_id"`
	Metadata       map[string]string `json:"metadata"`
	Prices         []ProductPrice    `json:"prices"`
}

// ProductPrice represents a price for a product
type ProductPrice struct {
	ID                     string    `json:"id"`
	CreatedAt              time.Time `json:"created_at"`
	ModifiedAt             time.Time `json:"modified_at"`
	AmountType             string    `json:"amount_type"`
	PriceAmount            *int      `json:"price_amount"`
	PriceCurrency          string    `json:"price_currency"`
	RecurringInterval      *string   `json:"recurring_interval"`
	RecurringIntervalCount *int      `json:"recurring_interval_count"`
}

// PolarSubscription represents a subscription in Polar
type PolarSubscription struct {
	ID                          string            `json:"id"`
	CreatedAt                   time.Time         `json:"created_at"`
	ModifiedAt                  time.Time         `json:"modified_at"`
	Amount                      int               `json:"amount"`
	Currency                    string            `json:"currency"`
	RecurringInterval           string            `json:"recurring_interval"`
	RecurringIntervalCount      int               `json:"recurring_interval_count"`
	Status                      string            `json:"status"`
	CurrentPeriodStart          time.Time         `json:"current_period_start"`
	CurrentPeriodEnd            time.Time         `json:"current_period_end"`
	CancelAtPeriodEnd           bool              `json:"cancel_at_period_end"`
	CanceledAt                  *time.Time        `json:"canceled_at"`
	StartedAt                   *time.Time        `json:"started_at"`
	EndsAt                      *time.Time        `json:"ends_at"`
	EndedAt                     *time.Time        `json:"ended_at"`
	TrialStart                  *time.Time        `json:"trial_start"`
	TrialEnd                    *time.Time        `json:"trial_end"`
	CustomerID                  string            `json:"customer_id"`
	ProductID                   string            `json:"product_id"`
	DiscountID                  *string           `json:"discount_id"`
	CheckoutID                  *string           `json:"checkout_id"`
	CustomerCancellationReason  *string           `json:"customer_cancellation_reason"`
	CustomerCancellationComment *string           `json:"customer_cancellation_comment"`
	Metadata                    map[string]string `json:"metadata"`
	Customer                    *PolarCustomer    `json:"customer"`
	Product                     *PolarProduct     `json:"product"`
	Prices                      []ProductPrice    `json:"prices"`
}

// PolarOrder represents an order in Polar
type PolarOrder struct {
	ID                 string             `json:"id"`
	CreatedAt          time.Time          `json:"created_at"`
	ModifiedAt         time.Time          `json:"modified_at"`
	Amount             int                `json:"amount"`
	TaxAmount          int                `json:"tax_amount"`
	Currency           string             `json:"currency"`
	BillingReason      string             `json:"billing_reason"`
	BillingAddress     *BillingAddress    `json:"billing_address"`
	CustomerID         string             `json:"customer_id"`
	ProductID          string             `json:"product_id"`
	ProductPriceID     string             `json:"product_price_id"`
	DiscountID         *string            `json:"discount_id"`
	SubscriptionID     *string            `json:"subscription_id"`
	CheckoutID         *string            `json:"checkout_id"`
	Metadata           map[string]string  `json:"metadata"`
	Customer           *PolarCustomer     `json:"customer"`
	Product            *PolarProduct      `json:"product"`
	Subscription       *PolarSubscription `json:"subscription"`
	IsInvoiceGenerated bool               `json:"is_invoice_generated"`
}

// PolarCheckout represents a checkout session in Polar
type PolarCheckout struct {
	ID                 string            `json:"id"`
	CreatedAt          time.Time         `json:"created_at"`
	ModifiedAt         time.Time         `json:"modified_at"`
	Status             string            `json:"status"`
	ClientSecret       string            `json:"client_secret"`
	URL                string            `json:"url"`
	ExpiresAt          time.Time         `json:"expires_at"`
	SuccessURL         string            `json:"success_url"`
	Amount             int               `json:"amount"`
	TaxAmount          int               `json:"tax_amount"`
	DiscountAmount     int               `json:"discount_amount"`
	NetAmount          int               `json:"net_amount"`
	TotalAmount        int               `json:"total_amount"`
	Currency           string            `json:"currency"`
	ProductID          string            `json:"product_id"`
	ProductPriceID     string            `json:"product_price_id"`
	DiscountID         *string           `json:"discount_id"`
	CustomerID         *string           `json:"customer_id"`
	CustomerEmail      *string           `json:"customer_email"`
	CustomerName       *string           `json:"customer_name"`
	CustomerExternalID *string           `json:"customer_external_id"`
	OrganizationID     string            `json:"organization_id"`
	Metadata           map[string]string `json:"metadata"`
}

// Standard Webhooks header names
const (
	HeaderWebhookID        = "webhook-id"
	HeaderWebhookTimestamp = "webhook-timestamp"
	HeaderWebhookSignature = "webhook-signature"

	// Signature tolerance: 5 minutes (Standard Webhooks recommends this)
	signatureToleranceSecs = 300
)

type PolarHandler struct {
	logger        *slog.Logger
	queries       *repository.Queries
	webhookSecret string
}

func NewPolarHandler(queries *repository.Queries, logger *slog.Logger, cfg config.PolarConfig) *PolarHandler {
	return &PolarHandler{
		queries:       queries,
		logger:        logger,
		webhookSecret: cfg.WebhookSecret,
	}
}

// Webhook handles incoming Polar webhook events
func (h *PolarHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Polar webhook received")

	// Read the raw body for signature verification
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed to read request body", "error", err)
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Failed to read request body")
		return
	}

	// Verify webhook signature if secret is configured
	if h.webhookSecret != "" {
		if err := h.verifySignature(r, body); err != nil {
			h.logger.Warn("webhook signature verification failed", "error", err)
			respondError(w, http.StatusUnauthorized, "INVALID_SIGNATURE", "Webhook signature verification failed")
			return
		}
	}

	// Parse the webhook event
	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		h.logger.Error("failed to parse webhook event", "error", err)
		respondError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "Failed to parse webhook payload")
		return
	}

	h.logger.Info("polar webhook event received", "type", event.Type, "timestamp", event.Timestamp)

	// Handle the event based on type
	if err := h.handleEvent(r.Context(), event); err != nil {
		h.logger.Error("failed to handle webhook event", "type", event.Type, "error", err)
		// Return 200 OK to prevent retries for business logic errors
		// Polar will retry on 4xx/5xx errors
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// verifySignature verifies the webhook signature following Standard Webhooks specification
// https://github.com/standard-webhooks/standard-webhooks/blob/main/spec/standard-webhooks.md
func (h *PolarHandler) verifySignature(r *http.Request, body []byte) error {
	webhookID := r.Header.Get(HeaderWebhookID)
	timestampStr := r.Header.Get(HeaderWebhookTimestamp)
	signatureHeader := r.Header.Get(HeaderWebhookSignature)

	if webhookID == "" || timestampStr == "" || signatureHeader == "" {
		return fmt.Errorf("missing required webhook headers")
	}

	// Parse timestamp
	var timestamp int64
	if _, err := fmt.Sscanf(timestampStr, "%d", &timestamp); err != nil {
		return fmt.Errorf("invalid timestamp format: %w", err)
	}

	// Check timestamp tolerance to prevent replay attacks
	now := time.Now().Unix()
	if math.Abs(float64(now-timestamp)) > signatureToleranceSecs {
		return fmt.Errorf("timestamp outside tolerance window")
	}

	// Construct the signed content: msg_id.timestamp.body
	signedContent := fmt.Sprintf("%s.%s.%s", webhookID, timestampStr, string(body))

	// Decode the webhook secret (may be base64 encoded with whsec_ prefix)
	secret := strings.TrimPrefix(h.webhookSecret, "whsec_")

	secretBytes, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		// If not base64, use raw bytes
		secretBytes = []byte(secret)
	}

	// Calculate expected signature using HMAC-SHA256
	mac := hmac.New(sha256.New, secretBytes)
	mac.Write([]byte(signedContent))
	expectedSig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	// Parse signatures from header (space-delimited list for key rotation)
	// Format: v1,signature1 v1,signature2
	signatures := strings.Split(signatureHeader, " ")
	for _, sig := range signatures {
		parts := strings.SplitN(sig, ",", 2)
		if len(parts) != 2 {
			continue
		}

		version := parts[0]
		signature := parts[1]

		// Only support v1 (symmetric HMAC-SHA256)
		if version != "v1" {
			continue
		}

		// Constant-time comparison to prevent timing attacks
		if hmac.Equal([]byte(signature), []byte(expectedSig)) {
			return nil
		}
	}

	return fmt.Errorf("no matching signature found")
}

// handleEvent dispatches the event to the appropriate handler
func (h *PolarHandler) handleEvent(ctx context.Context, event WebhookEvent) error {
	switch event.Type {
	// Subscription events
	case EventSubscriptionCreated:
		return h.handleSubscriptionCreated(ctx, event.Data)
	case EventSubscriptionUpdated, EventSubscriptionActive, EventSubscriptionCanceled,
		EventSubscriptionUncanceled, EventSubscriptionRevoked:
		return h.handleSubscriptionUpdated(ctx, event.Data)

	// // Order events
	// case EventOrderCreated, EventOrderPaid:
	// 	return h.handleOrderCreated(ctx, event.Data)
	// case EventOrderUpdated, EventOrderRefunded:
	// 	return h.handleOrderUpdated(ctx, event.Data)

	// Customer events
	case EventCustomerCreated, EventCustomerUpdated, EventCustomerStateChanged:
		return h.handleCustomerUpdated(ctx, event.Data)
	case EventCustomerDeleted:
		return h.handleCustomerDeleted(ctx, event.Data)

	// // Checkout events
	// case EventCheckoutCreated, EventCheckoutUpdated:
	// 	return h.handleCheckoutUpdated(ctx, event.Data)

	default:
		h.logger.Debug("unhandled webhook event type", "type", event.Type)

		queries := queriesFromContext(ctx, h.queries)
		newEvent, err := queries.CreateEvent(
			ctx,
			repository.CreateEventParams{
				UserId: uuid.UUID{},
				Type:   event.Type,
				Data:   event.Data,
			},
		)

		if err != nil {
			h.logger.Error("couldn't store the event in db", "err", err)
		} else {
			h.logger.Debug("created event", "event_id", newEvent.ID.String())
		}
		return nil
	}
}

// handleSubscriptionCreated handles new subscription creation
func (h *PolarHandler) handleSubscriptionCreated(ctx context.Context, data json.RawMessage) error {
	var subscription PolarSubscription
	if err := json.Unmarshal(data, &subscription); err != nil {
		return fmt.Errorf("failed to parse subscription data: %w", err)
	}

	userId, err := uuid.Parse(*subscription.Customer.ExternalID)
	if err != nil {
		return fmt.Errorf("Couldn't parse user Id: ExternalID: %s, err: %v", *subscription.Customer.ExternalID, err)
	}

	queries := queriesFromContext(ctx, h.queries)
	_, err = queries.CreateSubscription(
		ctx,
		repository.CreateSubscriptionParams{
			UserId:           userId,
			ExternalId:       &subscription.ID,
			Tier:             subscription.Product.ID,
			Status:           subscription.Status,
			CurrentPeriodEnd: &subscription.CurrentPeriodEnd,
		},
	)
	if err != nil {
		h.logger.Error("Couldn't create subscription",
			"err", err,
			"subscription_id", subscription.ID,
			"customer_id", subscription.Customer.ExternalID,
			"polar_customer_id", subscription.CustomerID,
			"product_id", subscription.ProductID,
			"status", subscription.Status,
		)

		return err
	}

	return nil
}

// handleSubscriptionUpdated handles subscription updates (active, canceled, etc.)
func (h *PolarHandler) handleSubscriptionUpdated(ctx context.Context, data json.RawMessage) error {
	var subscription PolarSubscription
	if err := json.Unmarshal(data, &subscription); err != nil {
		return fmt.Errorf("failed to parse subscription data: %w", err)
	}

	h.logger.Info("subscription updated",
		"subscription_id", subscription.ID,
		"status", subscription.Status,
		"cancel_at_period_end", subscription.CancelAtPeriodEnd,
	)

	userId, err := uuid.Parse(*subscription.Customer.ExternalID)
	if err != nil {
		return fmt.Errorf("Couldn't parse user Id: ExternalID: %s, err: %v", *subscription.Customer.ExternalID, err)
	}

	// Detect if user downgraded
	// var scheduledTier *string
	// if subscription.CancelAtPeriodEnd {
	// 	scheduledTier := subscription.Pro
	// }
	queries := queriesFromContext(ctx, h.queries)
	_, err = queries.UpdateSubscriptionByUserID(
		ctx,
		repository.UpdateSubscriptionByUserIDParams{
			UserId:            userId,
			ExternalId:        &subscription.ID,
			Tier:              subscription.Product.ID,
			CancelAtPeriodEnd: subscription.CancelAtPeriodEnd,
			Status:            subscription.Status,
			CurrentPeriodEnd:  &subscription.CurrentPeriodEnd,
		},
	)
	if err != nil {
		h.logger.Error("Couldn't update subscription",
			"err", err,
			"subscription_id", subscription.ID,
			"customer_id", subscription.Customer.ExternalID,
			"polar_customer_id", subscription.CustomerID,
			"product_id", subscription.ProductID,
			"status", subscription.Status,
		)

		return err
	}

	return nil
}

// handleOrderCreated handles new order creation
func (h *PolarHandler) handleOrderCreated(_ context.Context, data json.RawMessage) error {
	var order PolarOrder
	if err := json.Unmarshal(data, &order); err != nil {
		return fmt.Errorf("failed to parse order data: %w", err)
	}

	h.logger.Info("order created",
		"order_id", order.ID,
		"customer_id", order.CustomerID,
		"billing_reason", order.BillingReason,
		"amount", order.Amount,
	)

	// TODO: Implement order creation logic
	// - Track order for analytics
	// - Handle subscription renewals (billing_reason: subscription_cycle)

	return nil
}

// handleOrderUpdated handles order updates (refunds, etc.)
func (h *PolarHandler) handleOrderUpdated(_ context.Context, data json.RawMessage) error {
	var order PolarOrder
	if err := json.Unmarshal(data, &order); err != nil {
		return fmt.Errorf("failed to parse order data: %w", err)
	}

	h.logger.Info("order updated",
		"order_id", order.ID,
		"billing_reason", order.BillingReason,
	)

	// TODO: Implement order update logic
	// - Handle refunds

	return nil
}

// handleCustomerUpdated handles customer creation and updates
func (h *PolarHandler) handleCustomerUpdated(_ context.Context, data json.RawMessage) error {
	var customer PolarCustomer
	if err := json.Unmarshal(data, &customer); err != nil {
		return fmt.Errorf("failed to parse customer data: %w", err)
	}

	h.logger.Info("customer updated",
		"customer_id", customer.ID,
		"external_id", customer.ExternalID,
		"email", customer.Email,
	)

	// TODO: Implement customer update logic
	// - Sync customer data with local user records

	return nil
}

// handleCustomerDeleted handles customer deletion
func (h *PolarHandler) handleCustomerDeleted(_ context.Context, data json.RawMessage) error {
	var customer PolarCustomer
	if err := json.Unmarshal(data, &customer); err != nil {
		return fmt.Errorf("failed to parse customer data: %w", err)
	}

	h.logger.Info("customer deleted",
		"customer_id", customer.ID,
		"external_id", customer.ExternalID,
	)

	// TODO: Implement customer deletion logic

	return nil
}

// handleCheckoutUpdated handles checkout session updates
func (h *PolarHandler) handleCheckoutUpdated(_ context.Context, data json.RawMessage) error {
	var checkout PolarCheckout
	if err := json.Unmarshal(data, &checkout); err != nil {
		return fmt.Errorf("failed to parse checkout data: %w", err)
	}

	h.logger.Info("checkout updated",
		"checkout_id", checkout.ID,
		"status", checkout.Status,
		"customer_external_id", checkout.CustomerExternalID,
	)

	// TODO: Implement checkout update logic
	// - Track checkout conversion analytics

	return nil
}

// ParseSubscriptionEvent parses a raw webhook payload into a subscription
func ParseSubscriptionEvent(data json.RawMessage) (*PolarSubscription, error) {
	var subscription PolarSubscription
	if err := json.Unmarshal(data, &subscription); err != nil {
		return nil, fmt.Errorf("failed to parse subscription: %w", err)
	}
	return &subscription, nil
}

// ParseOrderEvent parses a raw webhook payload into an order
func ParseOrderEvent(data json.RawMessage) (*PolarOrder, error) {
	var order PolarOrder
	if err := json.Unmarshal(data, &order); err != nil {
		return nil, fmt.Errorf("failed to parse order: %w", err)
	}
	return &order, nil
}

// ParseCustomerEvent parses a raw webhook payload into a customer
func ParseCustomerEvent(data json.RawMessage) (*PolarCustomer, error) {
	var customer PolarCustomer
	if err := json.Unmarshal(data, &customer); err != nil {
		return nil, fmt.Errorf("failed to parse customer: %w", err)
	}
	return &customer, nil
}

// ParseCheckoutEvent parses a raw webhook payload into a checkout
func ParseCheckoutEvent(data json.RawMessage) (*PolarCheckout, error) {
	var checkout PolarCheckout
	if err := json.Unmarshal(data, &checkout); err != nil {
		return nil, fmt.Errorf("failed to parse checkout: %w", err)
	}
	return &checkout, nil
}

// GetUserIDFromSubscription extracts the user ID from a subscription's customer external_id
func GetUserIDFromSubscription(subscription *PolarSubscription) *string {
	if subscription.Customer != nil && subscription.Customer.ExternalID != nil {
		return subscription.Customer.ExternalID
	}
	return nil
}

// IsSubscriptionActive returns true if the subscription status is active or trialing
func IsSubscriptionActive(subscription *PolarSubscription) bool {
	return subscription.Status == SubscriptionStatusActive ||
		subscription.Status == SubscriptionStatusTrialing
}

// IsRenewalOrder returns true if the order is a subscription renewal
func IsRenewalOrder(order *PolarOrder) bool {
	return order.BillingReason == BillingReasonSubscriptionCycle
}
