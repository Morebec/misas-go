package entry

import "context"

type Entrypoint struct {
	OnStart func(ctx context.Context) error

	OnStop func(ctx context.Context) error

	OnError func(ctx context.Context) error
}
