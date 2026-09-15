package logic

import (
	"context"
	"encoding/json"
	"sync"

	"order/internal/store"
	"order/internal/svc"
)

// 默认栅格边长（米）：区域概览按格绘制，1 格按 10 米估算。
const defaultGridMeters = 10.0

// campusBlock 区域概览图元，字段与前端 utils/campusLayout.ts 对齐。
type campusBlock struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	Row        int    `json:"row"`
	Col        int    `json:"col"`
	RowSpan    int    `json:"rowSpan"`
	ColSpan    int    `json:"colSpan"`
	BuildingID int64  `json:"buildingId"`
	Label      string `json:"label"`
}

type campusLayoutJSON struct {
	Version int           `json:"version"`
	Cols    int           `json:"cols"`
	Rows    int           `json:"rows"`
	Blocks  []campusBlock `json:"blocks"`
}

// RoadNetwork 由区域概览的"道路"图元构建的栅格路网（4 邻接，边权 1 格）。
type RoadNetwork struct {
	cols, rows  int
	gridMeters  float64
	road        map[int]bool
	buildingIDs []int64
	entries     map[int64][]int
	isolated    map[int64]bool
	gap         map[int64]int
	dist        map[int64]map[int64]float64
	maxMeters   float64
}

func cellKey(cols, row, col int) int { return row*cols + col }
func cellRow(cols, key int) int      { return key / cols }
func cellCol(cols, key int) int      { return key % cols }

// buildRoadNetwork 解析区域概览 JSON，构建路网并预计算建筑间最短路距离。
// 布局为空、无道路或无建筑时返回 nil（调用方回退欧氏距离）。
func buildRoadNetwork(layoutJSON string, gridMeters float64) *RoadNetwork {
	if gridMeters <= 0 {
		gridMeters = defaultGridMeters
	}
	var layout campusLayoutJSON
	if err := json.Unmarshal([]byte(layoutJSON), &layout); err != nil {
		return nil
	}
	if layout.Cols <= 0 || layout.Rows <= 0 || len(layout.Blocks) == 0 {
		return nil
	}
	net := &RoadNetwork{
		cols:       layout.Cols,
		rows:       layout.Rows,
		gridMeters: gridMeters,
		road:       map[int]bool{},
		entries:    map[int64][]int{},
		isolated:   map[int64]bool{},
		gap:        map[int64]int{},
		dist:       map[int64]map[int64]float64{},
	}

	// 1) 道路格（支持跨格图元）
	for _, b := range layout.Blocks {
		if b.Kind != "road" {
			continue
		}
		rowSpan, colSpan := spanOf(b)
		for r := b.Row; r < b.Row+rowSpan; r++ {
			for c := b.Col; c < b.Col+colSpan; c++ {
				if r < 0 || c < 0 || r >= layout.Rows || c >= layout.Cols {
					continue
				}
				net.road[cellKey(layout.Cols, r, c)] = true
			}
		}
	}
	if len(net.road) == 0 {
		return nil
	}

	// 2) 建筑入口：与道路 4 邻接的建筑边界格
	seen := map[int64]bool{}
	for _, b := range layout.Blocks {
		if b.Kind != "building" || b.BuildingID <= 0 || seen[b.BuildingID] {
			continue
		}
		seen[b.BuildingID] = true
		rowSpan, colSpan := spanOf(b)
		entrySet := map[int]bool{}
		for r := b.Row; r < b.Row+rowSpan; r++ {
			for c := b.Col; c < b.Col+colSpan; c++ {
				if r < 0 || c < 0 || r >= layout.Rows || c >= layout.Cols {
					continue
				}
				for _, nb := range [][2]int{{r - 1, c}, {r + 1, c}, {r, c - 1}, {r, c + 1}} {
					if nb[0] < 0 || nb[1] < 0 || nb[0] >= layout.Rows || nb[1] >= layout.Cols {
						continue
					}
					if net.road[cellKey(layout.Cols, nb[0], nb[1])] {
						entrySet[cellKey(layout.Cols, nb[0], nb[1])] = true
					}
				}
			}
		}
		if len(entrySet) == 0 {
			// 建筑与道路不直接相邻时：取离建筑轮廓最近的道路格作为入口，
			// 这段步行距离（gap 格）会计入建筑间距离，保证任意绘制方式都能走路网。
			bestGap, bestKey := -1, -1
			for key := range net.road {
				g := boxGapCells(b, cellRow(layout.Cols, key), cellCol(layout.Cols, key))
				if bestGap < 0 || g < bestGap {
					bestGap, bestKey = g, key
				}
			}
			if bestKey >= 0 {
				entrySet[bestKey] = true
				net.gap[b.BuildingID] = bestGap
			} else {
				net.isolated[b.BuildingID] = true
			}
		}
		for k := range entrySet {
			net.entries[b.BuildingID] = append(net.entries[b.BuildingID], k)
		}
		net.buildingIDs = append(net.buildingIDs, b.BuildingID)
	}
	if len(net.buildingIDs) == 0 {
		return nil
	}

	// 3) 以每栋建筑的全部入口为源点做 BFS（边权相同，BFS 即最短路），
	//    得到"建筑 -> 建筑"的最短路格数，再折算为米。
	for _, from := range net.buildingIDs {
		if net.isolated[from] || len(net.entries[from]) == 0 {
			continue
		}
		reached := net.bfsFromEntries(net.entries[from])
		row := map[int64]float64{from: 0}
		for _, to := range net.buildingIDs {
			if to == from || net.isolated[to] {
				continue
			}
			best := -1
			for _, entry := range net.entries[to] {
				if d, ok := reached[entry]; ok && (best < 0 || d < best) {
					best = d
				}
			}
			if best >= 0 {
				row[to] = float64(best+net.gap[from]+net.gap[to]) * net.gridMeters
			}
		}
		for _, d := range row {
			if d > net.maxMeters {
				net.maxMeters = d
			}
		}
		net.dist[from] = row
	}
	return net
}

func spanOf(b campusBlock) (int, int) {
	rowSpan, colSpan := b.RowSpan, b.ColSpan
	if rowSpan <= 0 {
		rowSpan = 1
	}
	if colSpan <= 0 {
		colSpan = 1
	}
	return rowSpan, colSpan
}

// bfsFromEntries 从入口格集合出发做多源 BFS，返回"格号 -> 步数"。
func (n *RoadNetwork) bfsFromEntries(entries []int) map[int]int {
	visited := make(map[int]int, len(n.road))
	queue := make([]int, 0, len(n.road))
	for _, e := range entries {
		if _, ok := visited[e]; ok {
			continue
		}
		visited[e] = 0
		queue = append(queue, e)
	}
	for head := 0; head < len(queue); head++ {
		cur := queue[head]
		r, c := cellRow(n.cols, cur), cellCol(n.cols, cur)
		for _, nb := range [][2]int{{r - 1, c}, {r + 1, c}, {r, c - 1}, {r, c + 1}} {
			if nb[0] < 0 || nb[1] < 0 || nb[0] >= n.rows || nb[1] >= n.cols {
				continue
			}
			key := cellKey(n.cols, nb[0], nb[1])
			if !n.road[key] {
				continue
			}
			if _, ok := visited[key]; ok {
				continue
			}
			visited[key] = visited[cur] + 1
			queue = append(queue, key)
		}
	}
	return visited
}

// DistanceBetween 返回两栋建筑间的路网距离（米）；无路网、孤立建筑或不可达时返回 false。
func (n *RoadNetwork) DistanceBetween(from, to int64) (float64, bool) {
	if n == nil {
		return 0, false
	}
	if from == to {
		return 0, true
	}
	row, ok := n.dist[from]
	if !ok {
		return 0, false
	}
	d, ok := row[to]
	return d, ok
}

// MaxDistance 返回距离矩阵中的最大值（米），用于归一化距离得分。
func (n *RoadNetwork) MaxDistance() float64 {
	if n == nil {
		return 0
	}
	return n.maxMeters
}

// Buildings 返回参与路网计算的建筑 ID（含孤立建筑）。
func (n *RoadNetwork) Buildings() []int64 {
	if n == nil {
		return nil
	}
	out := make([]int64, len(n.buildingIDs))
	copy(out, n.buildingIDs)
	return out
}

// IsIsolated 表示该建筑未接入道路（距离退化为欧氏）。
func (n *RoadNetwork) IsIsolated(buildingID int64) bool {
	if n == nil {
		return true
	}
	if n.isolated[buildingID] {
		return true
	}
	_, known := n.entries[buildingID]
	return !known // 布局中不存在的楼栋同样视为未接入路网
}

// roadNetCache 缓存"区域概览 JSON -> 路网"，布局保存后 JSON 变化即自动重建。
type roadNetCache struct {
	mu         sync.RWMutex
	layoutJSON string
	net        *RoadNetwork
}

// network 读取默认区域概览并返回路网；无布局时返回 nil（调用方回退欧氏距离）。
func (c *roadNetCache) network(ctx context.Context, svcCtx *svc.ServiceContext) *RoadNetwork {
	layout, err := store.FindDefaultCampusLayout(ctx, svcCtx.DB)
	if err != nil || layout == nil || !layout.LayoutJson.Valid || layout.LayoutJson.String == "" {
		return nil
	}
	raw := layout.LayoutJson.String
	c.mu.RLock()
	cached, cachedRaw := c.net, c.layoutJSON
	c.mu.RUnlock()
	if cached != nil && cachedRaw == raw {
		return cached
	}
	built := buildRoadNetwork(raw, defaultGridMeters)
	c.mu.Lock()
	c.layoutJSON, c.net = raw, built
	c.mu.Unlock()
	return built
}

// DistancePair 两栋建筑间的路网距离（米）。
type DistancePair struct {
	FromID int64
	ToID   int64
	Meters float64
}

// Pairs 返回所有可计算的路网距离对（from < to，去重）。
func (n *RoadNetwork) Pairs() []DistancePair {
	if n == nil {
		return nil
	}
	pairs := make([]DistancePair, 0)
	for _, from := range n.buildingIDs {
		row, ok := n.dist[from]
		if !ok {
			continue
		}
		for _, to := range n.buildingIDs {
			if to <= from {
				continue
			}
			if d, ok := row[to]; ok {
				pairs = append(pairs, DistancePair{FromID: from, ToID: to, Meters: d})
			}
		}
	}
	return pairs
}

// NearestMeters 返回该建筑到最远可达建筑的路网距离（用于展示"最远通勤距离"）。
func (n *RoadNetwork) NearestMeters(buildingID int64) float64 {
	if n == nil {
		return 0
	}
	best := 0.0
	for _, d := range n.dist[buildingID] {
		if d > best {
			best = d
		}
	}
	return best
}

// EntryCount 与道路相邻的入口格数量。
func (n *RoadNetwork) EntryCount(buildingID int64) int {
	if n == nil {
		return 0
	}
	return len(n.entries[buildingID])
}

// RoadCellCount 栅格路网的道路格数量。
func (n *RoadNetwork) RoadCellCount() int {
	if n == nil {
		return 0
	}
	return len(n.road)
}

// boxGapCells 计算格点到建筑矩形轮廓的曼哈顿距离（0 表示紧贴建筑）。
func boxGapCells(b campusBlock, r, c int) int {
	rowSpan, colSpan := spanOf(b)
	dr, dc := 0, 0
	if r < b.Row {
		dr = b.Row - r
	} else if r > b.Row+rowSpan-1 {
		dr = r - (b.Row + rowSpan - 1)
	}
	if c < b.Col {
		dc = b.Col - c
	} else if c > b.Col+colSpan-1 {
		dc = c - (b.Col + colSpan - 1)
	}
	return dr + dc
}

// GapCells 建筑到其接入道路格的步行格数（0 表示与道路紧贴）。
func (n *RoadNetwork) GapCells(buildingID int64) int {
	if n == nil {
		return 0
	}
	return n.gap[buildingID]
}