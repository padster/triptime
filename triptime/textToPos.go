package triptime

import (
    // "fmt"

    "github.com/padster/triptime/fb"

    ctx "golang.org/x/net/context"
)

// Given text, use google's API to convert it to lat/long
func maybeTextToPosition(c ctx.Context, msg fb.Message, text string) *fb.Coordinates {
  // TODO
  return nil
}
