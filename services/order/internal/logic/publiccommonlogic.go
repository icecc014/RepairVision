package logic

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"map/mapclient"
	"order/internal/errs"
	"order/internal/store"
	"order/internal/svc"
)

// V7 公共报修共用逻辑：验证码、限流、敏感词、楼栋与房间校验。

const (
	publicCaptchaTTL      = 5 * time.Minute
	publicCaptchaLen      = 4
	publicCaptchaPrefix   = "captcha:"
	publicIPRatePrefix    = "rl:ip:"
	publicRoomRatePrefix  = "rl:room:"
	publicIPPerMinute     = 3
	publicRoomCooldown    = 30 * time.Minute
	publicMergeWindow     = 72 * time.Hour
	publicDescMin         = 5
	publicDescMax         = 200
	publicContactMax      = 32
	publicCaptchaAlphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	// publicFaultCodes 公开页面只暴露三类故障类型，映射到故障类型字典的 code。
)

var publicFaultCodes = []string{"electric", "water", "other"}

func publicFaultAllowed(code string) bool {
	for _, c := range publicFaultCodes {
		if c == code {
			return true
		}
	}
	return false
}

func randomCaptchaId() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	var b strings.Builder
	for i := 0; i < 12; i++ {
		b.WriteByte(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

func randomCaptchaCode(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteByte(publicCaptchaAlphabet[rand.Intn(len(publicCaptchaAlphabet))])
	}
	return b.String()
}

// captchaSVG 用 SVG 文本渲染验证码（无需图片字体依赖），带轻微抖动与干扰线。
func captchaSVG(code string) string {
	colors := []string{"#2462d9", "#1f2a3d", "#0f766e", "#b45309"}
	var b strings.Builder
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="120" height="40" viewBox="0 0 120 40">`)
	b.WriteString(`<rect width="120" height="40" rx="6" fill="#eef3fb"/>`)
	for i := 0; i < 3; i++ {
		fmt.Fprintf(&b, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-width="1" opacity="0.5"/>`,
			rand.Intn(20), rand.Intn(40), 100+rand.Intn(20), rand.Intn(40), colors[rand.Intn(len(colors))])
	}
	for i, ch := range code {
		x := 16 + i*26 + rand.Intn(6) - 3
		y := 28 + rand.Intn(8) - 4
		rotate := rand.Intn(30) - 15
		fmt.Fprintf(&b,
			`<text x="%d" y="%d" font-size="22" font-family="monospace" font-weight="700" fill="%s" transform="rotate(%d %d %d)">%c</text>`,
			x, y, colors[rand.Intn(len(colors))], rotate, x, y, ch)
	}
	b.WriteString(`</svg>`)
	return b.String()
}

// verifyPublicCaptcha 校验验证码（一次性：校验通过即删除）。
func verifyPublicCaptcha(ctx context.Context, svcCtx *svc.ServiceContext, id, input string) bool {
	id = strings.TrimSpace(id)
	input = strings.TrimSpace(input)
	if id == "" || input == "" {
		return false
	}
	key := publicCaptchaPrefix + id
	expect, ok := svcCtx.PubCache.Get(ctx, key)
	if !ok {
		return false
	}
	if !strings.EqualFold(expect, input) {
		return false
	}
	svcCtx.PubCache.Del(ctx, key)
	return true
}

// hitPublicSensitiveWord 命中敏感词时返回该词；读取失败不阻塞提交。
func hitPublicSensitiveWord(ctx context.Context, svcCtx *svc.ServiceContext, text string) string {
	words, err := store.ListSensitiveWords(ctx, svcCtx.DB)
	if err != nil {
		return ""
	}
	lower := strings.ToLower(text)
	for _, word := range words {
		if word == "" {
			continue
		}
		if strings.Contains(lower, strings.ToLower(word)) {
			return word
		}
	}
	return ""
}

// publicBuildingOf 取楼栋并按“楼层×100+序号”推导楼层、校验房间号。
func publicBuildingOf(ctx context.Context, svcCtx *svc.ServiceContext, buildingID int64, room string) (*mapclient.Building, int64, error) {
	if buildingID <= 0 {
		return nil, 0, errs.BadRequest("请选择报修楼栋")
	}
	resp, err := svcCtx.MapRpc.ListBuildings(ctx, &mapclient.ListBuildingsRequest{})
	if err != nil {
		return nil, 0, errs.Upstream()
	}
	var target *mapclient.Building
	for _, b := range resp.Buildings {
		if b.Id == buildingID {
			target = b
			break
		}
	}
	if target == nil {
		return nil, 0, errs.BadRequest("报修楼栋不存在")
	}
	room = strings.TrimSpace(room)
	num, err := strconv.Atoi(room)
	if err != nil || num <= 0 {
		return nil, 0, errs.BadRequest("房间号格式不正确（示例：401）")
	}
	floor := int64(num / 100)
	if floor <= 0 {
		return nil, 0, errs.BadRequest("房间号格式不正确（示例：401）")
	}
	if err := validateRoomForBuilding(room, floor, target); err != nil {
		return nil, 0, err
	}
	return target, floor, nil
}

// publicFaultNameOf 取故障类型显示名（公开三类之一）。
func publicFaultNameOf(ctx context.Context, svcCtx *svc.ServiceContext, code string) (string, error) {
	list, err := store.ListFaultTypes(ctx, svcCtx.DB)
	if err != nil {
		return "", errs.Internal(err)
	}
	for _, ft := range list {
		if ft.Code == code {
			return ft.Name, nil
		}
	}
	return "", errs.BadRequest("该故障类型暂未配置，请联系管理员")
}

// publicReporterText 报修人身份文案。
func publicReporterText(reporterType int64) string {
	switch reporterType {
	case 1:
		return "学生"
	case 2:
		return "教师"
	case 3:
		return "其他"
	default:
		return "用户"
	}
}
