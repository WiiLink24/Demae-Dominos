package dominos

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"net/http"
	"strings"
)

type Country string

const (
	USA               Country = "UNITED_STATES"
	CANADA            Country = "CANADA"
	QueryUserAreaCode         = `SELECT "user".area_code FROM "user" WHERE "user".wii_id = $1 LIMIT 1`
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
	country      Country
	apiURL       string
	imageURL     string
	jsonResponse *map[string]any
}

func NewDominos(db *pgxpool.Pool, r *http.Request) (*Dominos, error) {
	d := Dominos{}
	areaCode := r.URL.Query().Get("areaCode")

	if areaCode == "" {
		row := db.QueryRow(context.Background(), QueryUserAreaCode, r.Header.Get("X-WiiID"))
		err := row.Scan(&areaCode)
		if err != nil {
			return nil, err
		}
	}

	if strings.Split(areaCode, "")[0] == "0" {
		d.country = USA
	} else {
		d.country = CANADA
	}

	d.apiURL = apiURLS[d.country]
	d.imageURL = imageURLS[d.country]

	return &d, nil
}
