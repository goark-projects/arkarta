package registration

import "errors"

// ErrNilRegistry 表示注册项没有归属的注册表。
var ErrNilRegistry = errors.New("arkarta/servlet/registration: registry is nil")

// ErrRegistryFrozen 表示注册表已经冻结，不能再修改。
var ErrRegistryFrozen = errors.New("arkarta/servlet/registration: registry is frozen")

// ErrInvalidName 表示 Servlet、Filter 或 Listener 名称非法。
var ErrInvalidName = errors.New("arkarta/servlet/registration: invalid name")

// ErrInvalidInitParamName 表示初始化参数名称非法。
var ErrInvalidInitParamName = errors.New("arkarta/servlet/registration: invalid init parameter name")

// ErrInvalidDispatcherTypes 表示 DispatcherType 位集合非法。
var ErrInvalidDispatcherTypes = errors.New("arkarta/servlet/registration: invalid dispatcher types")

// ErrDuplicateRegistration 表示同名注册项已经存在。
var ErrDuplicateRegistration = errors.New("arkarta/servlet/registration: duplicate registration")

// ErrNilServlet 表示 Servlet 实例为空。
var ErrNilServlet = errors.New("arkarta/servlet/registration: servlet is nil")

// ErrNilFilter 表示 Filter 实例为空。
var ErrNilFilter = errors.New("arkarta/servlet/registration: filter is nil")

// ErrNilListener 表示 Listener 实例为空或类型不受支持。
var ErrNilListener = errors.New("arkarta/servlet/registration: listener is nil")
