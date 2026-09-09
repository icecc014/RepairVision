package logic

import (
	"context"
	"math"

	"github.com/zeromicro/go-zero/core/logx"
	"map/mapclient"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
	"worker/workerclient"
)

type AdminFeedbackStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminFeedbackStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminFeedbackStatsLogic {
	return &AdminFeedbackStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminFeedbackStatsLogic) AdminFeedbackStats() (resp *types.AdminFeedbackStatsResponse, err error) {
	total, avg, ratings, err := store.FeedbackStats(l.ctx, l.svcCtx.DB)
	if err != nil {
		return nil, errs.Internal(err)
	}
	recentRows, err := store.ListRecentFeedback(l.ctx, l.svcCtx.DB, 10)
	if err != nil {
		return nil, errs.Internal(err)
	}
	buildingResp, err := l.svcCtx.MapRpc.ListBuildings(l.ctx, &mapclient.ListBuildingsRequest{})
	if err != nil {
		return nil, errs.Upstream()
	}
	buildingNames := make(map[int64]string, len(buildingResp.Buildings))
	for _, b := range buildingResp.Buildings {
		buildingNames[b.Id] = b.Code + " " + b.Name
	}
	workerIDs := make([]int64, 0, len(recentRows))
	for _, row := range recentRows {
		if row.WorkerID.Valid {
			workerIDs = append(workerIDs, row.WorkerID.Int64)
		}
	}
	workerNames := make(map[int64]string)
	if len(workerIDs) > 0 {
		users, err := l.svcCtx.WorkerRpc.GetUsers(l.ctx, &workerclient.UserIdsRequest{Ids: workerIDs})
		if err != nil {
			return nil, errs.Upstream()
		}
		for _, u := range users.Users {
			workerNames[u.Id] = u.Name
		}
	}

	resp = &types.AdminFeedbackStatsResponse{
		Total:     total,
		AvgRating: math.Round(avg*100) / 100,
		Ratings:   make([]types.FeedbackRatingItem, 0, len(ratings)),
		Recent:    make([]types.FeedbackRecentItem, 0, len(recentRows)),
	}
	for _, r := range ratings {
		resp.Ratings = append(resp.Ratings, types.FeedbackRatingItem{Rating: r.Rating, Cnt: r.Cnt})
	}
	for _, row := range recentRows {
		comment := ""
		if row.Comment.Valid {
			comment = row.Comment.String
		}
		item := types.FeedbackRecentItem{
			OrderNo:      row.OrderNo,
			BuildingId:   row.BuildingID,
			BuildingName: buildingNames[row.BuildingID],
			Room:         row.Room,
			Rating:       row.Rating,
			Comment:      comment,
			CreatedAt:    row.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if row.WorkerID.Valid {
			item.WorkerName = workerNames[row.WorkerID.Int64]
		}
		resp.Recent = append(resp.Recent, item)
	}
	return resp, nil
}
