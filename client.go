package main

import "context"

type Client interface {
	HelloWorld(ctx context.Context)
}
