package session

import (
	"errors"
	"sort"
)

// TrackingMode 表示会话 ID 的传输方式。
type TrackingMode string

const (
	// TrackingCookie 表示通过 Cookie 传输 Session ID。
	TrackingCookie TrackingMode = "COOKIE"
	// TrackingURL 表示通过 URL 路径参数传输 Session ID。
	TrackingURL TrackingMode = "URL"
	// TrackingSSL 表示由容器通过安全连接标识传输 Session ID。
	TrackingSSL TrackingMode = "SSL"
)

// ErrInvalidTrackingMode 表示会话跟踪模式非法。
var ErrInvalidTrackingMode = errors.New("arkarta/servlet/session: invalid tracking mode")

// TrackingPolicy 描述当前应用允许的会话跟踪模式。
type TrackingPolicy struct {
	modes map[TrackingMode]struct{}
}

// DefaultTrackingPolicy 返回 Servlet 生态默认的 Cookie 跟踪策略。
func DefaultTrackingPolicy() TrackingPolicy {
	policy, _ := NewTrackingPolicy(TrackingCookie)
	return policy
}

// NewTrackingPolicy 创建会话跟踪策略；未传入时默认使用 Cookie。
func NewTrackingPolicy(modes ...TrackingMode) (TrackingPolicy, error) {
	if len(modes) == 0 {
		modes = []TrackingMode{TrackingCookie}
	}
	policy := TrackingPolicy{modes: make(map[TrackingMode]struct{}, len(modes))}
	for _, mode := range modes {
		if !validTrackingMode(mode) {
			return TrackingPolicy{}, ErrInvalidTrackingMode
		}
		policy.modes[mode] = struct{}{}
	}
	return policy, nil
}

// Allows 判断策略是否允许指定跟踪模式。
func (p TrackingPolicy) Allows(mode TrackingMode) bool {
	if len(p.modes) == 0 {
		return mode == TrackingCookie
	}
	_, ok := p.modes[mode]
	return ok
}

// Modes 返回稳定排序的跟踪模式。
func (p TrackingPolicy) Modes() []TrackingMode {
	if len(p.modes) == 0 {
		return []TrackingMode{TrackingCookie}
	}
	result := make([]TrackingMode, 0, len(p.modes))
	for mode := range p.modes {
		result = append(result, mode)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})
	return result
}

func validTrackingMode(mode TrackingMode) bool {
	switch mode {
	case TrackingCookie, TrackingURL, TrackingSSL:
		return true
	default:
		return false
	}
}
