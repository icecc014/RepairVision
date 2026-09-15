package logic

import (
	"context"
	"encoding/json"
	"sort"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/store"
	"order/internal/svc"
	"order/internal/types"
)

type CampusDistanceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCampusDistanceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CampusDistanceLogic {
	return &CampusDistanceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CampusDistances 用已保存的区域概览构建栅格路网，返回建筑入口信息与两两路网距离矩阵，
// 供管理端绘制时直观看到"路怎么走、楼与楼多远"。
func (l *CampusDistanceLogic) CampusDistances() (resp *types.CampusDistanceResponse, err error) {
	layout, err := store.FindDefaultCampusLayout(l.ctx, l.svcCtx.DB)
	if err != nil || layout == nil || !layout.LayoutJson.Valid {
		return &types.CampusDistanceResponse{Buildings: []types.CampusDistanceBuilding{}, Pairs: []types.CampusDistancePair{}}, nil
	}
	net := buildRoadNetwork(layout.LayoutJson.String, defaultGridMeters)
	out := &types.CampusDistanceResponse{
		GridMeters: defaultGridMeters,
		Buildings:  []types.CampusDistanceBuilding{},
		Pairs:      []types.CampusDistancePair{},
	}
	if net == nil {
		return out, nil
	}
	// 建筑名称：优先取概览图元上的名称（拿不到时回落为 #ID）
	names := map[int64]string{}
	var raw campusLayoutJSON
	if err := json.Unmarshal([]byte(layout.LayoutJson.String), &raw); err == nil {
		for _, b := range raw.Blocks {
			if b.Kind == "building" && b.BuildingID > 0 {
				if _, ok := names[b.BuildingID]; !ok {
					names[b.BuildingID] = b.Label
				}
			}
		}
	}
	for _, id := range net.Buildings() {
		name := names[id]
		if name == "" {
			name = "#" + strconv.FormatInt(id, 10)
		}
		out.Buildings = append(out.Buildings, types.CampusDistanceBuilding{
			BuildingId: id,
			Name:       name,
			EntryCount: int64(net.EntryCount(id)),
			Connected:  !net.IsIsolated(id),
			MaxMeters:  net.NearestMeters(id),
			GapCells:   int64(net.GapCells(id)),
		})
	}
	sort.Slice(out.Buildings, func(i, j int) bool { return out.Buildings[i].BuildingId < out.Buildings[j].BuildingId })
	for _, p := range net.Pairs() {
		out.Pairs = append(out.Pairs, types.CampusDistancePair{FromId: p.FromID, ToId: p.ToID, Meters: p.Meters})
	}
	out.MaxMeters = net.MaxDistance()
	out.RoadCells = int64(net.RoadCellCount())
	return out, nil
}
