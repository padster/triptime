// 1) Indexing in data code
// 2) Remove trip if it's the last stop
// 3) Add message if not stops remain
// 4) Arrive before TT

package triptime

import (
	"bytes"
	"fmt"
	// "log"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	// "time"

	"github.com/padster/triptime/fb"

	ctx "golang.org/x/net/context"
	gae "google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

func init() {
	http.HandleFunc("/_/verify", verifyHandler)
	http.HandleFunc("/policy.txt", policyHandler)
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	c := gae.NewContext(r)
	decoder := json.NewDecoder(r.Body)

	var data fb.RequestBody
	err := decoder.Decode(&data)

	if err != nil {
		log.Errorf(c, "Handler error: %+v", err)
		panic(err)
	}

	log.Infof(c, "RECV: %+v", data)

	for _, entry := range data.Entry {
		for _, msg := range entry.Message {
			handleMessage(c, entry, msg)
		}
	}
}

// Copy from https://github.com/ippy04/messengerbot/blob/master/webhook.go ?

func handleMessage(c ctx.Context, e fb.Entry, msg fb.Message) {
	if msg.Delivery != nil {
		log.Infof(c, "Ignoring delivery message")
		return
	}

	if msg.Postback != nil {
		log.Infof(c, "RECV pb: %s", msg.Postback.Payload)
		parts := strings.Split(msg.Postback.Payload, "/")
		switch parts[0] {
		case "liststops":
			handleListStops(c, msg, parts[1], parts[2])
		default:
			sendResponse(c, fb.OutboundMessage{
				msg.Sender,
				outMessageDataFromText("Ummm...I'm confused?"),
			})
		}
		return
	}

	pos := getCoordinates(msg.Message)
	if pos != nil {
		handleNextTrainRequest(c, msg.Sender, pos)
	} else {
		lowerText := strings.ToLower(strings.TrimSpace(msg.Message.Text))
		if lowerText == "help" || lowerText == "commands" {
			sendResponse(c, helpAction(c, msg))
		} else if strings.HasPrefix(lowerText, "next") {
			sendResponse(c, nextNLeavesAction(c, msg, strings.TrimSpace(lowerText[4:])))
		} else {
			cannedResponse := cannedResponseAction(c, msg, lowerText)
			if cannedResponse != nil {
				sendResponse(c, *cannedResponse)
			}
			posFromText := maybeTextToPosition(c, msg, lowerText)
			if posFromText != nil {
				handleNextTrainRequest(c, msg.Sender, posFromText)
			} else {
				sendResponse(c, helpAction(c, msg))
			}
		}
	}
}

func handleListStops(c ctx.Context, msg fb.Message, stopId string, tripId string) {
	text := ""
	tripsAdded := 0
	messageSent := false

	stopTimes := SortedStopTimesForTrip(tripId)
	found := false
	for _, stopTime := range stopTimes {
		if stopTime.StopId == stopId {
			found = true
		}
		if found {
			shortName := shortStopName(GetStop(stopTime.StopId).Name)
			text += fmt.Sprintf("🕓 %s - %s\n", normalizeTime(stopTime.Arrival), shortName)
			tripsAdded++
			// Chunk every 8 to avoid hitting message limit.
			if tripsAdded == 8 {
				tripsAdded = 0
				messageSent = true
				sendResponse(c, fb.OutboundMessage{
					msg.Sender,
					outMessageDataFromText(text),
				})
				text = ""
			}
		}
	}
	if text == "" && messageSent == false {
		text = "I got distracted, sorry... Please ask again."
	}
	if text != "" {
		sendResponse(c, fb.OutboundMessage{
			msg.Sender,
			outMessageDataFromText(text),
		})
	}
}

func handleNextTrainRequest(c ctx.Context, user fb.User, pos *fb.Coordinates) {
	t := getSFTime()

	closest := ClosestStop(pos)
	SetUserState(c, user.Id, UserState{
		*pos,
		closest,
	})
	closestPos := &fb.Coordinates{closest.Lat, closest.Long}
	distToClosestKM := CoordDistKM(closestPos, pos)
	nextTrips := NextTripsAtStop(closest)

	text :=
		fmt.Sprintf("Current time: %s\n", t.Format("15:04")) +
			fmt.Sprintf("Closest stop: %s (%0.2fkm away)\n", closest.Name, distToClosestKM)
	for _, trip := range nextTrips {
		routeName := GetRoute(trip.Trip.RouteId).LongName
		text += fmt.Sprintf(" 🚆 Next %s: %s (%s)\n", trip.Stop.PlatCode, trip.StopTime.Arrival, routeName)
	}

	response := buttonPayload(text)
	for _, trip := range nextTrips {
		caption := fmt.Sprintf("View %s stops", trip.Stop.PlatCode)
		payload := strings.Join([]string{"liststops", trip.Stop.StopId, trip.Trip.TripId}, "/")
		response.AddButton(callbackButton(caption, payload))
	}
	response.AddButton(urlButton("Directions", mapsDirections(pos, closestPos)))

	atch := templateAttachment(response)
	sendResponse(c, fb.OutboundMessage{
		user,
		outMessageDataFromAttachment(&atch),
	})
}

// TODO - move these utilities into FB package
func outMessageDataFromText(text string) fb.OutMessageData {
	return fb.OutMessageData{text, nil}
}

func outMessageDataFromAttachment(atch *fb.OutAttachment) fb.OutMessageData {
	return fb.OutMessageData{"", atch}
}

func templateAttachment(payload fb.AttachmentPayload) fb.OutAttachment {
	return fb.OutAttachment{"template", payload}
}

func getCoordinates(msg fb.MessageData) *fb.Coordinates {
	if msg.Attachment == nil {
		return nil
	}
	for _, attachment := range msg.Attachment {
		if attachment.Type == "location" {
			return &attachment.Payload.Coordinates
		}
	}
	return nil
}

func sendResponse(c ctx.Context, msg fb.OutboundMessage) {
	// POST to https://graph.facebook.com/v2.6/me/messages?access_token=
	asJson, err := json.Marshal(msg)
	if err != nil {
		log.Errorf(c, "Invalid outbound message: %+v", msg)
		panic("Can't marshal json")
	}
	log.Infof(c, "Sending: %s", asJson)

	if !gae.IsDevAppServer() {
		client := urlfetch.Client(c)
		r, e := client.Post(SEND_URL, "application/json", bytes.NewBuffer(asJson))
		if e != nil {
			log.Errorf(c, "Send error: %+v", e)
			panic("Can't send")
		}
		response, _ := ioutil.ReadAll(r.Body)
		log.Infof(c, "received back: %s", string(response))
	} else {
		log.Infof(c, "...or not, skipping on dev app server")
	}
}

func mapsDirections(from *fb.Coordinates, to *fb.Coordinates) string {
	return fmt.Sprintf(
		"https://www.google.com.au/maps/dir/%.7f,%.7f/'%.7f,%.7f'",
		from.Lat, from.Long, to.Lat, to.Long)
}

func buttonPayload(text string) fb.ButtonPayload {
	return fb.ButtonPayload{
		"button",
		text,
		[]fb.Button{},
	}
}

func urlButton(caption string, url string) fb.Button {
	return fb.Button{
		"web_url",
		caption,
		url,
		"",
	}
}

func callbackButton(caption string, payload string) fb.Button {
	return fb.Button{
		"postback",
		caption,
		"",
		payload,
	}
}

func shortStopName(name string) string {
	pos := strings.LastIndex(name, " Caltrain")
	if pos != -1 {
		return name[:pos]
	} else {
		return name
	}
}

// Serve user data policy static page
func policyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(`TripTime messenger bot privacy policy:

Information provided by each user is stored for only 10 minutes 
  in order to remember it during the conversation with the bot.
Location information (if provided) is used to find local time and nearby transport only, 
  and accessible only to bot in that 10 minutes, 
  not shared with others. 
No other details about users are handled.
    `))
}
