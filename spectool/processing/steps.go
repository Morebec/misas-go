package processing

import "context"

type Step[T context.Context] func(ctx T) error

func RunSteps[T context.Context](ctx T, steps ...Step[T]) error {
	for _, s := range steps {
		err := s(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
