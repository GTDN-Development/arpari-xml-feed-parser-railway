package feed

import (
	"context"
	"fmt"
	"io"
)

type Generator interface {
	Name() string
	Filename() string
	Generate(ctx context.Context, w io.Writer) (Result, error)
}

type Result struct {
	ItemsProcessed int
	ItemsSkipped   int
}

func Find(name string) (Generator, error) {
	for _, generator := range All() {
		if generator.Name() == name {
			return generator, nil
		}
	}

	return nil, fmt.Errorf("unknown supplier %q", name)
}

func All() []Generator {
	return []Generator{
		Hello{},
		StimaProducts{},
		StimaStock{},
		StimaStockPrice{},
	}
}
