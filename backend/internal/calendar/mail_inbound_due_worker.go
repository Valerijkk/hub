package calendar

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	calendarInboundMessageDueWorkerPollInterval          = 30 * time.Second
	calendarInboundMessageDueWorkerClaimBatchSize       = 50
	calendarInboundMessageDueWorkerConcurrency          = 4
	calendarInboundMessageDueWorkerItemTimeout          = 45 * time.Second
	calendarInboundMessageDueWorkerCleanupBatchSize     = 500
	calendarInboundMessageDueWorkerCleanupPollIntervals = 120
)

// CalendarInboundMessageProcessor processes one already-claimed inbound
// receipt. CalendarInboundMessageService satisfies this narrow worker port.
type CalendarInboundMessageProcessor interface {
	ProcessCalendarInboundMessage(context.Context, CalendarInboundMessageWork) error
}

// CalendarInboundMessageDueWorkerTimer bounds polling and item lifetime.
type CalendarInboundMessageDueWorkerTimer interface {
	Wait(context.Context, time.Duration) error
	WithTimeout(context.Context, time.Duration) (context.Context, context.CancelFunc)
}

// CalendarInboundMessageDueWorker recovers expired or scheduled inbound
// receipts without duplicating repository policy or service behavior.
type CalendarInboundMessageDueWorker struct {
	repository CalendarInboundMessageRepository
	processor  CalendarInboundMessageProcessor
	timer      CalendarInboundMessageDueWorkerTimer

	guard   sync.Mutex
	running bool
}

type calendarInboundMessageDueWorkerRuntimeTimer struct{}

var (
	_ CalendarInboundMessageProcessor      = (*CalendarInboundMessageService)(nil)
	_ CalendarInboundMessageDueWorkerTimer = calendarInboundMessageDueWorkerRuntimeTimer{}
)

// NewCalendarInboundMessageDueWorker constructs the fixed-policy recovery
// worker. Receipt retention stays owned by its repository.
func NewCalendarInboundMessageDueWorker(
	repository CalendarInboundMessageRepository,
	processor CalendarInboundMessageProcessor,
	timer CalendarInboundMessageDueWorkerTimer,
) (*CalendarInboundMessageDueWorker, error) {
	if nilCalendarInboundMessageRepository(repository) ||
		calendarInboundMessageDueWorkerDependencyNil(processor) ||
		calendarInboundMessageDueWorkerDependencyNil(timer) {
		return nil, ErrEventInvalid
	}
	return &CalendarInboundMessageDueWorker{
		repository: repository,
		processor:  processor,
		timer:      timer,
	}, nil
}

// NewCalendarInboundMessageDueWorkerTimer returns the fixed production timer.
func NewCalendarInboundMessageDueWorkerTimer() CalendarInboundMessageDueWorkerTimer {
	return calendarInboundMessageDueWorkerRuntimeTimer{}
}

func (calendarInboundMessageDueWorkerRuntimeTimer) Wait(ctx context.Context, duration time.Duration) error {
	if calendarInboundMessageDueWorkerDependencyNil(ctx) || duration <= 0 {
		return ErrEventInvalid
	}
	timer := time.NewTimer(duration)
	defer func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (calendarInboundMessageDueWorkerRuntimeTimer) WithTimeout(
	ctx context.Context,
	duration time.Duration,
) (context.Context, context.CancelFunc) {
	if calendarInboundMessageDueWorkerDependencyNil(ctx) || duration <= 0 {
		return nil, nil
	}
	return context.WithTimeout(ctx, duration)
}

// Run performs one immediate recovery cycle, then polls until the parent is
// cancelled. A second concurrent Run is rejected without affecting the first.
func (worker *CalendarInboundMessageDueWorker) Run(ctx context.Context) (resultErr error) {
	if !worker.ready() || calendarInboundMessageDueWorkerDependencyNil(ctx) {
		return ErrEventInvalid
	}
	if !worker.enter() {
		return ErrEventInvalid
	}
	defer func() {
		worker.leave()
		if recover() != nil {
			resultErr = ErrEventInvalid
		}
	}()

	for pollCycle := uint64(1); ; pollCycle++ {
		if ctx.Err() != nil {
			return nil
		}
		worker.runCycle(ctx)
		if ctx.Err() != nil {
			return nil
		}
		if pollCycle == 1 || (pollCycle-1)%calendarInboundMessageDueWorkerCleanupPollIntervals == 0 {
			worker.cleanup(ctx)
		}
		if ctx.Err() != nil {
			return nil
		}
		if err := worker.wait(ctx); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return ErrEventInvalid
		}
	}
}

func (worker *CalendarInboundMessageDueWorker) runCycle(ctx context.Context) {
	work, err, panicked := worker.claimDue(ctx)
	if panicked || err != nil || ctx.Err() != nil {
		return
	}
	batch, valid := calendarInboundMessageDueWorkerBatch(work)
	if !valid {
		return
	}
	worker.processBatch(ctx, batch)
}

func (worker *CalendarInboundMessageDueWorker) processBatch(
	ctx context.Context,
	batch []CalendarInboundMessageWork,
) {
	if len(batch) == 0 || ctx.Err() != nil {
		return
	}
	concurrency := calendarInboundMessageDueWorkerConcurrency
	if len(batch) < concurrency {
		concurrency = len(batch)
	}

	jobs := make(chan CalendarInboundMessageWork)
	var workers sync.WaitGroup
	workers.Add(concurrency)
	for index := 0; index < concurrency; index++ {
		go func() {
			defer workers.Done()
			for work := range jobs {
				if ctx.Err() != nil {
					continue
				}
				worker.processItem(ctx, work)
			}
		}()
	}

dispatch:
	for _, work := range batch {
		select {
		case <-ctx.Done():
			break dispatch
		case jobs <- work:
		}
	}
	close(jobs)
	workers.Wait()
}

func (worker *CalendarInboundMessageDueWorker) processItem(
	ctx context.Context,
	work CalendarInboundMessageWork,
) {
	defer func() {
		_ = recover()
	}()
	if ctx.Err() != nil {
		return
	}
	itemContext, cancel, valid := worker.withItemTimeout(ctx)
	if !valid {
		return
	}
	defer calendarInboundMessageDueWorkerCancel(cancel)
	if itemContext.Err() != nil {
		return
	}
	_, _ = worker.process(itemContext, work)
}

func (worker *CalendarInboundMessageDueWorker) claimDue(
	ctx context.Context,
) (work []CalendarInboundMessageWork, err error, panicked bool) {
	defer func() {
		if recover() != nil {
			work = nil
			err = nil
			panicked = true
		}
	}()
	work, err = worker.repository.ClaimDueCalendarInboundMessages(ctx, calendarInboundMessageDueWorkerClaimBatchSize)
	return work, err, false
}

func (worker *CalendarInboundMessageDueWorker) cleanup(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	_, _, _ = worker.cleanupFinalized(ctx)
}

func (worker *CalendarInboundMessageDueWorker) cleanupFinalized(
	ctx context.Context,
) (removed int, err error, panicked bool) {
	defer func() {
		if recover() != nil {
			removed = 0
			err = nil
			panicked = true
		}
	}()
	removed, err = worker.repository.CleanupFinalizedCalendarInboundMessages(ctx, calendarInboundMessageDueWorkerCleanupBatchSize)
	return removed, err, false
}

func (worker *CalendarInboundMessageDueWorker) withItemTimeout(
	ctx context.Context,
) (itemContext context.Context, cancel context.CancelFunc, valid bool) {
	defer func() {
		if recover() != nil {
			itemContext = nil
			cancel = nil
			valid = false
		}
	}()
	itemContext, cancel = worker.timer.WithTimeout(ctx, calendarInboundMessageDueWorkerItemTimeout)
	if calendarInboundMessageDueWorkerDependencyNil(itemContext) || cancel == nil {
		return nil, nil, false
	}
	return itemContext, cancel, true
}

func (worker *CalendarInboundMessageDueWorker) process(
	ctx context.Context,
	work CalendarInboundMessageWork,
) (err error, panicked bool) {
	defer func() {
		if recover() != nil {
			err = nil
			panicked = true
		}
	}()
	err = worker.processor.ProcessCalendarInboundMessage(ctx, work)
	return err, false
}

func (worker *CalendarInboundMessageDueWorker) wait(ctx context.Context) (err error) {
	defer func() {
		if recover() != nil {
			err = ErrEventInvalid
		}
	}()
	return worker.timer.Wait(ctx, calendarInboundMessageDueWorkerPollInterval)
}

func (worker *CalendarInboundMessageDueWorker) enter() bool {
	worker.guard.Lock()
	defer worker.guard.Unlock()
	if worker.running {
		return false
	}
	worker.running = true
	return true
}

func (worker *CalendarInboundMessageDueWorker) leave() {
	worker.guard.Lock()
	worker.running = false
	worker.guard.Unlock()
}

func (worker *CalendarInboundMessageDueWorker) ready() bool {
	return worker != nil &&
		!nilCalendarInboundMessageRepository(worker.repository) &&
		!calendarInboundMessageDueWorkerDependencyNil(worker.processor) &&
		!calendarInboundMessageDueWorkerDependencyNil(worker.timer)
}

func calendarInboundMessageDueWorkerBatch(
	work []CalendarInboundMessageWork,
) ([]CalendarInboundMessageWork, bool) {
	if len(work) > calendarInboundMessageDueWorkerClaimBatchSize {
		return nil, false
	}
	batch := make([]CalendarInboundMessageWork, 0, len(work))
	receiptIDs := make(map[uuid.UUID]struct{}, len(work))
	claims := make(map[CalendarInboundMessageClaim]struct{}, len(work))
	for _, value := range work {
		if !value.valid() {
			return nil, false
		}
		copied, err := NewCalendarInboundMessageWork(value.Registration(), value.Claim())
		if err != nil {
			return nil, false
		}
		claim := copied.Claim()
		if _, duplicate := receiptIDs[claim.ReceiptID]; duplicate {
			return nil, false
		}
		if _, duplicate := claims[claim]; duplicate {
			return nil, false
		}
		receiptIDs[claim.ReceiptID] = struct{}{}
		claims[claim] = struct{}{}
		batch = append(batch, copied)
	}
	return batch, true
}

func calendarInboundMessageDueWorkerCancel(cancel context.CancelFunc) {
	defer func() {
		_ = recover()
	}()
	if cancel != nil {
		cancel()
	}
}

func calendarInboundMessageDueWorkerDependencyNil(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
