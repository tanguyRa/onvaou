# handlers

```tree
handlers/
├── README.md
├── auth.go
│   ├── type AuthHandler {logger: *slog.Logger, queries: *repository.Queries}
│   ├── func NewAuthHandler(queries *repository.Queries, logger *slog.Logger) *AuthHandler
│   └── func (*AuthHandler) UserFromRequest(w http.ResponseWriter, r *http.Request)
├── auth_test.go
│   └── func TestAuthUserFromRequest(t *testing.T)
├── context.go
│   ├── type queriesContextKey {}
│   ├── func WithQueries(ctx context.Context, queries *repository.Queries) context.Context
│   └── func queriesFromContext(ctx context.Context, fallback *repository.Queries) *repository.Queries
├── handlers.go
│   ├── type Handlers {queries: *repository.Queries, logger: *slog.Logger, config: config.Config, llm: *llmclient.Client, Auth: *AuthHandler, Polar: *PolarHandler}
│   ├── func New(queries *repository.Queries, logger *slog.Logger, cfg config.Config) (*Handlers, error)
│   └── func (*Handlers) Ping(w http.ResponseWriter, r *http.Request)
├── ping_test.go
│   └── func TestPing(t *testing.T)
├── polar.go
│   ├── type WebhookEvent {Type: string, Timestamp: time.Time, Data: json.RawMessage}
│   ├── type PolarCustomer {ID: string, CreatedAt: time.Time, ModifiedAt: time.Time, Email: string, EmailVerified: bool, Name: *string, ExternalID: *string, OrganizationID: string, AvatarURL: *string, Metadata: map[string]string, BillingAddress: *BillingAddress, TaxID: []string, DeletedAt: *time.Time}
│   ├── type BillingAddress {Line1: *string, Line2: *string, City: *string, State: *string, PostalCode: *string, Country: string}
│   ├── type PolarProduct {ID: string, CreatedAt: time.Time, ModifiedAt: time.Time, Name: string, Description: *string, IsRecurring: bool, IsArchived: bool, OrganizationID: string, Metadata: map[string]string, Prices: []ProductPrice}
│   ├── type ProductPrice {ID: string, CreatedAt: time.Time, ModifiedAt: time.Time, AmountType: string, PriceAmount: *int, PriceCurrency: string, RecurringInterval: *string, RecurringIntervalCount: *int}
│   ├── type PolarSubscription {ID: string, CreatedAt: time.Time, ModifiedAt: time.Time, Amount: int, Currency: string, RecurringInterval: string, RecurringIntervalCount: int, Status: string, CurrentPeriodStart: time.Time, CurrentPeriodEnd: time.Time, CancelAtPeriodEnd: bool, CanceledAt: *time.Time, StartedAt: *time.Time, EndsAt: *time.Time, EndedAt: *time.Time, TrialStart: *time.Time, TrialEnd: *time.Time, CustomerID: string, ProductID: string, DiscountID: *string, CheckoutID: *string, CustomerCancellationReason: *string, CustomerCancellationComment: *string, Metadata: map[string]string, Customer: *PolarCustomer, Product: *PolarProduct, Prices: []ProductPrice}
│   ├── type PolarOrder {ID: string, CreatedAt: time.Time, ModifiedAt: time.Time, Amount: int, TaxAmount: int, Currency: string, BillingReason: string, BillingAddress: *BillingAddress, CustomerID: string, ProductID: string, ProductPriceID: string, DiscountID: *string, SubscriptionID: *string, CheckoutID: *string, Metadata: map[string]string, Customer: *PolarCustomer, Product: *PolarProduct, Subscription: *PolarSubscription, IsInvoiceGenerated: bool}
│   ├── type PolarCheckout {ID: string, CreatedAt: time.Time, ModifiedAt: time.Time, Status: string, ClientSecret: string, URL: string, ExpiresAt: time.Time, SuccessURL: string, Amount: int, TaxAmount: int, DiscountAmount: int, NetAmount: int, TotalAmount: int, Currency: string, ProductID: string, ProductPriceID: string, DiscountID: *string, CustomerID: *string, CustomerEmail: *string, CustomerName: *string, CustomerExternalID: *string, OrganizationID: string, Metadata: map[string]string}
│   ├── type PolarHandler {logger: *slog.Logger, queries: *repository.Queries, webhookSecret: string}
│   ├── func NewPolarHandler(queries *repository.Queries, logger *slog.Logger, cfg config.PolarConfig) *PolarHandler
│   ├── func (*PolarHandler) HandleWebhook(w http.ResponseWriter, r *http.Request)
│   ├── func (*PolarHandler) verifySignature(r *http.Request, body []byte) error
│   ├── func (*PolarHandler) handleEvent(ctx context.Context, event WebhookEvent) error
│   ├── func (*PolarHandler) handleSubscriptionCreated(ctx context.Context, data json.RawMessage) error
│   ├── func (*PolarHandler) handleSubscriptionUpdated(ctx context.Context, data json.RawMessage) error
│   ├── func (*PolarHandler) handleOrderCreated(_ context.Context, data json.RawMessage) error
│   ├── func (*PolarHandler) handleOrderUpdated(_ context.Context, data json.RawMessage) error
│   ├── func (*PolarHandler) handleCustomerUpdated(_ context.Context, data json.RawMessage) error
│   ├── func (*PolarHandler) handleCustomerDeleted(_ context.Context, data json.RawMessage) error
│   ├── func (*PolarHandler) handleCheckoutUpdated(_ context.Context, data json.RawMessage) error
│   ├── func ParseSubscriptionEvent(data json.RawMessage) (*PolarSubscription, error)
│   ├── func ParseOrderEvent(data json.RawMessage) (*PolarOrder, error)
│   ├── func ParseCustomerEvent(data json.RawMessage) (*PolarCustomer, error)
│   ├── func ParseCheckoutEvent(data json.RawMessage) (*PolarCheckout, error)
│   ├── func GetUserIDFromSubscription(subscription *PolarSubscription) *string
│   ├── func IsSubscriptionActive(subscription *PolarSubscription) bool
│   └── func IsRenewalOrder(order *PolarOrder) bool
├── polar_test.go
│   └── func TestPolarWebhook(t *testing.T)
├── response.go
│   ├── func respondJSON(w http.ResponseWriter, status int, data any)
│   └── func respondError(w http.ResponseWriter, status int, code, message string)
└── test_helpers_test.go
    ├── type mockQueue {entries: []queueEntry, err: error}
    ├── type queueEntry {jobType: string, payload: any}
    ├── type mockStore {data: map[string][]byte}
    ├── func (*mockQueue) Enqueue(ctx context.Context, jobType string, payload any) error
    ├── func newMockStore() *mockStore
    ├── func (*mockStore) Put(ctx context.Context, key string, contentType string, data []byte) (string, error)
    ├── func (*mockStore) Get(ctx context.Context, key string) ([]byte, error)
    ├── func requireLocalDB(t *testing.T) *pgxpool.Pool
    ├── func withTx(t *testing.T, pool *pgxpool.Pool) (context.Context, repository.DBTX, *repository.Queries, func())
    ├── func createTestUser(t *testing.T, ctx context.Context, queries *repository.Queries) session.UserInfo
    ├── func withUserContext(ctx context.Context, user session.UserInfo, queries *repository.Queries) context.Context
    ├── func doJSONRequest(t *testing.T, handler http.HandlerFunc, method, path string, body any, ctx context.Context) *httptest.ResponseRecorder
    ├── func makeLogger() *slog.Logger
    ├── func mustParseTime(t *testing.T, value string) time.Time
    ├── func mustParseUUID(t *testing.T, value string) uuid.UUID
    └── func requireContains(t *testing.T, haystack, needle string)
```
