package logic

import (
	"context"
	"fmt"
)

func requireAdminUser(ctx context.Context) error {
	if currentAdminID(ctx) == 0 {
		return fmt.Errorf("admin permission required")
	}
	return nil
}
