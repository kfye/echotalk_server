// Package validatorx 把 gin/validator 的参数校验错误翻译为可读的中文提示。
package validatorx

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

// Message 返回第一条校验错误的可读提示；非校验类错误兜底「参数错误」。
func Message(err error) string {
	var ves validator.ValidationErrors
	if !errors.As(err, &ves) || len(ves) == 0 {
		return "参数错误"
	}
	fe := ves[0]
	field := fe.Field()
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s 不能为空", field)
	case "email":
		return fmt.Sprintf("%s 必须是有效邮箱地址", field)
	case "min":
		return fmt.Sprintf("%s 长度/值不能小于 %s", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s 长度/值不能大于 %s", field, fe.Param())
	case "len":
		return fmt.Sprintf("%s 长度必须为 %s", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("%s 必须是 [%s] 之一", field, fe.Param())
	default:
		return fmt.Sprintf("%s 校验失败(%s)", field, fe.Tag())
	}
}
