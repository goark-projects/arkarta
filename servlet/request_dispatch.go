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

type dispatchSnapshot struct {
	path         string
	queryString  string
	dispatchType DispatchType
	servletPath  string
	pathInfo     string
	mapping      RequestMapping
}

func (r *Request) dispatchSnapshot() dispatchSnapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return dispatchSnapshot{
		path:         r.path,
		queryString:  r.queryString,
		dispatchType: r.dispatchType,
		servletPath:  r.servletPath,
		pathInfo:     r.pathInfo,
		mapping:      r.mapping,
	}
}

func (r *Request) applyDispatch(path, queryString string, dispatchType DispatchType) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.path = path
	r.queryString = queryString
	r.dispatchType = dispatchType
}

func (r *Request) restoreDispatch(snapshot dispatchSnapshot) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.path = snapshot.path
	r.queryString = snapshot.queryString
	r.dispatchType = snapshot.dispatchType
	r.servletPath = snapshot.servletPath
	r.pathInfo = snapshot.pathInfo
	r.mapping = snapshot.mapping
}

func (r *Request) applyMapping(mapping RequestMapping) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.servletPath = mapping.ServletPath()
	r.pathInfo = mapping.PathInfo()
	r.mapping = mapping
}
