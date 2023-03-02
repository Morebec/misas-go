# Messaging: Queries

## Implement a Query
```go
type GetUserByIdQuery struct {
	ID string `json:"id"`
}

func (c ImplementQuery) TypeName() query.PayloadTypeName {
	return "user.register"
}
```

## Implement a Query Handler
```go
func GetUserByIdQueryHandler() query.HandlerFunc {
	return func(ctx context.Context, q query.Query) (any, error) {
	    qry, ok := q.Payload.(GetUserByIdQuery)
	    if !ok {
                return errors.New("invalid_query")
            }
        // Implement logic ...
    }
}
```

## Register a Query Handler with the Query Bus
```go
cb := query.NewInMemoryBus()
cb.RegisterHandler(GetUserByIdQuery{}.TypeName(), GetUserByIdQueryHandler())
```

## Send query to the query bus
```go
cb := query.NewInMemoryBus()
cb.Send(context.Background(), query.New(GetUserByIdQuery{
	Username: "misas",
	Password: "a_password"
}))
```