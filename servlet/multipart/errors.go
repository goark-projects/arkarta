package multipart

import "errors"

// ErrNotMultipart 表示请求不是 multipart 类型。
var ErrNotMultipart = errors.New("arkarta/servlet/multipart: request is not multipart")

// ErrBodyTooLarge 表示请求体超过配置上限。
var ErrBodyTooLarge = errors.New("arkarta/servlet/multipart: body too large")
