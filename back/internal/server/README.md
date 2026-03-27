# server

```tree
server/
├── README.md
├── middleware.go
│   ├── func (*Server) logRequest(next http.Handler) http.Handler
│   ├── func (*Server) recoverPanic(next http.Handler) http.Handler
│   ├── func (*Server) requireAuthentication(next http.Handler) http.Handler
│   ├── func (*Server) withDBSession(next http.Handler) http.Handler
│   ├── func (*Server) authenticate(next http.Handler) http.Handler
│   ├── func isInternalRequest(r *http.Request, token string) bool
│   └── func (*Server) requireUserOwnership(next http.Handler) http.Handler
├── middleware_test.go
│   ├── func TestAuthenticateMiddleware(t *testing.T)
│   ├── func TestRequireAuthenticationMiddleware(t *testing.T)
│   ├── func TestWithDBSessionMiddleware(t *testing.T)
│   └── func requireLocalDB(t *testing.T) *pgxpool.Pool
├── routes.go
│   └── func (*Server) initRoutes() http.Handler
├── server.go
│   ├── type Server {config: config.Config, logger: *slog.Logger, pool: *pgxpool.Pool, queries: *repository.Queries, handlers: *handlers.Handlers, authVerificationKeyset: *jwk.Set}
│   ├── func New(cfg config.Config) *Server
│   ├── func (*Server) Start() error
│   └── func (*Server) Shutdown()
└── utils.go
    ├── func (*Server) serverError(w http.ResponseWriter, r *http.Request, err error)
    └── func (*Server) clientError(w http.ResponseWriter, status int)
```
