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
		StimaProductsTest{},
		StimaMissingVariants{},
		StimaStock{},
		StimaStockPrice{},
		AutronicProducts{},
		AutronicProductsTest{},
		AutronicAvailability{},
		Drevocal{},
		DrevocalTest{},
		Sakypaky{},
		SakypakyTest{},
		Sego{},
		SegoTest{},
		Hon{},
		HonTest{},
	}
}

func Scheduled() []Generator {
	return []Generator{
		StimaProducts{},
		StimaStock{},
		StimaStockPrice{},
		AutronicProducts{},
		AutronicAvailability{},
		Drevocal{},
		DrevocalTest{},
		Sakypaky{},
		SakypakyTest{},
		Sego{},
		Hon{},
	}
}
