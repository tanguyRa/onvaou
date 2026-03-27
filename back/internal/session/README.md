# session

```tree
session/
├── README.md
└── context.go
    ├── type contextKey string
    ├── type UserInfo {ID: string, Email: string, Name: string}
    └── func UserFromContext(ctx context.Context) (*UserInfo, bool)
```
