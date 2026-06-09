package orders

type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func (l *Location) SameLocation(loc Location) bool {
	return l.Lat == loc.Lat && l.Lng == loc.Lng
}
