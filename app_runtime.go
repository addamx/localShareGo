package main

import (
	"context"

	"localShareGo/internal/runtimeapp"
)

func newAppRuntime(ctx context.Context) (*runtimeapp.AppRuntime, error) {
	return runtimeapp.New(ctx, assets)
}
