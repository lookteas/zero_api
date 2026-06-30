package logic

import (
	"context"
	"fmt"

	"api/internal/svc"
	"api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListFreemodePracticesLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListFreemodePracticesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListFreemodePracticesLogic {
	return &ListFreemodePracticesLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListFreemodePracticesLogic) ListFreemodePractices() (*types.FreeModePracticeListResp, error) {
	if l.svcCtx.FreeModePracticesModel == nil {
		return &types.FreeModePracticeListResp{Code: 0, Message: "ok", Data: types.FreeModePracticeListData{}}, nil
	}

	userID := currentUserID(l.ctx)
	items, err := l.svcCtx.FreeModePracticesModel.FindByUserID(l.ctx, userID, 100)
	if err != nil {
		return nil, fmt.Errorf("query free mode practices: %w", err)
	}

	list := make([]types.FreeModePracticeInfo, 0, len(items))
	for i := range items {
		item := items[i]
		list = append(list, freeModePracticeToInfo(&item))
	}

	return &types.FreeModePracticeListResp{
		Code:    0,
		Message: "ok",
		Data: types.FreeModePracticeListData{
			List: list,
		},
	}, nil
}
