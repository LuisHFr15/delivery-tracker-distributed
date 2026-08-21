package order

type Location struct {
	Lat float64 `json:"lat" dynamodbav:"Lat"`
	Lng float64 `json:"lng" dynamodbav:"Lng"`
}

func (l *Location) SameLocation(loc Location) bool {
	return l.Lat == loc.Lat && l.Lng == loc.Lng
}
