package rebuild

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/feed"
	runstatus "github.com/fanda/arpari-xml-feed-parser-railway/internal/status"
	"github.com/fanda/arpari-xml-feed-parser-railway/internal/storage"
)

var (
	ErrAlreadyRunning  = errors.New("rebuild already running")
	ErrUnknownSupplier = errors.New("unknown supplier")
)

type Result struct {
	Supplier       string `json:"supplier"`
	Filename       string `json:"filename"`
	LastRunAt      string `json:"lastRunAt,omitempty"`
	Status         string `json:"status"`
	ItemsProcessed int    `json:"itemsProcessed"`
	ItemsSkipped   int    `json:"itemsSkipped"`
	Error          string `json:"error,omitempty"`
}

type Runner struct {
	DataDir string

	mu       sync.Mutex
	running  map[string]struct{}
	statusMu sync.Mutex
}

func NewRunner(dataDir string) *Runner {
	return &Runner{
		DataDir: dataDir,
		running: map[string]struct{}{},
	}
}

func (runner *Runner) RunName(ctx context.Context, supplier string) (Result, error) {
	generator, err := feed.Find(supplier)
	if err != nil {
		return Result{
			Supplier: supplier,
			Status:   runstatus.Failed,
			Error:    err.Error(),
		}, fmt.Errorf("%w: %s", ErrUnknownSupplier, supplier)
	}

	return runner.Run(ctx, generator)
}

func (runner *Runner) RunScheduled(ctx context.Context) []Result {
	generators := feed.Scheduled()
	results := make([]Result, 0, len(generators))
	for _, generator := range generators {
		result, err := runner.Run(ctx, generator)
		if err != nil && result.Error == "" {
			result.Error = err.Error()
		}
		results = append(results, result)
	}
	return results
}

func (runner *Runner) Run(ctx context.Context, generator feed.Generator) (Result, error) {
	unlock, ok := runner.tryLock(generator.Name())
	if !ok {
		result := Result{
			Supplier: generator.Name(),
			Filename: generator.Filename(),
			Status:   "running",
			Error:    ErrAlreadyRunning.Error(),
		}
		return result, ErrAlreadyRunning
	}
	defer unlock()

	publisher := storage.NewPublisher(runner.DataDir)
	statusStore := runstatus.NewStore(runner.DataDir)

	var generated feed.Result
	err := publisher.Publish(generator.Filename(), func(w io.Writer) error {
		var generateErr error
		generated, generateErr = generator.Generate(ctx, w)
		return generateErr
	})
	if err != nil {
		feedStatus := runstatus.NewFeedStatus(generator.Filename(), runstatus.Failed, generated.ItemsProcessed, generated.ItemsSkipped, err.Error(), time.Now())
		result := resultFromStatus(generator.Name(), feedStatus)
		if statusErr := runner.updateStatus(statusStore, generator.Name(), feedStatus); statusErr != nil {
			return result, errors.Join(err, fmt.Errorf("write failed feed status: %w", statusErr))
		}
		return result, err
	}

	feedStatus := runstatus.NewFeedStatus(generator.Filename(), runstatus.Success, generated.ItemsProcessed, generated.ItemsSkipped, "", time.Now())
	result := resultFromStatus(generator.Name(), feedStatus)
	if err := runner.updateStatus(statusStore, generator.Name(), feedStatus); err != nil {
		result.Error = err.Error()
		return result, err
	}

	return result, nil
}

func (runner *Runner) tryLock(supplier string) (func(), bool) {
	runner.mu.Lock()
	defer runner.mu.Unlock()

	if _, exists := runner.running[supplier]; exists {
		return nil, false
	}

	runner.running[supplier] = struct{}{}
	return func() {
		runner.mu.Lock()
		defer runner.mu.Unlock()
		delete(runner.running, supplier)
	}, true
}

func (runner *Runner) updateStatus(statusStore runstatus.Store, supplier string, feedStatus runstatus.FeedStatus) error {
	runner.statusMu.Lock()
	defer runner.statusMu.Unlock()
	return statusStore.Update(supplier, feedStatus)
}

func resultFromStatus(supplier string, feedStatus runstatus.FeedStatus) Result {
	return Result{
		Supplier:       supplier,
		Filename:       feedStatus.Filename,
		LastRunAt:      feedStatus.LastRunAt,
		Status:         feedStatus.Status,
		ItemsProcessed: feedStatus.ItemsProcessed,
		ItemsSkipped:   feedStatus.ItemsSkipped,
		Error:          feedStatus.Error,
	}
}
