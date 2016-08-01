package triptime

import (
	"fmt"
	"time"
)

func getSFTZ() *time.Location {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		panic("Don't have the LA timezone?!?! :(")
	}
	return loc
}

func getSFTime() time.Time {
	return time.Now().In(getSFTZ())
}

func dateAsString(t time.Time) string {
	return t.Format("20060102")
}

func secondsUntil(t time.Time, until string) int64 {
	hh, mm, ss := 0, 0, 0
	fmt.Sscanf(until, "%d:%d:%d", &hh, &mm, &ss)
	untilT := time.Date(t.Year(), t.Month(), t.Day(), hh, mm, ss, 0, t.Location())
	return untilT.Unix() - t.Unix()
}
