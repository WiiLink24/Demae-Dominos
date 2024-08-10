package dominos

import (
	"net/http"
)

type Country string

const (
	USA    Country = "UNITED_STATES"
	CANADA Country = "CANADA"
)

var apiURLS = map[Country]string{
	USA:    "https://order.dominos.com",
	CANADA: "https://order.dominos.ca",
}
var imageURLS = map[Country]string{
	USA:    "https://cache.dominos.com/olo/6_92_1/assets/build/market/US/_en/images/img/products/larges",
	CANADA: "https://cache.dominos.com/nolo/ca/en/6_88_3/assets/build/market/CA/_en/images/img/products/larges",
}

type Dominos struct {
	country    Country
	apiURL     string
	imageURL   string
	response   []byte
	placeOrder bool
}

func (d *Dominos) SetResponse(data []byte) {
	d.response = data
}

func (d *Dominos) GetResponse() []byte {
	if d.response == nil {
		return []byte{0}
	}

	return d.response
}

func NewDominos(r *http.Request) (*Dominos, error) {
	d := Dominos{placeOrder: false}
	countryCode := r.Header.Get("X-WiiCountryCode")

	if countryCode == "18" {
		d.country = CANADA
	} else if countryCode == "49" {
		d.country = USA
	} else {
		return nil, InvalidCountry
	}

	d.apiURL = apiURLS[d.country]
	d.imageURL = imageURLS[d.country]

	return &d, nil
}
