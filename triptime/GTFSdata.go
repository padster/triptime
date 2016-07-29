package triptime

import (
    "encoding/csv"
    "os"
    "reflect"
    "strconv"
)

// stop_id,stop_code,stop_name,stop_lat,stop_lon,zone_id,stop_url,location_type,parent_station,platform_code,wheelchair_boarding
type Stop struct {
    StopId      string
    Code        string
    Name        string
    Lat         float64
    Long        float64
    ZoneId      string
    StopUrl     string
    Type        int
    Parent      string
    PlatCode    string
    WChair      int
}

// trip_id,arrival_time,departure_time,stop_id,stop_sequence,pickup_type,drop_off_type
type StopTime struct {
    TripId      string
    Arrival     string
    Departure   string
    StopId      string
    StopSeq     int
    PickupType  string // ignore, always 0?
    DropoffType string // ignore, always 0?
}

// route_id,service_id,trip_id,trip_headsign,trip_short_name,direction_id,shape_id,wheelchair_accessible,bikes_allowed
type Trip struct {
    RouteId     string
    ServiceId   string
    TripId      string
    HeadSign    string
    ShortName   string
    DirectionId string
    ShapeId     string
    WChair      bool
    Bikes       bool
}

// service_id,monday,tuesday,wednesday,thursday,friday,saturday,sunday,start_date,end_date
type ServiceDate struct {
    ServiceId   string
    OnMon       bool
    OnTue       bool
    OnWed       bool
    OnThu       bool
    OnFri       bool
    OnSat       bool
    OnSun       bool
    StartDate   string // ignore
    EndDate     string // ignore
}

// service_id,date,exception_type
type ServiceDateException struct {
    ServiceId       string
    Date            string // YYYYMMDD
    ExceptionType   int    // 1 = new service, 2 = old service
}

// route_id,route_short_name,route_long_name,route_type,route_color
type Route struct {
    RouteId         string
    ShortName       string
    LongName        string
    Type            int
    Color           string
}

func ReadStops() []Stop {
    f, err := os.Open("gtfs/stops.txt")
    defer f.Close()
    if err != nil {
        panic("can't open file...")
    }
    reader := csv.NewReader(f)

    result := []Stop{}
    for i := 0; ; i++ {
        if i == 0 {
            // First row, ignore as it's column headers.
            reader.Read()
        } else {
            row := Stop{}
            if err := Unmarshal(reader, &row); err != "" {
                if err == "DONE" {
                    break
                }
                panic(err)
            }
            result = append(result, row)
        }
    }
    return result
}

func ReadStopTimes() []StopTime {
    f, err := os.Open("gtfs/stop_times.txt")
    defer f.Close()
    if err != nil {
        panic("can't open file...")
    }
    reader := csv.NewReader(f)

    result := []StopTime{}
    for i := 0; ; i++ {
        if i == 0 {
            // First row, ignore as it's column headers.
            reader.Read()
        } else {
            row := StopTime{}
            if err := Unmarshal(reader, &row); err != "" {
                if err == "DONE" {
                    break
                }
                panic(err)
            }
            result = append(result, row)
        }
    }
    return result
}

func ReadTrips() []Trip {
    f, err := os.Open("gtfs/trips.txt")
    defer f.Close()
    if err != nil {
        panic("can't open file...")
    }
    reader := csv.NewReader(f)

    result := []Trip{}
    for i := 0; ; i++ {
        if i == 0 {
            // First row, ignore as it's column headers.
            reader.Read()
        } else {
            row := Trip{}
            if err := Unmarshal(reader, &row); err != "" {
                if err == "DONE" {
                    break
                }
                panic(err)
            }
            result = append(result, row)
        }
    }
    return result
}


func ReadServiceDates() []ServiceDate {
    f, err := os.Open("gtfs/calendar.txt")
    defer f.Close()
    if err != nil {
        panic("can't open file...")
    }
    reader := csv.NewReader(f)

    result := []ServiceDate{}
    for i := 0; ; i++ {
        if i == 0 {
            // First row, ignore as it's column headers.
            reader.Read()
        } else {
            row := ServiceDate{}
            if err := Unmarshal(reader, &row); err != "" {
                if err == "DONE" {
                    break
                }
                panic(err)
            }
            result = append(result, row)
        }
    }
    return result
}

func ReadServiceDateExceptions() []ServiceDateException {
    f, err := os.Open("gtfs/calendar_dates.txt")
    defer f.Close()
    if err != nil {
        panic("can't open file...")
    }
    reader := csv.NewReader(f)

    result := []ServiceDateException{}
    for i := 0; ; i++ {
        if i == 0 {
            // First row, ignore as it's column headers.
            reader.Read()
        } else {
            row := ServiceDateException{}
            if err := Unmarshal(reader, &row); err != "" {
                if err == "DONE" {
                    break
                }
                panic(err)
            }
            result = append(result, row)
        }
    }
    return result
}

func ReadRoutes() []Route {
    f, err := os.Open("gtfs/routes.txt")
    defer f.Close()
    if err != nil {
        panic("can't open file...")
    }
    reader := csv.NewReader(f)

    result := []Route{}
    for i := 0; ; i++ {
        if i == 0 {
            // First row, ignore as it's column headers.
            reader.Read()
        } else {
            row := Route{}
            if err := Unmarshal(reader, &row); err != "" {
                if err == "DONE" {
                    break
                }
                panic(err)
            }
            result = append(result, row)
        }
    }
    return result
}


func Unmarshal(reader *csv.Reader, v interface{}) string {
    record, err := reader.Read()
    if err != nil {
        return "DONE"
    }
    s := reflect.ValueOf(v).Elem()
    if s.NumField() != len(record) {
        return "Field mismatch"
    }
    for i := 0; i < s.NumField(); i++ {
        f := s.Field(i)
        switch f.Type().String() {
        case "string":
            f.SetString(record[i])
        case "int":
            ival, err := strconv.ParseInt(record[i], 10, 0)
            if err != nil {
                return "Can't parse int"
            }
            f.SetInt(ival)
        case "float64":
            fval, err := strconv.ParseFloat(record[i], 64)
            if err != nil {
                return "Can't parse float"
            }
            f.SetFloat(fval)
        case "bool":
            f.SetBool(record[i] == "1")
        default:
            return "Unknown type: " + f.Type().String()
        }
    }
    return ""
}

/*

var GTFS_STOPS = [...]Stop{
    Stop{"San Francisco",     Coordinates{37.776439,   -122.394323}},
    Stop{"22nd St",           Coordinates{37.757674,   -122.392636}},
    Stop{"Bayshore",          Coordinates{37.709544,   -122.401318}},
    Stop{"So. San Francisco", Coordinates{37.655852,   -122.405429}},
    Stop{"San Bruno",         Coordinates{37.631106,   -122.412018}},
    Stop{"Millbrae",          Coordinates{37.600006,   -122.386534}},
    Stop{"Broadway",          Coordinates{37.587466,   -122.363233}},
    Stop{"Burlingame",        Coordinates{37.579719,   -122.345266}},
    Stop{"San Mateo",         Coordinates{37.568209,   -122.323933}},
    Stop{"Hayward Park",      Coordinates{37.552346,   -122.308916}},
    Stop{"Hillsdale",         Coordinates{37.537503,   -122.298001}},
    Stop{"Belmont",           Coordinates{37.520504,   -122.276075}},
    Stop{"San Carlos",        Coordinates{37.507361,   -122.260365}},
    Stop{"Redwood City",      Coordinates{37.485412,   -122.231957}},
    Stop{"Atherton",          Coordinates{37.464349,   -122.198106}},
    Stop{"Menlo Park",        Coordinates{37.454604,   -122.182518}},
    Stop{"Palo Alto",         Coordinates{37.443070,   -122.164900}},
    Stop{"California Ave",    Coordinates{37.428835,   -122.142703}},
    Stop{"San Antonio",       Coordinates{37.407157,   -122.107231}},
    Stop{"Mt View",           Coordinates{37.393879,   -122.076327}},
    Stop{"Sunnyvale",         Coordinates{37.378427,   -122.030742}},
    Stop{"Lawrence",          Coordinates{37.370815,   -121.997258}},
    Stop{"Santa Clara",       Coordinates{37.352915,   -121.936376}},
    Stop{"College Park",      Coordinates{37.342170,   -121.914998}},
    Stop{"San Jose Diridon",  Coordinates{37.329392,   -121.902181}},
    Stop{"Tamien",            Coordinates{37.311640,   -121.883900}},
    Stop{"Capitol",           Coordinates{37.284359,   -121.841589}},
    Stop{"Blossom Hill",      Coordinates{37.252801,   -121.797369}},
    Stop{"Morgan Hill",       Coordinates{37.129081,   -121.650721}},
    Stop{"San Martin",        Coordinates{37.085775,   -121.610809}},
    Stop{"Gilroy",            Coordinates{37.003606,   -121.566497}},
}
*/