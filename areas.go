package main

import (
	"context"
	"encoding/xml"
	"math/rand"
	"strconv"
)

const InsertUser = `UPDATE "user" SET area_code = $1 WHERE wii_id = $2`

var canadianProvinces = map[string]string{
	"01": "Alberta",
	"02": "British Colombia",
	"03": "Manitoba",
	"04": "New Brunswick",
	"05": "Newfoundland and Labrador",
	"07": "Nova Scotia",
	"08": "Ontario",
	"09": "Prince Edward Island",
	"10": "Quebec",
	"11": "Saskatchewan",
	"12": "Yukon",
	"13": "Northwest Territories",
}

var americanStates = map[string]string{
	"AK": "Alaska",
	"AL": "Alabama",
	"AR": "Arkansas",
	"AZ": "Arizona",
	"CA": "California",
	"CO": "Colorado",
	"CT": "Connecticut",
	"DC": "District of Columbia",
	"DE": "Delaware",
	"FL": "Florida",
	"GA": "Georgia",
	"HI": "Hawaii",
	"IA": "Iowa",
	"ID": "Idaho",
	"IL": "Illinois",
	"IN": "Indiana",
	"KS": "Kansas",
	"KY": "Kentucky",
	"LA": "Louisiana",
	"MA": "Massachusetts",
	"MD": "Maryland",
	"ME": "Maine",
	"MI": "Michigan",
	"MN": "Minnesota",
	"MO": "Missouri",
	"MS": "Mississippi",
	"MT": "Montana",
	"NC": "North Carolina",
	"ND": "North Dakota",
	"NE": "Nebraska",
	"NH": "New Hampshire",
	"NJ": "New Jersey",
	"NM": "New Mexico",
	"NV": "Nevada",
	"NY": "New York",
	"OH": "Ohio",
	"OK": "Oklahoma",
	"OR": "Oregon",
	"PA": "Pennsylvania",
	"RI": "Rhode Island",
	"SC": "South Carolina",
	"SD": "South Dakota",
	"TN": "Tennessee",
	"TX": "Texas",
	"UT": "Utah",
	"VA": "Virginia",
	"VT": "Vermont",
	"WA": "Washington",
	"WI": "Wisconsin",
	"WV": "West Virginia",
	"WY": "Wyoming",
}

func GetAmericanStates() []AreaNames {
	var areaNames []AreaNames

	for s, s2 := range americanStates {
		areaNames = append(areaNames, AreaNames{
			AreaName: CDATA{s2},
			AreaCode: CDATA{s},
		})
	}

	return areaNames
}

func GetCanadianProvinces() []AreaNames {
	var areaNames []AreaNames

	for s, s2 := range canadianProvinces {
		areaNames = append(areaNames, AreaNames{
			AreaName: CDATA{s2},
			AreaCode: CDATA{s},
		})
	}

	return areaNames
}

func GetStateName(stateCode string) string {
	if IsAreaAmerican(stateCode) {
		return americanStates[stateCode]
	} else {
		return canadianProvinces[stateCode]
	}
}

func IsAreaAmerican(stateCode string) bool {
	if _, ok := canadianProvinces[stateCode]; ok {
		return false
	} else {
		return true
	}
}

func GetCitiesByStateCode(stateCode, areaCode string) []Area {
	var cities []Area

	for _, city := range geonameCities {
		if city.CountryCode == "CA" || city.CountryCode == "US" {
			if (city.Admin1Code == stateCode && *city.Population > 50000) || (city.Admin1Code == "WV" && *city.Population > 30000) {
				cities = append(cities, Area{
					AreaName:   CDATA{city.Name},
					AreaCode:   CDATA{areaCode},
					IsNextArea: CDATA{0},
					Display:    CDATA{1},
					Kanji1:     CDATA{GetStateName(stateCode)},
					Kanji2:     CDATA{city.Name},
					Kanji3:     CDATA{""},
					Kanji4:     CDATA{""},
				})
			}
		}
	}

	return cities
}

const numberBytes = "0123456789"

func GenerateAreaCode(areaCode string) string {
	b := make([]byte, 10)
	for i := range b {
		b[i] = numberBytes[rand.Intn(len(numberBytes))]
	}

	if IsAreaAmerican(areaCode) {
		return "0" + string(b)
	} else {
		return "1" + string(b)
	}
}

func areaList(r *Response) {
	areaCode := r.request.URL.Query().Get("areaCode")

	// Nintendo, for whatever reason, require a separate "selectedArea" element
	// as a root node within output.
	// This violates about every XML specification in existence.
	// I am reasonably certain there was a mistake as their function to
	// interpret nodes at levels accepts a parent node, to which they seem to
	// have passed NULL instead of response.
	//
	// We are not going to bother spending time to deal with this.
	if r.request.URL.Query().Get("zipCode") != "" {
		version, apiStatus := GenerateVersionAndAPIStatus()
		r.ResponseFields = []any{
			KVFieldWChildren{
				XMLName: xml.Name{Local: "response"},
				Value: []any{
					KVFieldWChildren{
						XMLName: xml.Name{Local: "areaList"},
						Value: []any{
							KVField{
								XMLName: xml.Name{Local: "segment"},
								Value:   "United States",
							},
							KVFieldWChildren{
								XMLName: xml.Name{Local: "list"},
								Value: []any{
									KVFieldWChildren{
										XMLName: xml.Name{Local: "areaPlace"},
										Value: []any{AreaNames{
											AreaName: CDATA{"place name"},
											AreaCode: CDATA{2},
										}},
									},
								},
							},
						},
					},
					KVField{
						XMLName: xml.Name{Local: "areaCount"},
						Value:   "1",
					},
					version,
					apiStatus,
				},
			},
			KVFieldWChildren{
				XMLName: xml.Name{Local: "selectedArea"},
				Value: []any{
					KVField{
						XMLName: xml.Name{Local: "areaCode"},
						Value:   1,
					},
				},
			},
		}
		return
	}

	if areaCode == "0" {
		r.AddKVWChildNode("areaList", []any{
			KVFieldWChildren{
				XMLName: xml.Name{Local: "place"},
				Value: []any{
					KVField{
						XMLName: xml.Name{Local: "segment"},
						Value:   "United States",
					},
					KVFieldWChildren{
						XMLName: xml.Name{Local: "list"},
						Value: []any{
							GetAmericanStates()[:],
						},
					},
				},
			},
			KVFieldWChildren{
				XMLName: xml.Name{Local: "place1"},
				Value: []any{
					KVField{
						XMLName: xml.Name{Local: "segment"},
						Value:   "Canada",
					},
					KVFieldWChildren{
						XMLName: xml.Name{Local: "list"},
						Value: []any{
							GetCanadianProvinces()[:],
						},
					},
				},
			},
		})
		r.AddKVNode("areaCount", "2")
		return
	}

	newAreaCode := GenerateAreaCode(areaCode)
	_, err := pool.Exec(context.Background(), InsertUser, newAreaCode, r.GetHollywoodId())
	if err != nil {
		r.ReportError(err)
		return
	}

	cities := GetCitiesByStateCode(areaCode, newAreaCode)
	r.AddKVWChildNode("areaList", KVFieldWChildren{
		XMLName: xml.Name{Local: "place"},
		Value: []any{
			KVField{
				XMLName: xml.Name{Local: "container0"},
				Value:   "aaaaa",
			},
			KVField{
				XMLName: xml.Name{Local: "segment"},
				Value:   GetStateName(areaCode),
			},
			KVFieldWChildren{
				XMLName: xml.Name{Local: "list"},
				Value: []any{
					cities[:],
				},
			},
		},
	})
	r.AddKVNode("areaCount", strconv.FormatInt(int64(len(cities)), 10))
}
