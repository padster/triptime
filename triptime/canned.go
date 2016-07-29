package triptime

import (
    "strings"

    "github.com/padster/triptime/fb"

    ctx "golang.org/x/net/context"
)

// A simplified response for when the user replies with text we don't know.
func cannedResponseAction(c ctx.Context, msg fb.Message, fromUser string) fb.OutboundMessage {
    return fb.OutboundMessage{
        msg.Sender,
        outMessageDataFromText(cannedResponses(fromUser)),
    }
}

func cannedResponses(msg string) string {
  msg = strings.ToLower(msg)
  if msg == "thanks" || msg == "thanks!" || msg == "cheers" || msg == "cheers!" {
    return "No problem, enjoy the trip."
  } else if msg == "hi" || msg == "hi!" || msg == "hello" || msg == "hello!" || msg == "hey" || msg == "hey!" {
    return "G'day, I'm TripTime!\nIf you're looking for transport details, send me your location via the map."
  } else if msg == "ty" || msg == "ta" {
    return "np 🤘"
  }
  return "OK"
}
