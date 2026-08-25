package mail

import (
	"context"
	"io"
	"sync"

	"github.com/valerii/hub/backend/internal/mail/stalwart"
)

var _ stalwart.ContextReadCloser = (*mailboxDraftMIMEStream)(nil)
var _ io.ReadCloser = (*mailboxDraftMIMEActiveSource)(nil)

type mailboxDraftMIMEStream struct {
	plan    mailboxDraftMIMEPlan
	baseCtx context.Context

	reader *io.PipeReader
	writer *io.PipeWriter
	done   chan struct{}

	startOnce sync.Once
	closeOnce sync.Once

	mu       sync.Mutex
	cancel   context.CancelFunc
	active   *mailboxDraftMIMEActiveSource
	started  bool
	closed   bool
	startErr error
}

func newMailboxDraftMIMEStream(ctx context.Context, plan mailboxDraftMIMEPlan) (*mailboxDraftMIMEStream, int64, error) {
	if ctx == nil {
		return nil, 0, fakeEngineError(EngineErrorRejectedPolicy)
	}
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	if plan.calendarPresent {
		var calendarValid bool
		plan, calendarValid = mailboxDraftMIMEPlanWithResealedCalendar(plan)
		if !calendarValid {
			return nil, 0, fakeEngineError(EngineErrorRejectedPolicy)
		}
		pieces := make([]mailboxDraftMIMEPiece, len(plan.pieces))
		copy(pieces, plan.pieces)
		for index := range pieces {
			if pieces[index].body != nil {
				pieces[index].body = append([]byte(nil), pieces[index].body...)
			}
		}
		plan.pieces = pieces
	}
	if !mailboxDraftMIMEStreamPlanValid(plan) {
		return nil, 0, fakeEngineError(EngineErrorRejectedPolicy)
	}

	reader, writer := io.Pipe()
	return &mailboxDraftMIMEStream{
		plan:    plan,
		baseCtx: ctx,
		reader:  reader,
		writer:  writer,
		done:    make(chan struct{}),
	}, plan.size, nil
}

func mailboxDraftMIMEStreamPlanValid(plan mailboxDraftMIMEPlan) bool {
	return mailboxDraftMIMEPlanValid(plan)
}

func (stream *mailboxDraftMIMEStream) Read(buffer []byte) (int, error) {
	if stream == nil {
		return 0, io.ErrClosedPipe
	}
	return stream.ReadContext(stream.baseCtx, buffer)
}

func (stream *mailboxDraftMIMEStream) ReadContext(ctx context.Context, buffer []byte) (int, error) {
	if stream == nil || ctx == nil {
		return 0, io.ErrClosedPipe
	}
	if err := ctx.Err(); err != nil {
		_ = stream.Close()
		return 0, err
	}
	if len(buffer) == 0 {
		if stream.isClosed() {
			return 0, io.ErrClosedPipe
		}
		return 0, nil
	}
	if err := stream.start(ctx); err != nil {
		return 0, err
	}

	n, err := stream.reader.Read(buffer)
	if contextErr := ctx.Err(); contextErr != nil {
		_ = stream.Close()
		return n, contextErr
	}
	return n, err
}

func (stream *mailboxDraftMIMEStream) Close() error {
	if stream == nil {
		return nil
	}

	stream.closeOnce.Do(func() {
		stream.mu.Lock()
		stream.closed = true
		cancel := stream.cancel
		active := stream.active
		started := stream.started
		reader := stream.reader
		done := stream.done
		stream.mu.Unlock()

		if cancel != nil {
			cancel()
		}
		if reader != nil {
			_ = reader.CloseWithError(io.ErrClosedPipe)
		}
		if active != nil {
			_ = active.Close()
		}
		if started {
			<-done
		}
	})
	return nil
}

func (stream *mailboxDraftMIMEStream) isClosed() bool {
	stream.mu.Lock()
	defer stream.mu.Unlock()
	return stream.closed
}

type mailboxDraftMIMEActiveSource struct {
	stream *mailboxDraftMIMEStream
	reader io.ReadCloser

	closeOnce sync.Once
	err       error
}

func (source *mailboxDraftMIMEActiveSource) Read(buffer []byte) (int, error) {
	if source == nil || source.reader == nil {
		return 0, io.ErrClosedPipe
	}
	return source.reader.Read(buffer)
}

func (source *mailboxDraftMIMEActiveSource) Close() error {
	if source == nil || source.reader == nil {
		return nil
	}
	source.closeOnce.Do(func() {
		source.err = source.reader.Close()
		if source.stream != nil {
			source.stream.unregisterActiveSource(source)
		}
	})
	return source.err
}

func (stream *mailboxDraftMIMEStream) registerActiveSource(reader io.ReadCloser) (io.ReadCloser, error) {
	if stream == nil || reader == nil {
		return nil, io.ErrClosedPipe
	}
	active := &mailboxDraftMIMEActiveSource{stream: stream, reader: reader}

	stream.mu.Lock()
	if stream.closed || stream.active != nil {
		stream.mu.Unlock()
		_ = active.Close()
		return nil, io.ErrClosedPipe
	}
	stream.active = active
	stream.mu.Unlock()
	return active, nil
}

func (stream *mailboxDraftMIMEStream) unregisterActiveSource(active *mailboxDraftMIMEActiveSource) {
	stream.mu.Lock()
	defer stream.mu.Unlock()
	if stream.active == active {
		stream.active = nil
	}
}

func (stream *mailboxDraftMIMEStream) planWithActiveSources() mailboxDraftMIMEPlan {
	plan := stream.plan
	if plan.calendarPresent {
		var calendarValid bool
		plan, calendarValid = mailboxDraftMIMEPlanWithResealedCalendar(plan)
		if !calendarValid {
			return mailboxDraftMIMEPlan{}
		}
	}
	if plan.sources == nil {
		return plan
	}

	capability := *plan.sources
	capability.sources = make([]*mailboxDraftAttachmentSource, len(plan.sources.sources))
	for index, original := range plan.sources.sources {
		if original == nil {
			return plan
		}
		original := original
		capability.sources[index] = &mailboxDraftAttachmentSource{
			attachment: original.attachment,
			sizeBytes:  original.sizeBytes,
			scanStatus: original.scanStatus,
			open: func(ctx context.Context) (io.ReadCloser, error) {
				reader, err := original.openOnce(ctx)
				if err != nil {
					return nil, err
				}
				return stream.registerActiveSource(reader)
			},
		}
	}
	plan.sources = &capability
	return plan
}

func (stream *mailboxDraftMIMEStream) start(ctx context.Context) error {
	stream.startOnce.Do(func() {
		stream.mu.Lock()
		if stream.closed {
			stream.startErr = io.ErrClosedPipe
			stream.mu.Unlock()
			return
		}

		producerCtx, cancel := context.WithCancel(ctx)
		stream.cancel = cancel
		stream.started = true
		stream.mu.Unlock()

		producerPlan := stream.planWithActiveSources()
		go func() {
			err := producerPlan.writeTo(producerCtx, stream.writer)
			_ = stream.writer.CloseWithError(err)
			cancel()
			close(stream.done)
		}()
	})

	stream.mu.Lock()
	defer stream.mu.Unlock()
	return stream.startErr
}
