# Models

This folder contains the application's GORM models split by domain for easier maintenance.

Guidelines

- Each file should use `package models`.
- Keep related models together (e.g., RBAC models in `rbac.go`, auth models in `auth_provider.go` and `auth_token.go`).
- When adding a new model struct, add it to the `AutoMigrate` call in `config/dbConfig.go`.
  For example:

```go
if err := DB.AutoMigrate(
    &models.NewModel{},
    // ... other models
); err != nil {
    // handle error
}
```

- Use GORM tags to define indexes, constraints and cascade behavior.

- Run `go build ./...` (or `go vet ./...`) after changes to ensure everything compiles.
