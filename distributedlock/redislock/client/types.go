package client

import (
	"context"
	"time"
)

type LockClient interface {
	SetNEX(ctx context.Context, key, value string, expire time.Duration) (int64, error)
	Eval(ctx context.Context, src string, keyCount int, keysAndArgs []interface{}) (interface{}, error)
}
