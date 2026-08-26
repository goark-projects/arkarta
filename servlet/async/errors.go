package async

import "errors"

// ErrNilRequest 表示异步上下文缺少请求。
var ErrNilRequest = errors.New("arkarta/servlet/async: request is nil")

// ErrNilResponse 表示异步上下文缺少响应。
var ErrNilResponse = errors.New("arkarta/servlet/async: response is nil")

// ErrNilHandler 表示异步分发缺少处理器。
var ErrNilHandler = errors.New("arkarta/servlet/async: handler is nil")

// ErrCompleted 表示异步上下文已经完成。
var ErrCompleted = errors.New("arkarta/servlet/async: context completed")

// ErrTimeout 表示异步上下文超时。
var ErrTimeout = errors.New("arkarta/servlet/async: timeout")
