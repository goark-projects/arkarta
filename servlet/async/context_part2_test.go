package async

import (
	"io"
	"net/http"
	"sync"

	"goark.dev/arkarta/servlet"
)

func (r *asyncResponse) Reset() error {
	r.body.Reset()
	r.committed = false
	return nil
}

func (r *asyncResponse) BodyWriter() io.Writer {
	return r
}

type blockingAsyncResponse struct {
	asyncResponse
	writeStarted chan struct{}
	releaseWrite chan struct{}
	flushCalled  chan struct{}
	flushOnce    sync.Once
}

func newBlockingAsyncResponse() *blockingAsyncResponse {
	return &blockingAsyncResponse{
		asyncResponse: asyncResponse{
			header: servlet.NewHeader(),
			status: http.StatusOK,
		},
		writeStarted: make(chan struct{}),
		releaseWrite: make(chan struct{}),
		flushCalled:  make(chan struct{}),
	}
}

func (r *blockingAsyncResponse) Write(data []byte) (int, error) {
	close(r.writeStarted)
	<-r.releaseWrite
	return r.asyncResponse.Write(data)
}

func (r *blockingAsyncResponse) Flush() error {
	r.flushOnce.Do(func() {
		close(r.flushCalled)
	})
	return r.asyncResponse.Flush()
}
