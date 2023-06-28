package controller

import (
	"context"
	"encoding/json"
	"log"
	"time"

	pUtils "github.com/bengosborn/cue/proximity/utils"
	"github.com/bengosborn/cue/utils"
)

const (
	radius   = 5
	syncTime = time.Minute * 5
)

// Routing logic for all broker messages
func Controller(ctx context.Context, location *pUtils.Location, brokerIn utils.Broker, brokerOut utils.Broker, lock *utils.ResourceLockDistributed, logger *log.Logger) {
	// Background sync
	go func() {
		for {
			timer := time.After(syncTime)

			select {
			case <-ctx.Done():
				return
			case <-timer:
				location.Sync()
			}
		}
	}()

	// Listen for new messages
	if err := brokerIn.Listen(func(msg *utils.BrokerMessage) bool {
		switch msg.EventType {
		case (utils.ProximitySendLocation):
			// Extract user data
			userData := &pUtils.UserData{}
			if err := json.Unmarshal([]byte(msg.Body), userData); err != nil {
				logger.Println("controller.error: ", err)
				return false
			}

			if err := location.Upsert(msg.User, userData.Lat, userData.Long); err != nil {
				logger.Println("controller.error: ", err)
				return false
			}

			logger.Println("controller.success: upserted user location data")

			return true

		case (utils.ProximityRequestNearby):
			// Request a list of users from the request
			out, err := location.Nearby(msg.User, radius)
			if err != nil {
				brokerOut.Send(&utils.BrokerMessage{Id: msg.Id, Receiver: msg.Receiver, User: msg.User, EventType: utils.Error, Body: err.Error()})
			}

			// Send the users to the user
			data, err := json.Marshal(out)
			if err != nil {
				brokerOut.Send(&utils.BrokerMessage{Id: msg.Id, Receiver: msg.Receiver, User: msg.User, EventType: utils.Error, Body: err.Error()})
			}

			brokerOut.Send(&utils.BrokerMessage{Id: msg.Id, Receiver: msg.Receiver, User: msg.User, EventType: utils.ProximityRequestNearby, Body: string(data)})

			return true

		default:
			return true
		}

	}, lock); err != nil {
		logger.Fatalln("controller.error:", err)
	}
}
