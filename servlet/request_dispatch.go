package servlet

// RunWithDispatchType 在指定分发类型下执行函数并恢复请求状态。
func (r *Request) RunWithDispatchType(dispatchType DispatchType, fn func() error) error {
	snapshot := r.dispatchSnapshot()
	r.applyDispatch(snapshot.path, snapshot.queryString, dispatchType)
	defer r.restoreDispatch(snapshot)
	if fn == nil {
		return nil
	}
	return fn()
}
