package container

import (
	"context"

	"github.com/docker/go-connections/nat"
)

type Docker interface {
	Start(context.Context, []string) error
	Stop(context.Context) error
	Port(context.Context, nat.Port) (nat.Port, error)
	Status(context.Context) string
	Name(context.Context) string
}
