package domain

const (
	StatusAvailable = "available"
	StatusHeld      = "held"
	StatusBooked    = "booked"
)

type SeatMap struct {
	SessionID string   `json:"sessionId"`
	Event     EventRef `json:"event"`
	Hall      HallRef  `json:"hall"`
	Seats     []Seat   `json:"seats"`
}

type EventRef struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type HallRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Seat struct {
	SeatID string  `json:"seatId"`
	Row    string  `json:"row"`
	Number int     `json:"number"`
	Status string  `json:"status"`
	Price  float64 `json:"price"`
}
