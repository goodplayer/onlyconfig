package domains

import "context"

type UserStore interface {
	QueryUserByUsername(ctx context.Context, username string) (*User, error)
}
