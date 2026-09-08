package errs

func Upstream() *Error {
	return New(500, "上游服务暂不可用，请稍后重试")
}
