package logic

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/crypto/bcrypt"
	"worker/internal/store"
	"worker/internal/svc"
	"worker/worker"
)

type CreateUserLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateUserLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateUserLogic {
	return &CreateUserLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateUserLogic) CreateUser(in *worker.CreateUserRequest) (*worker.UsersResponse, error) {
	username := strings.TrimSpace(in.Username)
	name := strings.TrimSpace(in.Name)
	password := in.Password
	if username == "" || name == "" || password == "" {
		return nil, errors.New("用户名、姓名和密码不能为空")
	}
	if len(password) < 6 {
		return nil, errors.New("密码长度不能少于6位")
	}
	if in.Role < 1 || in.Role > 3 {
		return nil, errors.New("角色不合法")
	}
	if in.Role == 3 && in.BuildingId <= 0 {
		return nil, errors.New("宿管账号必须绑定楼栋")
	}
	if in.Role == 2 && len(in.BuildingIds) == 0 {
		return nil, errors.New("工人工号必须至少管辖一栋楼")
	}
	existing, err := store.FindUserByUsernameAll(l.ctx, l.svcCtx.DB, username)
	if err == nil && existing != nil {
		return nil, errors.New("用户名已存在")
	}
	if err != nil && !errors.Is(err, sqlx.ErrNotFound) {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	u := &store.User{
		Username: username,
		Password: string(hash),
		Role:     in.Role,
		Name:     name,
		Status:   1,
	}
	phone := strings.TrimSpace(in.Phone)
	if phone != "" {
		u.Phone = sql.NullString{String: phone, Valid: true}
	}
	if in.Role == 3 {
		u.BuildingID = sql.NullInt64{Int64: in.BuildingId, Valid: true}
	}
	id, err := store.InsertUser(l.ctx, l.svcCtx.DB, u)
	if err != nil {
		return nil, err
	}
	if in.Role == 2 {
		for _, b := range in.BuildingIds {
			if err := store.InsertWorkerBuilding(l.ctx, l.svcCtx.DB, id, b); err != nil {
				return nil, err
			}
		}
	}
	u.ID = id
	return &worker.UsersResponse{Users: []*worker.User{userToPbWithBuildings(*u, in.BuildingIds)}}, nil
}
