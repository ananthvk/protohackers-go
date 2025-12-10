package internal

import (
	"sync"
)

type Plate string
type Road uint16

type PlateObservation struct {
	road      Road
	mile      uint16
	timestamp uint32
	plate     Plate
}

type Ticket struct {
	plate      Plate
	road       Road
	mile1      uint16
	mile2      uint16
	speed      uint16
	timestamp1 uint32
	timestamp2 uint32
}

type Store struct {
	mu           sync.Mutex
	observations map[Plate]map[Road][]PlateObservation
	pending      map[Road][]Ticket
	limits       map[Road]uint16 // Since a single road will have the same limit at every point on the road
	issued       map[Plate]map[uint32]struct{}
}

func NewStore() *Store {
	return &Store{
		observations: map[Plate]map[Road][]PlateObservation{},
		pending:      map[Road][]Ticket{},
		limits:       map[Road]uint16{},
		issued:       map[Plate]map[uint32]struct{}{},
	}
}
func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
func day(timestamp uint32) uint32 {
	return timestamp / 86400
}

// AddObservation adds the observation to the store, and if a ticket is generated, it's returned.
// If no ticket is generated, or a ticket has already been issued during the same duration, nil is returned
func (s *Store) AddObservation(observation PlateObservation) *Ticket {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if a camera has been registered for this road
	if _, ok := s.limits[observation.road]; !ok {
		return nil
	}

	// Check if the plate has been added before
	plateMap, ok := s.observations[observation.plate]
	if !ok {
		plateMap = make(map[Road][]PlateObservation)
		s.observations[observation.plate] = plateMap
	}

	// Check if adding the observation caused the vehicle to exceed the speed limit
	ticket := s.generateTicket(plateMap, observation)

	// Add the observation to the map
	plateMap[observation.road] = append(plateMap[observation.road], observation)

	return ticket
}

// IsIssued returns true if a ticket has been issued in any day between observation1 & observation2 (including both ends)
// Note: This function does not acquire a lock, and hence is unsafe for concurrent use
func (s *Store) isIssued(observation1, observation2 PlateObservation) bool {
	plate := observation1.plate
	issuedMap, ok := s.issued[plate]
	if !ok {
		return false
	}
	// Make observation 1 the smaller observation
	if observation1.timestamp > observation2.timestamp {
		observation1, observation2 = observation2, observation1
	}
	startDay := day(observation1.timestamp)
	endDay := day(observation2.timestamp)
	for i := startDay; i <= endDay; i++ {
		if _, ok := issuedMap[i]; ok {
			return true
		}
	}
	return false
}

// SetIsIssued marks all days between observation1 & observation2 as issued (including both ends)
// Note: This function does not acquire a lock, and hence is unsafe for concurrent use
func (s *Store) setIsIssued(observation1, observation2 PlateObservation) {
	plate := observation1.plate
	issuedMap, ok := s.issued[plate]
	if !ok {
		issuedMap = make(map[uint32]struct{})
		s.issued[plate] = issuedMap
	}
	if observation1.timestamp > observation2.timestamp {
		observation1, observation2 = observation2, observation1
	}
	startDay := day(observation1.timestamp)
	endDay := day(observation2.timestamp)
	for i := startDay; i <= endDay; i++ {
		issuedMap[i] = struct{}{}
	}
}

// generateTicket should be called when adding a new observation, but before it's added to the map. Note: This method is not safe for concurrent access, and
// hence the caller should hold the lock while calling this functin.
func (s *Store) generateTicket(plateMap map[Road][]PlateObservation, observation PlateObservation) *Ticket {
	limit := s.limits[observation.road]
	observations := plateMap[observation.road]
	for _, obs := range observations {
		deltaLength := abs(int64(obs.mile) - int64(observation.mile))
		// Get deltaTime in mph
		deltaTime := float64(abs(int64(obs.timestamp)-int64(observation.timestamp))) / (60.0 * 60.0)
		if deltaTime == 0 {
			continue
		}
		// Get the speed in miles per hour
		speed := float64(deltaLength) / float64(deltaTime)
		if speed >= float64(limit) {
			var mile1, mile2 uint16
			var timestamp1, timestamp2 uint32
			if obs.timestamp < observation.timestamp {
				mile1, timestamp1 = obs.mile, obs.timestamp
				mile2, timestamp2 = observation.mile, observation.timestamp
			} else {
				mile1, timestamp1 = observation.mile, observation.timestamp
				mile2, timestamp2 = obs.mile, obs.timestamp
			}
			if s.isIssued(observation, obs) {
				// Can't generate this ticket because it overlaps with another ticket
				continue
			}

			ticket := &Ticket{
				plate:      observation.plate,
				road:       observation.road,
				mile1:      mile1,
				mile2:      mile2,
				speed:      uint16(speed * 100),
				timestamp1: timestamp1,
				timestamp2: timestamp2,
			}
			s.setIsIssued(observation, obs)
			return ticket
		}
	}
	return nil
}

// AddPending stores the ticket in a pending list, that can be returned to a dispatcher when it comes online
func (s *Store) AddPending(ticket Ticket) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pending[ticket.road] = append(s.pending[ticket.road], ticket)
}

// GetPending returns the pending tickets that have yet to be sent for a particular road. After the pending tickets are
// returned (if any), all pending tickets for the road are deleted from the store.
func (s *Store) GetPending(road uint16) []Ticket {
	s.mu.Lock()
	defer s.mu.Unlock()
	pending := s.pending[Road(road)]
	s.pending[Road(road)] = nil
	return pending
}

func (s *Store) SetLimit(road Road, limit uint16) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.limits[road] = limit
}
