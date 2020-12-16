package service

import (
	"context"
	"github.com/dowenliu-xyz/Go-000/Week04/api/crm/customer_follow/v1"
	"github.com/dowenliu-xyz/Go-000/Week04/internal/customer_follow/ent"
	xerrors "github.com/dowenliu-xyz/Go-000/Week04/internal/pkg/errors"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CustomerFollowUseCase interface {
	WriteFollow(ctx context.Context, follow *ent.CustomerFollow) error
	ListFollow(ctx context.Context, customerId int64) ([]*ent.CustomerFollow, error)
}

type CustomerFollowHttpService struct {
	u CustomerFollowUseCase
}

func NewCustomerFollowService(u CustomerFollowUseCase) *CustomerFollowHttpService {
	return &CustomerFollowHttpService{
		u: u,
	}
}

func (s *CustomerFollowHttpService) WriteFollow(ctx context.Context, req *customer_follow.WriteFollowReq) (*customer_follow.OpReply, error) {
	follow := &ent.CustomerFollow{
		StaffId:    req.GetStaffId(),
		CustomerId: req.GetCustomerId(),
		Content:    req.GetContent(),
	}
	err := s.u.WriteFollow(ctx, follow)
	if err != nil {
		return nil, errors.WithMessage(err, "写跟进失败")
	}
	return &customer_follow.OpReply{}, nil
}

func (s *CustomerFollowHttpService) ListFollow(ctx context.Context, req *customer_follow.ListFollowReq) (*customer_follow.ListFollowResp, error) {
	dos, err := s.u.ListFollow(ctx, req.CustomerId)
	if err != nil {
		if xerrors.IsNotFound(err) {
			return &customer_follow.ListFollowResp{Follows: make([]*customer_follow.CustomerFollow, 0)}, nil
		}
		return nil, err
	}
	follows := make([]*customer_follow.CustomerFollow, 0, len(dos))
	for _, do := range dos {
		follows = append(follows, &customer_follow.CustomerFollow{
			StaffId:    do.StaffId,
			CustomerId: do.CustomerId,
			Content:    do.Content,
			CreateTime: timestamppb.New(do.CreateTime),
		})
	}
	return &customer_follow.ListFollowResp{Follows: follows}, nil
}
