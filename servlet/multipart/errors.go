package multipart

import "errors"

// ErrNotMultipart 表示请求不是 multipart 类型。
var ErrNotMultipart = errors.New("arkarta/servlet/multipart: request is not multipart")

// ErrBodyTooLarge 表示请求体超过配置上限。
var ErrBodyTooLarge = errors.New("arkarta/servlet/multipart: body too large")

// ErrFileTooLarge 表示单个上传文件超过配置上限。
var ErrFileTooLarge = errors.New("arkarta/servlet/multipart: file too large")
