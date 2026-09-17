package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"order/internal/svc"
	"order/internal/types"
)

type PublicCaptchaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPublicCaptchaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublicCaptchaLogic {
	return &PublicCaptchaLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// Captcha 生成公共报修页的图形验证码（SVG，5 分钟有效、一次性）。
func (l *PublicCaptchaLogic) Captcha() (*types.PublicCaptchaResponse, error) {
	code := randomCaptchaCode(publicCaptchaLen)
	id := randomCaptchaId()
	l.svcCtx.PubCache.Setex(l.ctx, publicCaptchaPrefix+id, code, publicCaptchaTTL)
	if !l.svcCtx.PubCache.UsingRedis() {
		l.Logger.Infof("public captcha stored in memory fallback (redis unavailable)")
	}
	return &types.PublicCaptchaResponse{
		CaptchaId: id,
		Svg:       captchaSVG(code),
		ExpiresIn: int(publicCaptchaTTL.Seconds()),
	}, nil
}
