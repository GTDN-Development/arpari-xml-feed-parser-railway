package feed

import (
	"context"
	"io"

	"github.com/fanda/arpari-xml-feed-parser-railway/internal/shoptet"
)

type Hello struct{}

func (Hello) Name() string {
	return "hello"
}

func (Hello) Filename() string {
	return "hello.xml"
}

func (Hello) Generate(_ context.Context, w io.Writer) (Result, error) {
	feed := shoptet.Feed{
		Items: []shoptet.Item{
			{
				Code:     "HELLO-001",
				Name:     "Hello world product",
				PriceVAT: "123.45",
				Stock:    "7",
			},
		},
	}
	if err := shoptet.Write(w, feed); err != nil {
		return Result{}, err
	}

	return Result{ItemsProcessed: 1}, nil
}
