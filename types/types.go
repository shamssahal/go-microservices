package types

type OBUData struct {
	OBUID     int     `json:"obuID"`
	CurrLat   float64 `json:"currLat"`
	CurrLong  float64 `json:"currLong"`
	PrevLat   float64 `json:"prevLat"`
	PrevLong  float64 `json:"prevLong"`
	RequestID string  `json:"requestId"`
}

type Distance struct {
	Value     float64 `json:"value"`
	OBUID     int     `json:"obuID"`
	Unix      int64   `json:"unix"`
	RequestID string  `json:"requestId"`
}

type Invoice struct {
	OBUID         int     `json:"obuID"`
	TotalDistance float64 `json:"totalDistance"`
	TotalAmount   float64 `json:"totalAmount"`
}
