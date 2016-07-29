package triptime

import (
    "fmt"
    // "strconv"
    // "strings"

    "github.com/padster/triptime/fb"


    ctx "golang.org/x/net/context"
    // "google.golang.org/appengine/log"
)

// Given a location, tell the times of the next N trains leaving from this station.
func helpAction(c ctx.Context, msg fb.Message) fb.OutboundMessage {
    state := GetUserState(c, msg.Sender.Id)
    if state == nil {
      return welcomeMessage(msg)
    } else {
      return stateMessage(msg, state)
    }
}

func welcomeMessage(msg fb.Message) fb.OutboundMessage {
  text := "Welcome to Triptime 🚆\n"
  text += "I don't know where you are, so first please send me your location using the marker ⇓\n"
  return fb.OutboundMessage{
      msg.Sender,
      outMessageDataFromText(text),
  }
}

func stateMessage(msg fb.Message, state *UserState) fb.OutboundMessage {
  text := fmt.Sprintf("You're near %s, so use 'Next' to see the next trains near you.\n", state.StopAt.Name)
  text += "You can state how many to see, and which direction you want - e.g. Next 5 NB\n"
  return fb.OutboundMessage{
      msg.Sender,
      outMessageDataFromText(text),
  }
}

