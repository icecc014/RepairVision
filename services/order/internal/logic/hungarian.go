package logic

// hungarianCostINF 表示不存在的边。
const hungarianCostINF int64 = 1 << 55

// hungarian 求解 n×n 最小代价指派，返回每行分配的列下标（0 基），
// 无匹配时返回 -1。算法为 KM/Hungarian O(n^3)，代价矩阵需为非负。
func hungarian(cost [][]int64) []int {
	n := len(cost)
	result := make([]int, n)
	for i := range result {
		result[i] = -1
	}
	if n == 0 {
		return result
	}

	u := make([]int64, n+1)
	v := make([]int64, n+1)
	p := make([]int, n+1) // p[j] 表示第 j 列匹配的行（1 基），0 表示空闲
	way := make([]int, n+1)

	for i := 1; i <= n; i++ {
		p[0] = i
		j0 := 0
		minv := make([]int64, n+1)
		for j := range minv {
			minv[j] = hungarianCostINF
		}
		used := make([]bool, n+1)
		for {
			used[j0] = true
			i0 := p[j0]
			delta := hungarianCostINF
			j1 := 0
			for j := 1; j <= n; j++ {
				if used[j] {
					continue
				}
				cur := cost[i0-1][j-1] - u[i0] - v[j]
				if cur < minv[j] {
					minv[j] = cur
					way[j] = j0
				}
				if minv[j] < delta {
					delta = minv[j]
					j1 = j
				}
			}
			for j := 0; j <= n; j++ {
				if used[j] {
					u[p[j]] += delta
					v[j] -= delta
				} else {
					minv[j] -= delta
				}
			}
			j0 = j1
			if p[j0] == 0 {
				break
			}
		}
		for {
			j1 := way[j0]
			p[j0] = p[j1]
			j0 = j1
			if j0 == 0 {
				break
			}
		}
	}

	for j := 1; j <= n; j++ {
		if p[j] > 0 {
			result[p[j]-1] = j - 1
		}
	}
	return result
}
