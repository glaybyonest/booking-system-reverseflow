package domain

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

const ProviderReactSeatToolkit = "react_seat_toolkit"

type StoredSeatLayout struct {
	Version int                 `json:"version"`
	Canvas  StoredLayoutCanvas  `json:"canvas"`
	Stage   *StoredLayoutStage  `json:"stage,omitempty"`
	Seats   []StoredLayoutSeat  `json:"seats"`
	Meta    map[string]any      `json:"meta,omitempty"`
}

type StoredLayoutCanvas struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type StoredLayoutStage struct {
	Label  string `json:"label,omitempty"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type StoredLayoutSeat struct {
	Key      string  `json:"key"`
	Label    string  `json:"label,omitempty"`
	Row      string  `json:"row"`
	Number   int     `json:"number"`
	X        int     `json:"x"`
	Y        int     `json:"y"`
	Price    float64 `json:"price"`
	Category string  `json:"category,omitempty"`
}

func DecodeStoredSeatLayout(raw []byte) (*StoredSeatLayout, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	var layout StoredSeatLayout
	if err := json.Unmarshal(raw, &layout); err != nil {
		return nil, fmt.Errorf("decode seat layout: %w", err)
	}
	if err := layout.Validate(); err != nil {
		return nil, err
	}
	return &layout, nil
}

func (layout StoredSeatLayout) Validate() error {
	if layout.Version <= 0 {
		return fmt.Errorf("layout.version must be greater than zero")
	}
	if layout.Canvas.Width <= 0 || layout.Canvas.Height <= 0 {
		return fmt.Errorf("layout.canvas width and height must be greater than zero")
	}
	if len(layout.Seats) == 0 {
		return fmt.Errorf("layout must contain at least one seat")
	}

	keys := make(map[string]struct{}, len(layout.Seats))
	labels := make(map[string]struct{}, len(layout.Seats))
	positions := make(map[string]struct{}, len(layout.Seats))
	for _, seat := range layout.Seats {
		key := strings.TrimSpace(seat.Key)
		row := strings.TrimSpace(seat.Row)
		label := strings.TrimSpace(seat.Label)

		if key == "" {
			return fmt.Errorf("layout.seats[].key is required")
		}
		if row == "" {
			return fmt.Errorf("layout.seats[].row is required")
		}
		if seat.Number <= 0 {
			return fmt.Errorf("layout.seats[].number must be greater than zero")
		}
		if seat.Price < 0 {
			return fmt.Errorf("layout.seats[].price must be zero or greater")
		}
		if _, exists := keys[key]; exists {
			return fmt.Errorf("duplicate layout seat key: %s", key)
		}
		keys[key] = struct{}{}
		positionKey := fmt.Sprintf("%s-%d", row, seat.Number)
		if _, exists := positions[positionKey]; exists {
			return fmt.Errorf("duplicate layout seat position: %s", positionKey)
		}
		positions[positionKey] = struct{}{}
		if label != "" {
			if _, exists := labels[label]; exists {
				return fmt.Errorf("duplicate layout seat label: %s", label)
			}
			labels[label] = struct{}{}
		}
	}
	return nil
}

func (layout StoredSeatLayout) SortedSeats() []StoredLayoutSeat {
	items := append([]StoredLayoutSeat(nil), layout.Seats...)
	sort.Slice(items, func(i, j int) bool {
		if items[i].Row == items[j].Row {
			if items[i].Number == items[j].Number {
				return items[i].Key < items[j].Key
			}
			return items[i].Number < items[j].Number
		}
		return items[i].Row < items[j].Row
	})
	return items
}

func (layout StoredSeatLayout) SeatLabel(seat StoredLayoutSeat) string {
	if label := strings.TrimSpace(seat.Label); label != "" {
		return label
	}
	return fmt.Sprintf("%s-%d", strings.TrimSpace(seat.Row), seat.Number)
}

func (layout StoredSeatLayout) Dimensions() (rowsCount int, seatsPerRow int) {
	rowCounts := make(map[string]int)
	for _, seat := range layout.Seats {
		rowCounts[strings.TrimSpace(seat.Row)]++
	}
	rowsCount = len(rowCounts)
	for _, count := range rowCounts {
		if count > seatsPerRow {
			seatsPerRow = count
		}
	}
	return rowsCount, seatsPerRow
}
