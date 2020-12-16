package biz

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/dowenliu-xyz/Go-000/Week04/internal/customer_follow/ent"
	"github.com/dowenliu-xyz/Go-000/Week04/internal/customer_follow/ent/customerfollow"
	xerrors "github.com/dowenliu-xyz/Go-000/Week04/internal/pkg/errors"
	"github.com/pkg/errors"
)

type CustomerFollowUseCase struct {
	client *ent.Client
}

func NewCustomerFollowUserCase(client *ent.Client) *CustomerFollowUseCase {
	return &CustomerFollowUseCase{
		client: client,
	}
}

func (c *CustomerFollowUseCase) WriteFollow(ctx context.Context, follow *ent.CustomerFollow) error {
	_, err := c.client.CustomerFollow.Create().
		SetStaffId(follow.StaffId).
		SetCustomerId(follow.CustomerId).
		SetContent(follow.Content).
		Save(ctx)
	if err != nil {
		return errors.WithStack(xerrors.SaveFailed(fmt.Sprintf("保存跟进记录操作失败: %v", err)))
	}
	return nil
}

func (c *CustomerFollowUseCase) ListFollow(ctx context.Context, customerId int64) ([]*ent.CustomerFollow, error) {
	follows, err := c.client.CustomerFollow.Query().
		Where(customerfollow.CustomerIdEQ(customerId)).
		All(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, xerrors.NotFound(fmt.Sprintf("未找到任何跟进: %v", err))
		}
		return nil, xerrors.QueryFailed(fmt.Sprintf("查询跟进失败: %v", err))
	}
	return follows, nil
}
