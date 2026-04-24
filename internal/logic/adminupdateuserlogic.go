package logic

import (
	"context"
	"fmt"

	"api/internal/svc"
	"api/internal/types"
	"api/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminUpdateUserLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUpdateUserLogic {
	return &AdminUpdateUserLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminUpdateUserLogic) AdminUpdateUser(req *types.AdminUpdateUserReq) (resp *types.SimpleResp, err error) {
	if err = requireAdminUser(l.ctx); err != nil {
		return nil, err
	}
	if l.svcCtx.UsersModel == nil {
		return okSimple("用户资料已更新（演示数据）"), nil
	}
	if req.Id == 0 {
		return nil, fmt.Errorf("user id is required")
	}

	input, err := normalizeAdminUserProfileInput(req)
	if err != nil {
		return nil, err
	}

	item, err := l.svcCtx.UsersModel.FindOne(l.ctx, req.Id)
	if err != nil {
		return nil, err
	}

	if err = ensureAdminUserProfileUnique(l.ctx, l.svcCtx, req.Id, input); err != nil {
		return nil, err
	}

	item.Nickname = input.Nickname
	item.Email = nullString(input.Email)
	item.Mobile = nullString(input.Mobile)
	item.Avatar = input.Avatar
	item.Status = uint64(input.Status)

	if err = l.svcCtx.UsersModel.Update(l.ctx, item); err != nil {
		return nil, mysqlDuplicateMessage("用户", err)
	}

	return okSimple("用户资料已更新"), nil
}

func ensureAdminUserProfileUnique(ctx context.Context, svcCtx *svc.ServiceContext, userID uint64, input adminUserProfileInput) error {
	if input.Email != "" {
		matched, err := svcCtx.UsersModel.FindOneByEmailString(ctx, input.Email)
		if err != nil && err != model.ErrNotFound {
			return err
		}
		if err == nil && matched.Id != userID {
			return fmt.Errorf("email already exists")
		}
	}

	if input.Mobile != "" {
		matched, err := svcCtx.UsersModel.FindOneByMobileString(ctx, input.Mobile)
		if err != nil && err != model.ErrNotFound {
			return err
		}
		if err == nil && matched.Id != userID {
			return fmt.Errorf("mobile already exists")
		}
	}

	return nil
}
