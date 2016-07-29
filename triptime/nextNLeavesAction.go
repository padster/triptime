package triptime

import (
    "fmt"
    "strconv"
    "strings"

    "github.com/padster/triptime/fb"

    ctx "golang.org/x/net/context"
    // "google.golang.org/appengine/log"
)

const MAX_BUTTONS = 3 // Bot API limit

// Given a location, tell the times of the next N trains leaving from this station.
func nextNLeavesAction(c ctx.Context, msg fb.Message, nInput string) fb.OutboundMessage {
    var state *UserState
    var err *fb.OutboundMessage
    if state, err = NeedUserState(c, msg); err != nil {
      return *err
    }
    stopAt := state.StopAt

    parts := strings.Split(nInput, " ")
    if len(parts) > 2 {
      return nextNUsage(msg);
    }

    var n int64
    var parseError error
    direction := ""
    n, parseError = strconv.ParseInt(parts[0], 10, 32) 
    if parseError != nil {
      n = 2
      if len(parts) == 1 {
        direction = parts[0]
      } else {
        return nextNUsage(msg)
      }
    } else {
      if n < 1 || n > 8 {
        return nextNUsage(msg);
      }
      if len(parts) == 2 {
        direction = parts[1]
      }
    }
    direction = strings.ToUpper(direction)

    return nextNLeavesResult(c, msg, stopAt, direction, int(n))
}

func nextNLeavesResult(c ctx.Context, msg fb.Message, stopAt Stop, direction string, n int) fb.OutboundMessage {
    nextTrips := NextNTripsFromStop(stopAt, direction, n)
    t := getSFTime()

    directionMsg := ""
    if direction != "" {
      directionMsg = direction + " "
    }
    nStopMsg := fmt.Sprintf("%strain", directionMsg)
    if n > 1 {
      nStopMsg = fmt.Sprintf("%d %strains", n, directionMsg)
    }

    text := 
        fmt.Sprintf("Current time: %s\n", t.Format("15:04")) +
        fmt.Sprintf("Next %s from %s:\n", nStopMsg, stopAt.Name)
    for _, trip := range nextTrips {
      routeName := GetRoute(trip.Trip.RouteId).LongName
      codeMsg := ""
      if direction == "" {
        codeMsg = fmt.Sprintf("%s @ ", trip.Stop.PlatCode)
      }
      text += fmt.Sprintf(" 🚆 %s%s (%s)\n", codeMsg, trip.StopTime.Arrival, routeName)
    }

    response := buttonPayload(text)
    for i, trip := range nextTrips {
        caption := fmt.Sprintf("View %s @ %s stops", trip.Stop.PlatCode, trip.StopTime.Arrival)
        payload := strings.Join([]string{"liststops", trip.Stop.StopId, trip.Trip.TripId}, "/")
        if i < MAX_BUTTONS {
          response.AddButton(callbackButton(caption, payload))
        }
    }

    atch := templateAttachment(response)
    return fb.OutboundMessage{
        msg.Sender,
        outMessageDataFromAttachment(&atch),
    }
}

func nextNUsage(msg fb.Message) fb.OutboundMessage {
    usage := "Try in the form: Next [#trains] [NB/SB]\n"
    usage += "#trains defaults to 2, is at most 8 and omit NB/SB to get both"
    return fb.OutboundMessage{
        msg.Sender,
        outMessageDataFromText(usage),
    }
}
