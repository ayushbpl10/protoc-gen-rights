package right

import "context"

type UserIDer interface {
	UserID(ctx context.Context) string
}

type UserIDerFunc func(ctx context.Context) string

func (f UserIDerFunc) UserID(ctx context.Context) string {
	return f(ctx)
}
