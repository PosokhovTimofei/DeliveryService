package models

type CountryCoordinates struct {
	Name      string
	Country   string
	Latitude  float64
	Longitude float64
	Zone      string
}

type CountryDoc struct {
	Name     string `bson:"name"`
	Location struct {
		Type        string    `bson:"type"`
		Coordinates []float64 `bson:"coordinates"`
	} `bson:"location"`
	Zone string `bson:"zone"`
}
