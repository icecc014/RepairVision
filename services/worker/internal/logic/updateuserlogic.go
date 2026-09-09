package logic

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"worker/internal/store"
	"worker/internal/svc"
	"worker/worker"
)

type UpdateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserLogic {
	return &UpdateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateUserLogic) UpdateUser(in *worker.UpdateUserRequest) (*worker.UsersResponse, error) {
	existing, err := store.FindUserByID(l.ctx, l.svcCtx.DB, in.Id)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, errors.New("姓名不能为空")
	}
	if in.Role < 1 || in.Role > 3 {
		return nil, errors.New("角色不合法")
	}
	if existing.Role == 1 && in.Role != 1 {
		return nil, errors.New("管理员角色不可修改")
	}
	if in.Role == 3 && in.BuildingId <= 0 {
		return nil, errors.New("宿管账号必须绑定楼栋")
	}
	if in.Role == 2 && len(in.BuildingIds) == 0 {
		return nil, errors.New("工人工号必须至少管辖一栋楼")
	}
	maxConcurrent := in.MaxConcurrent
	if maxConcurrent <= 0 {
		maxConcurrent = 3
	}
	if maxConcurrent < 1 || maxConcurrent > 10 {
		return nil, errors.New("最大并发数需在1-10之间")
	}
	buildingID := sql.NullInt64{}
	if in.Role == 3 {
		buildingID = sql.NullInt64{Int64: in.BuildingId, Valid: true}
	}
	if err := store.UpdateUserProfile(l.ctx, l.svcCtx.DB, in.Id, name, strings.TrimSpace(in.Phone), in.Role, in.Status, maxConcurrent, buildingID); err != nil {
		return nil, err
	}
	if in.Role == 2 {
		if err := store.DeleteWorkerBuildings(l.ctx, l.svcCtx.DB, in.Id); err != nil {
			return nil, err
		}
		for _, b := range in.BuildingIds {
			if err := store.InsertWorkerBuilding(l.ctx, l.svcCtx.DB, in.Id, b); err != nil {
				return nil, err
			}
		}
	}
	u := existing
	u.Name = name
	u.Role = in.Role
	u.MaxConcurrent = maxConcurrent
	u.BuildingID = buildingID
	phone := strings.TrimSpace(in.Phone)
	u.Phone = sql.NullString{}
	if phone != "" {
		u.Phone = sql.NullString{String: phone, Valid: true}
	}
	return &worker.UsersResponse{Users: []*worker.User{userToPbWithBuildings(*u, in.BuildingIds)}}, nil
}
