package triptime

import (
    "math"
    "sort"
    "strings"
    "time"

    "github.com/padster/triptime/fb"

    // ctx "golang.org/x/net/context"
    // "google.golang.org/appengine/log"
)

const (
    RADIUS_KM = 6371.0 // Earth radius in m
)


type GTFSData struct {
    Routes                  []Route
    Stops                   []Stop
    StopTimes               []StopTime
    ServiceDates            []ServiceDate
    ServiceDateExceptions   []ServiceDateException
    Trips                   []Trip
}

var DATA = GTFSData{
    ReadRoutes(),
    ReadStops(),
    ReadStopTimes(),
    ReadServiceDates(),
    ReadServiceDateExceptions(),
    ReadTrips(),
}

type NextTripResult struct {
    StopTime    *StopTime
    Stop        Stop
    Trip        *Trip
}


func NextTripsAtStop(at Stop) []NextTripResult {
    t := getSFTime()
    sId := ServiceIdForToday(t)
    trips := TripsForServiceId(sId)
    
    allStops := DirectionalStops(at, "")
    result := []NextTripResult{}
    for _, stopAt := range allStops {
        nextStopTime := NextStopTime(stopAt, trips, t)
        if nextStopTime.StopTime != nil {
            result = append(result, nextStopTime)
        }
    }
    return result
}

func NextNTripsFromStop(at Stop, direction string, n int) []NextTripResult {
    t := getSFTime()
    sId := ServiceIdForToday(t)
    trips := TripsForServiceId(sId)
    allStops := DirectionalStops(at, direction)

    // insertion sort to find the best N
    bestTrip := make([]*Trip, n, n)
    bestStopTime := make([]*StopTime, n, n)
    stopsAt := make([]Stop, n, n)
    for _, stopAt := range allStops {
        for _, trip := range trips {
            stopTime := TimeForStopAndTrip(stopAt.StopId, trip.TripId)
            if stopTime == nil {
                continue
            }
            for i := 0; i < n; i++ {
                if isBetterStop(stopTime, bestStopTime[i], t) {
                    for j := n - 1; j > i; j-- {
                        bestTrip[j] = bestTrip[j - 1]
                        bestStopTime[j] = bestStopTime[j - 1]
                        stopsAt[j] = stopsAt[j - 1]
                    }
                    // log.Infof(c, "Inserting %q at %d", stopTime, i)

                    bestTrip[i] = trip
                    bestStopTime[i] = stopTime
                    stopsAt[i] = stopAt
                    break;
                }
            }
        }
    }

    resultCount := 0
    for ; resultCount < n && bestTrip[resultCount] != nil; {
        resultCount++
    }
    result := make([]NextTripResult, resultCount, resultCount)
    for i := 0; i < resultCount; i++ {
        result[i] = NextTripResult{
            bestStopTime[i], stopsAt[i], bestTrip[i],
        }
    }
    return result
}

func NextStopTime(stopAt Stop, trips []*Trip, t time.Time) NextTripResult {
    var bestTrip *Trip
    var bestStopTime *StopTime

    // var bestStopTime *StopTime
    for _, trip := range trips {
        stopTime := TimeForStopAndTrip(stopAt.StopId, trip.TripId)
        if isBetterStop(stopTime, bestStopTime, t) {
            bestTrip = trip
            bestStopTime = stopTime
        }

    }
    return NextTripResult {
        bestStopTime, stopAt, bestTrip,
    }
}

func ServiceIdForToday(t time.Time) string {
    asDate := dateAsString(t)
    for _, ex := range DATA.ServiceDateExceptions {
        if ex.ExceptionType == 1 && ex.Date == asDate {
            return ex.ServiceId
        }
    }
    day := t.Format("Mon")
    for _, sd := range DATA.ServiceDates {
        switch day {
        case "Mon": if sd.OnMon { return sd.ServiceId }
        case "Tue": if sd.OnTue { return sd.ServiceId }
        case "Wed": if sd.OnWed { return sd.ServiceId }
        case "Thu": if sd.OnThu { return sd.ServiceId }
        case "Fri": if sd.OnFri { return sd.ServiceId }
        case "Sat": if sd.OnSat { return sd.ServiceId }
        case "Sun": if sd.OnSun { return sd.ServiceId }
        }
    }
    panic("No service for " + asDate + "!")
}

func TripsForServiceId(id string) []*Trip {
    // TODO - index?
    trips := []*Trip{}
    for i, trip := range DATA.Trips {
        if trip.ServiceId == id {
            trips = append(trips, &DATA.Trips[i])
        }
    }
    return trips
}

func ClosestStop(at *fb.Coordinates) Stop {
    bestStop := -1
    bestDist := 0.0
    for i, stop := range DATA.Stops {
        if stop.Type == 1 { // Non-directional. 0 = directional
            stopAt := fb.Coordinates{stop.Lat, stop.Long}
            dist := CoordDistKM(at, &stopAt)
            if bestStop == -1 || dist < bestDist {
                bestStop = i
                bestDist = dist
            }
        }
    }
    return DATA.Stops[bestStop]
}

func DirectionalStops(stop Stop, direction string) []Stop {
    // TODO - index
    if stop.Type != 1 {
        panic("Must pass in non-directional stop")
    }
    stops := []Stop{}
    for _, child := range DATA.Stops {
        if child.Type == 0 && child.Parent == stop.StopId {
            if direction == "" || direction == child.PlatCode {
                stops = append(stops, child)
            }
        }
    }
    return stops
}

func TimeForStopAndTrip(stopId string, tripId string) *StopTime {
    // TODO - index
    for _, stopTime := range DATA.StopTimes {
        if stopTime.TripId == tripId && stopTime.StopId == stopId {
            return &stopTime
        }
    }
    return nil
}

func SortedStopTimesForTrip(tripId string) []StopTime {
    result := []StopTime{}
    for _, stopTime := range DATA.StopTimes {
        if stopTime.TripId == tripId {
            result = append(result, stopTime)
        }
    }
    sort.Sort(TimesByTime(result))
    return result
}

func GetStop(stopId string) *Stop {
    // TODO - index
    for _, stop := range DATA.Stops {
        if stop.StopId == stopId {
            return &stop
        }
    }
    return nil
}

func GetRoute(routeId string) *Route {
    // TODO - index
    for _, route := range DATA.Routes {
        if route.RouteId == routeId {
            return &route
        }
    }
    return nil
}

// http://stackoverflow.com/questions/27928/calculate-distance-between-two-latitude-longitude-points-haversine-formula
func CoordDistKM(a *fb.Coordinates, b *fb.Coordinates) float64 {
    dLat  := deg2rad(b.Lat  - a.Lat)
    dLong := deg2rad(b.Long - a.Long)
    x := math.Sin(dLat/2) * math.Sin(dLat/2) +
        math.Cos(deg2rad(a.Lat)) * math.Cos(deg2rad(b.Lat)) * 
        math.Sin(dLong/2) * math.Sin(dLong/2)
    y := 2.0 * math.Atan2(math.Sqrt(x), math.Sqrt(1-x))
    return RADIUS_KM * y
}

func deg2rad(deg float64) float64 {
    return deg * math.Pi / 180.0
}

func isBetterStop(curr *StopTime, best *StopTime, t time.Time) bool {
    if curr == nil {
        return false
    }
    untilCurr := secondsUntil(t, curr.Arrival)
    if untilCurr < 0 {
        return false
    }

    if best == nil {
        return true
    }
    untilBest := secondsUntil(t, best.Arrival)
    if untilBest < 0 {
        return true
    }

    return untilCurr < untilBest
}


func normalizeTime(timeFmt string) string {
    // H:MM:SS -> HH:MM:SS
    if len(timeFmt) == 7 {
        return "0" + timeFmt
    } else {
        return timeFmt
    }
}


/// Grr, Go 
type TimesByTime []StopTime
func (times TimesByTime) Len() int {
    return len(times)
}
func (times TimesByTime) Swap(i, j int) {
    times[i], times[j] = times[j], times[i]
}
func (times TimesByTime) Less(i, j int) bool {
    return strings.Compare(normalizeTime(times[i].Arrival), normalizeTime(times[j].Arrival)) < 0
}