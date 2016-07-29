package triptime

import (
    "time"

    "github.com/padster/triptime/fb"

    ctx "golang.org/x/net/context"
    "google.golang.org/appengine/log"
    "google.golang.org/appengine/memcache"
)

type UserState struct {
    Position    fb.Coordinates
    StopAt      Stop
}

func GetUserState(c ctx.Context, userID string) *UserState {
    key := "ustate/" + userID
    var item UserState
    if _, err := memcache.JSON.Get(c, key, &item); err == memcache.ErrCacheMiss {
        return nil
    } else if err != nil {
        log.Errorf(c, "error getting state for user %s: %v", userID, err)
        return nil
    } else {
        return &item
    }
}

func NeedUserState(c ctx.Context, msg fb.Message) (*UserState, *fb.OutboundMessage) {
    state := GetUserState(c, msg.Sender.Id)
    if state == nil {
        text := "Hey, sorry, forgot where you are ☹ - can you send it again."
        return nil, &fb.OutboundMessage{
            msg.Sender,
            outMessageDataFromText(text),
        }
    } else {
        return state, nil
    }
} 

func SetUserState(c ctx.Context, userID string, state UserState) {
    key := "ustate/" + userID
    err := memcache.JSON.Set(c, &memcache.Item{
        Key:        key,
        Object:     state,
        Expiration: 600 * time.Second,
    })
    if err != nil {
        log.Errorf(c, "error writing state for user %s: %v", userID, err)
    }
}
