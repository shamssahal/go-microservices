package types

type OBUData struct {
	OBUID    int     `json:"obuID"`
	CurrLat  float64 `json:"currLat"`
	CurrLong float64 `json:"currLong"`
	PrevLat  float64 `json:"prevLat"`
	PrevLong float64 `json:"prevLong"`
}
