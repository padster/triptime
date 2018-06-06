package triptime

import (
	// "fmt"

  "googlemaps.github.io/maps"

	"github.com/padster/triptime/fb"

	ctx "golang.org/x/net/context"
  "google.golang.org/appengine/log"
  "google.golang.org/appengine/urlfetch"
)

// Given text, use google's API to convert it to lat/long
func maybeTextToPosition(c ctx.Context, msg fb.Message, text string) *fb.Coordinates {
  log.Infof(c, "Converting \"%s\" to position", text)

  log.Infof(c, "Opening client...")
  client, err := maps.NewClient(
    maps.WithAPIKey(MAPS_API_KEY),
    maps.WithHTTPClient(urlfetch.Client(c)),
  )

  if err != nil {
      log.Errorf(c, "fatal error: %s", err)
      return nil
  }
  r := &maps.GeocodingRequest{
      Address: text,
  }
  log.Infof(c, "Sending request...")
  resp, err := client.Geocode(c, r)
  if err != nil {
    // Err status = ZERO_RESULTS when no location can be ascertained.
    return nil
  }

  if len(resp) == 0 {
    return nil
  }
  log.Infof(c, "RECV: %v", resp)
  return &fb.Coordinates{
    Lat:  resp[0].Geometry.Location.Lat,
    Long: resp[0].Geometry.Location.Lng,
  }
}
