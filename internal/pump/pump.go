package pump

import (
	"errors"
	"sync"

	"github.com/google/uuid"

	"drainnet/internal/audit"
	"drainnet/internal/level"
	"drainnet/internal/station"
	"drainnet/internal/store"
)

var ErrPumpNotFound = errors.New("pump not found")

var ErrPumpAlreadyRunning = errors.New("pump already running")

var ErrNoStandby = errors.New("no standby pump available")

type Controller struct {
	stations *station.Registry
	audits   *audit.Service
	levels   *level.Service
	mu       sync.Mutex
	lastRaw  map[string]float64
	speeds   map[string]string
	store    *store.Store
}

func NewController(st *store.Store, stations *station.Registry, audits *audit.Service, levels *level.Service) *Controller {
	return &Controller{
		store:    st,
		stations: stations,
		audits:   audits,
		levels:   levels,
		lastRaw:  map[string]float64{},
		speeds:   map[string]string{},
	}
}

func (c *Controller) Register(stationID string, ref station.PumpRef) (station.PumpRef, error) {
	if ref.ID == "" {
		ref.ID = uuid.NewString()
	}
	registered, err := c.stations.AddPump(stationID, ref)
	if err != nil {
		return station.PumpRef{}, err
	}
	if err := c.audits.Record(audit.Event{Type: "pump.register", EntityID: ref.ID, Message: "pump registered"}); err != nil {
		return station.PumpRef{}, err
	}
	return registered, nil
}

func (c *Controller) StateOf(stationID string, pumpID string) (station.PumpRef, error) {
	value, err := c.stations.Get(stationID)
	if err != nil {
		return station.PumpRef{}, err
	}
	for _, ref := range value.Pumps {
		if ref.ID == pumpID {
			return ref, nil
		}
	}
	return station.PumpRef{}, ErrPumpNotFound
}

func (c *Controller) List(stationID string) ([]station.PumpRef, error) {
	value, err := c.stations.Get(stationID)
	if err != nil {
		return nil, err
	}
	return value.Pumps, nil
}

func (c *Controller) Speed(stationID string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	speed, ok := c.speeds[stationID]
	if !ok {
		return "low", nil
	}
	return speed, nil
}
