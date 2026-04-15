# dedup

```tree
dedup/
├── README.md
└── dedup.go
    ├── type Decision {EventID: uuid.UUID, Action: string}
    ├── func CheckDuplicate(ctx context.Context, tx pgx.Tx, event model.Event) (Decision, error)
    ├── func tokenSortRatio(left, right string) int
    ├── func sortTokens(value string) string
    ├── func levenshtein(left, right []rune) int
    └── func min(values ...int) int
```
