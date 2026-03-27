```tree
app/
├── README.md
└── internal/
    ├── config/
    │   ├── README.md
    │   └── config.go
    ├── crypto/
    │   ├── README.md
    │   ├── crypto.go
    │   └── crypto_test.go
    ├── handlers/
    │   ├── README.md
    │   ├── auth.go
    │   ├── auth_test.go
    │   ├── context.go
    │   ├── handlers.go
    │   ├── ping_test.go
    │   ├── polar.go
    │   ├── polar_test.go
    │   ├── response.go
    │   └── test_helpers_test.go
    ├── llm/
    │   ├── README.md
    │   ├── anthropic.go
    │   ├── anthropic_test.go
    │   ├── gemini.go
    │   ├── llm.go
    │   ├── llm_test.go
    │   └── mock.go
    ├── middleware/
    │   ├── README.md
    │   ├── chain.go
    │   ├── chain_test.go
    │   ├── cors.go
    │   ├── ratelimit.go
    │   └── ratelimit_test.go
    ├── providers/
    │   └── llmclient/
    │       ├── README.md
    │       └── llmclient.go
    ├── repository/
    │   ├── README.md
    │   ├── auth_accounts.sql.go
    │   ├── auth_jwks.sql.go
    │   ├── auth_sessions.sql.go
    │   ├── auth_users.sql.go
    │   ├── db.go
    │   ├── events.sql.go
    │   ├── models.go
    │   ├── pay_events.sql.go
    │   └── pay_subscriptions.sql.go
    ├── server/
    │   ├── README.md
    │   ├── middleware.go
    │   ├── middleware_test.go
    │   ├── routes.go
    │   ├── server.go
    │   └── utils.go
    ├── session/
    │   ├── README.md
    │   └── context.go
    ├── storage/
    │   ├── README.md
    │   ├── minio.go
    │   └── storage.go
    └── utils/
        ├── README.md
        └── date_parser.go
```
