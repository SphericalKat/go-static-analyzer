package main

import "context"

func main() {
	ctx, _ := context.WithCancel(context.Background())

	// just to discard ctx
	_ = ctx
}