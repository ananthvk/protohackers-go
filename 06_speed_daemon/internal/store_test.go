package internal

import "testing"

func TestStore_AddObservation(t *testing.T) {
	tests := []struct {
		name         string
		setupLimits  map[Road]uint16
		observations []PlateObservation
		wantTickets  []*Ticket
	}{
		{
			name:        "no camera registered for road",
			setupLimits: map[Road]uint16{},
			observations: []PlateObservation{
				{road: 66, mile: 100, timestamp: 123456, plate: "UN1X"},
			},
			wantTickets: []*Ticket{nil},
		},
		{
			name:        "single observation no ticket",
			setupLimits: map[Road]uint16{66: 60},
			observations: []PlateObservation{
				{road: 66, mile: 100, timestamp: 123456, plate: "UN1X"},
			},
			wantTickets: []*Ticket{nil},
		},
		{
			name:        "speed limit not exceeded",
			setupLimits: map[Road]uint16{66: 60},
			observations: []PlateObservation{
				{road: 66, mile: 60, timestamp: 3600, plate: "UN1X"},
				{road: 66, mile: 45, timestamp: 0, plate: "UN1X"},
			},
			wantTickets: []*Ticket{nil, nil},
		},
		{
			name:        "speed limit exceeded",
			setupLimits: map[Road]uint16{66: 60},
			observations: []PlateObservation{
				{road: 66, mile: 60, timestamp: 3600, plate: "UN1X"},
				{road: 66, mile: 0, timestamp: 0, plate: "UN1X"},
			},
			wantTickets: []*Ticket{nil,
				{
					plate: "UN1X", road: 66, mile1: 0, mile2: 60,
					speed: 60, timestamp1: 0, timestamp2: 3600,
				},
			},
		},
		{
			name:        "different plates no interaction",
			setupLimits: map[Road]uint16{66: 60},
			observations: []PlateObservation{
				{road: 66, mile: 8, timestamp: 0, plate: "UN1X"},
				{road: 66, mile: 9, timestamp: 30, plate: "RE05BKG"},
			},
			wantTickets: []*Ticket{nil, nil},
		},
		{
			name:        "same timestamp no ticket",
			setupLimits: map[Road]uint16{66: 60},
			observations: []PlateObservation{
				{road: 66, mile: 8, timestamp: 100, plate: "UN1X"},
				{road: 66, mile: 9, timestamp: 100, plate: "UN1X"},
			},
			wantTickets: []*Ticket{nil, nil},
		},
		{
			name:        "multiple tickets on the same day should not be generated",
			setupLimits: map[Road]uint16{66: 60},
			observations: []PlateObservation{
				{road: 66, mile: 0, timestamp: 0, plate: "UN1X"},
				{road: 66, mile: 60, timestamp: 3600, plate: "UN1X"},
				{road: 66, mile: 120, timestamp: 7200, plate: "UN1X"},
				{road: 66, mile: 240, timestamp: 20, plate: "UN1X"},
				{road: 66, mile: 80, timestamp: 90000, plate: "UN1X"},
				{road: 66, mile: 200, timestamp: 93600, plate: "UN1X"},
				{road: 66, mile: 800, timestamp: 93700, plate: "UN1X"},
			},
			wantTickets: []*Ticket{nil,
				{
					plate: "UN1X", road: 66, mile1: 0, mile2: 60,
					speed: 60, timestamp1: 0, timestamp2: 3600,
				}, nil, nil, nil,
				{
					plate: "UN1X", road: 66, mile1: 80, mile2: 200,
					speed: 120, timestamp1: 90000, timestamp2: 93600,
				}, nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewStore()
			// Setup speed limits
			store.limits = tt.setupLimits

			for i, obs := range tt.observations {
				got := store.AddObservation(obs)
				want := tt.wantTickets[i]

				if want == nil && got != nil {
					t.Errorf("AddObservation() = %v, want nil", got)
				} else if want != nil && got == nil {
					t.Errorf("AddObservation() = nil, want %v", want)
				} else if want != nil && got != nil {
					if got.plate != want.plate || got.road != want.road ||
						got.mile1 != want.mile1 || got.mile2 != want.mile2 ||
						got.timestamp1 != want.timestamp1 || got.timestamp2 != want.timestamp2 {
						t.Errorf("AddObservation() = %v, want %v", got, want)
					}
				}
			}
		})
	}
}

func TestStore_AddGetPending(t *testing.T) {
	store := NewStore()
	ticket := Ticket{
		plate: "TEST123", road: 66, mile1: 8, mile2: 9,
		speed: 80, timestamp1: 0, timestamp2: 45,
	}
	store.AddPending(ticket)
	got := store.GetPending(66)
	if len(got) != 1 {
		t.Errorf("GetPending() = %v, want one element", got)
	}
	if got[0].plate != "TEST123" {
		t.Errorf("want %q, got %q", "TEST123", got[0].plate)
	}

	got = store.GetPending(66)
	if len(got) != 0 {
		t.Errorf("GetPending() = %v, want empty slice", got)
	}
}
