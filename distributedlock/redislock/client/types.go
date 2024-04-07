package client

import (
	"context"
)

type LockClient interface {
	Eval(ctx context.Context, src string, keyCount int, keysAndArgs []interface{}) (interface{}, error)
}
