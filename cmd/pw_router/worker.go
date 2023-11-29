package main

import (
	"context"
	"github.com/golang/protobuf/proto"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/maps"
	"math"
	"plane.watch/lib/export"
	"sync"
)

type (
	worker struct {
		router             *pwRouter
		destRoutingKeyLow  string
		destRoutingKeyHigh string
		spreadUpdates      bool

		ds *DataStream

		wireDecoder func([]byte) (*export.PlaneAndLocationInfoMsg, error)
		wireEncoder func(*export.PlaneAndLocationInfoMsg) ([]byte, error)
	}
)

const SigHeadingChange = 1.0        // at least 1.0 degrees change.
const SigVerticalRateChange = 180.0 // at least 180 fpm change (3ft in 1min)
const SigAltitudeChange = 10.0      // at least 10 ft in altitude change.

var poolUpdatesJSON sync.Pool
var poolUpdatesProtobuf sync.Pool

func init() {
	poolUpdatesJSON = sync.Pool{
		New: func() any {
			return export.NewEmptyPlaneLocationJSON()
		},
	}
	poolUpdatesProtobuf = sync.Pool{
		New: func() any {
			return export.NewPlaneAndLocationInfoMsg()
		},
	}
}

// workerInputJSON handles the incoming stream of json bytes and turns it into our msg type
func workerInputJSON(msg []byte) (*export.PlaneAndLocationInfoMsg, error) {
	var err error

	var json = jsoniter.ConfigFastest
	// unmarshal the JSON and ensure it's valid.
	// report the error if not and skip this message.
	update := poolUpdatesJSON.Get().(*export.PlaneLocationJSON)
	defer poolUpdatesJSON.Put(update)
	maps.Clear(update.SourceTags)
	if err = json.Unmarshal(msg, update); nil != err {
		log.Error().Err(err).Msg("Unable to unmarshal JSON")
		updatesError.Inc()
		return nil, err
	}

	if update.Icao == "" {
		log.Debug().Str("payload", string(msg)).Msg("empty ICAO")
		updatesError.Inc()
		return nil, nil
	}

	pliMsg := poolUpdatesProtobuf.Get().(*export.PlaneAndLocationInfoMsg)
	maps.Clear(pliMsg.SourceTags)
	err = update.ToProtobuf(pliMsg)
	return pliMsg, err
}

// workerOutputJSON handles an incoming protobuf encoded stream of bytes and turns it into our msg
func workerOutputJSON(msg *export.PlaneAndLocationInfoMsg) ([]byte, error) {
	return msg.ToJSONBytes()
}

// workerInputProtobuf handles an incoming protobuf encoded stream of bytes and turns it into our msg
func workerInputProtobuf(protobufBytes []byte) (*export.PlaneAndLocationInfoMsg, error) {
	msg := poolUpdatesProtobuf.Get().(*export.PlaneAndLocationInfoMsg)
	maps.Clear(msg.SourceTags)

	err := proto.Unmarshal(protobufBytes, msg.PlaneAndLocationInfo)
	if nil == msg.SourceTags {
		msg.SourceTags = make(map[string]uint32)
	}
	return msg, err
}

// workerOutputProtobuf handles an incoming protobuf encoded stream of bytes and turns it into our msg
func workerOutputProtobuf(msg *export.PlaneAndLocationInfoMsg) ([]byte, error) {
	return msg.ToProtobufBytes()
}

func (w *worker) run(ctx context.Context, ch <-chan []byte) {
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				log.Error().Msg("Worker ending due to error.")
				return
			}

			var gErr error
			if gErr = w.handleMsg(msg); nil != gErr {
				log.Error().Err(gErr).Send()
			}
		case <-ctx.Done():
			log.Debug().Msg("Ending Worker")
			return
		}
	}
}

func (w *worker) isSignificant(last, candidate *export.PlaneAndLocationInfoMsg) bool {
	// check the candidate vs last, if any of the following have changed
	// - Heading, VerticalRate, Velocity, Altitude, FlightNumber, FlightStatus, OnGround, Special, Squawk

	sigLog := log.With().
		Uint32("aircraft", candidate.Icao).
		Dur("diff_time", candidate.LastMsg.AsTime().Sub(last.LastMsg.AsTime())).
		Logger()

	if candidate.HasOnGround && candidate.OnGround {
		if log.Debug().Enabled() {
			sigLog.Debug().
				Msg("Aircraft is on ground")
		}
		return true
	}

	// if any of these fields differ, indicate this update is significant
	if candidate.HasHeading && last.HasHeading && math.Abs(candidate.Heading-last.Heading) > SigHeadingChange {
		if candidate.Updates.Heading.AsTime().After(last.Updates.Heading.AsTime()) {
			if log.Debug().Enabled() {
				sigLog.Debug().
					Float64("last", last.Heading).
					Float64("current", candidate.Heading).
					Float64("diff_value", last.Heading-candidate.Heading).
					Msg("Significant heading change.")
			}
			return true
		}
	}

	if candidate.HasVelocity && last.HasVelocity && candidate.Velocity != last.Velocity {
		if candidate.Updates.Velocity.AsTime().After(last.Updates.Velocity.AsTime()) {
			if log.Debug().Enabled() {
				sigLog.Debug().
					Float64("last", last.Velocity).
					Float64("current", candidate.Velocity).
					Float64("diff_value", last.Velocity-candidate.Velocity).
					Msg("Significant velocity change.")
			}
			return true
		}
	}

	if candidate.HasVerticalRate && last.HasVerticalRate && math.Abs(float64(candidate.VerticalRate-last.VerticalRate)) > SigVerticalRateChange {
		if candidate.Updates.VerticalRate.AsTime().After(last.Updates.VerticalRate.AsTime()) {
			if log.Debug().Enabled() {
				sigLog.Debug().
					Int32("last", last.VerticalRate).
					Int32("current", candidate.VerticalRate).
					Int32("diff_value", last.VerticalRate-candidate.VerticalRate).
					Msg("Significant vertical rate change.")
			}
		}
		return true
	}

	if math.Abs(float64(candidate.Altitude-last.Altitude)) > SigAltitudeChange {
		if candidate.Updates.Altitude.AsTime().After(last.Updates.Altitude.AsTime()) {
			if log.Debug().Enabled() {
				sigLog.Debug().
					Int32("last", last.Altitude).
					Int32("current", candidate.Altitude).
					Int32("diff_value", last.Altitude-candidate.Altitude).
					Msg("Significant altitude change.")
			}
			return true
		}
	}

	if candidate.FlightStatus != last.FlightStatus {
		if candidate.Updates.FlightStatus.AsTime().After(last.Updates.FlightStatus.AsTime()) {
			if log.Debug().Enabled() {
				sigLog.Debug().
					Str("last", export.FlightStatus_name[int32(last.FlightStatus)]).
					Str("current", export.FlightStatus_name[int32(candidate.FlightStatus)]).
					Msg("Significant FlightStatus change.")
			}
			return true
		}
	}

	if candidate.OnGround != last.OnGround {
		if candidate.Updates.OnGround.AsTime().After(last.Updates.OnGround.AsTime()) {
			if log.Debug().Enabled() {
				sigLog.Debug().
					Bool("last", last.OnGround).
					Bool("current", candidate.OnGround).
					Msg("Significant OnGround change.")
			}
			return true
		}
	}

	if candidate.Special != last.Special {
		if candidate.Updates.Special.AsTime().After(last.Updates.Special.AsTime()) {
			if log.Debug().Enabled() {
				sigLog.Debug().
					Str("last", last.Special).
					Str("current", candidate.Special).
					Msg("Significant Special change.")
			}
			return true
		}
	}

	if candidate.Squawk != last.Squawk {
		if candidate.Updates.Squawk.AsTime().After(last.Updates.Squawk.AsTime()) {
			if log.Debug().Enabled() {
				sigLog.Debug().
					Str("last", last.SquawkStr()).
					Str("current", candidate.SquawkStr()).
					Msg("Significant Squawk change.")
			}
			return true
		}
	}

	if candidate.TileLocation != last.TileLocation {
		if candidate.Updates.Location.AsTime().After(last.Updates.Location.AsTime()) {
			if log.Debug().Enabled() {
				sigLog.Debug().
					Str("last", last.TileLocation).
					Str("current", candidate.TileLocation).
					Msg("Significant TileLocation change.")
			}
			return true
		}
	}

	if log.Trace().Enabled() {
		sigLog.Trace().Msg("Ignoring insignificant event.")
	}

	return false
}

func (w *worker) handleMsg(msg []byte) error {
	update, err := w.wireDecoder(msg)
	if err != nil {
		log.Error().Err(err).Msg("failed to decode")
		updatesError.Inc()
		return nil
	}

	if update.Icao == 0 {
		log.Debug().Str("payload", string(msg)).Msg("empty ICAO")
		updatesError.Inc()
		return nil
	}

	updatesProcessed.Inc()

	// lookup what we know about this plane.
	item, ok := w.router.syncSamples.Load(update.Icao)

	// if this Icao is not in the cache, it's new.
	if !ok {
		if nil == update.SourceTags {
			update.SourceTags = make(map[string]uint32)
		}
		update.SourceTags[update.SourceTag]++
		w.router.syncSamples.Store(update.Icao, update)

		w.handleNewUpdate(update, msg)
		return nil // finish here, no significance check as we have nothing to compare.
	}

	// is this update significant versus the previous one
	lastRecord := item.(*export.PlaneAndLocationInfoMsg)
	merged, err := export.MergePlaneLocations(lastRecord, update)
	if nil != err {
		return nil
	}
	w.router.syncSamples.Store(merged.Icao, merged)

	mergedMsg, err := w.wireEncoder(merged)
	if err != nil {
		return err
	}

	if w.isSignificant(lastRecord, update) {
		w.handleSignificantUpdate(merged, mergedMsg)
	} else {
		w.handleInsignificantUpdate(merged, mergedMsg)
	}

	poolUpdatesProtobuf.Put(update)

	return nil
}

func (w *worker) handleRemovedUpdate(update export.PlaneLocationJSON, msg []byte) {
	// check if this is a removed record and purge it from the cache and emit an event
	// this ensures downstream pipeline components always know about a removed record.
	// we get the removed flag from pw_ingest - this shortcuts our cache expiry for efficiency.
	w.router.syncSamples.Delete(update.Icao)
	cacheEntries.Dec()
	cacheEvictions.Inc()

	// emit the event to both queues
	w.publishLocationUpdate(w.destRoutingKeyLow, msg)  // to the reduced feed queue
	w.publishLocationUpdate(w.destRoutingKeyHigh, msg) // to the full-feed queue

	if w.spreadUpdates {
		w.publishLocationUpdate(update.TileLocation+qSuffixLow, msg)  // to the low-speed tile-queue.
		w.publishLocationUpdate(update.TileLocation+qSuffixHigh, msg) // to the high-speed tile-queue.
	}
}

func (w *worker) handleSignificantUpdate(update *export.PlaneAndLocationInfoMsg, msg []byte) {
	// store the new update in-place of the old one
	// w.router.syncSamples.Store(update.Icao, update)
	updatesSignificant.Inc()

	// emit the new lastSignificant
	w.publishLocationUpdate(w.destRoutingKeyLow, msg)  // all low speed messages
	w.publishLocationUpdate(w.destRoutingKeyHigh, msg) // all high speed messages
	if w.spreadUpdates {
		w.publishLocationUpdate(update.TileLocation+qSuffixLow, msg)
		w.publishLocationUpdate(update.TileLocation+qSuffixHigh, msg)
	}
	if nil != w.ds {
		w.ds.AddLow(update)
	}
}

func (w *worker) handleNewUpdate(update *export.PlaneAndLocationInfoMsg, msg []byte) {
	// store the new update
	cacheEntries.Inc()

	if log.Debug().Enabled() {
		log.Debug().
			Str("aircraft", update.IcaoStr()).
			Msg("First time seeing aircraft.")
	}

	// new messages go to both queues
	w.publishLocationUpdate(w.destRoutingKeyLow, msg)  // all low speed messages
	w.publishLocationUpdate(w.destRoutingKeyHigh, msg) // all high speed messages

	// if spreading updates is enabled, output to spread queues
	if w.spreadUpdates {
		w.publishLocationUpdate(update.TileLocation+qSuffixLow, msg)
		w.publishLocationUpdate(update.TileLocation+qSuffixHigh, msg)
	}
}

func (w *worker) handleInsignificantUpdate(update *export.PlaneAndLocationInfoMsg, msg []byte) {
	updatesInsignificant.Inc()

	w.publishLocationUpdate(w.destRoutingKeyHigh, msg) // all high speed messages

	if w.spreadUpdates {
		// always publish updates to the high queue.
		w.publishLocationUpdate(update.TileLocation+qSuffixHigh, msg)
	}

	if nil != w.ds {
		w.ds.AddHigh(update)
	}
}

func (w *worker) publishLocationUpdate(routingKey string, msg []byte) {
	if log.Trace().Enabled() {
		log.Trace().Str("routing-key", routingKey).Bytes("Location", msg).Msg("Publish")
	}

	if err := w.router.nats.publish(routingKey, msg); nil != err {
		log.Warn().Err(err).Msg("Failed to send update")
		return
	}

	if log.Trace().Enabled() {
		log.Trace().Str("routingKey", routingKey).Msg("Sent msg")
	}
	updatesPublished.Inc()
}
