// +build wireinject
// The build tag makes sure the stub is not built in the final build.

package di

import (
	"github.com/dowenliu-xyz/Go-000/Week04/internal/customer_follow/biz"
	"github.com/dowenliu-xyz/Go-000/Week04/internal/customer_follow/service"
	"github.com/google/wire"
)

//go:generate wire gen
func InitHttpService(database Database) (*service.CustomerFollowHttpService, func(), error) {
	panic(wire.Build(NewClient, biz.NewCustomerFollowUserCase, wire.Bind(new(service.CustomerFollowUseCase), new(*biz.CustomerFollowUseCase)), service.NewCustomerFollowService))
}
