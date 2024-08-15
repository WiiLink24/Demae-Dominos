package main

import (
	"DemaeDominos/dominos"
	"encoding/xml"
	"net/http"
)

/*
Since Domino's in house API is really only used in their apps, some stuff makes sense to them to be hardcoded.
Such things are the topping selections in the pizza and pasta categories. We will be hard coding the names of
the selection sections, but pull available toppings from the API.
*/

func getToppings(r *http.Request) ([]any, error) {
	dom, err := dominos.NewDominos(r)
	if err != nil {
		return nil, err
	}

	toppingData, err := dom.GetToppings(r.URL.Query().Get("shopCode"), r.URL.Query().Get("itemCode"))
	if err != nil {
		return nil, err
	}

	sideData, err := dom.GetSides(r.URL.Query().Get("shopCode"), r.URL.Query().Get("itemCode"))
	if err != nil {
		return nil, err
	}

	var categories []ItemOne

	if toppingData != nil {
		categories = []ItemOne{
			{
				XMLName: xml.Name{Local: "container0"},
				Info:    CDATA{"Please choose a sauce."},
				Code:    CDATA{0},
				Type:    CDATA{"radio"},
				Name:    CDATA{"Sauce"},
				List: KVFieldWChildren{
					XMLName: xml.Name{Local: "list"},
				},
			},
			{
				XMLName: xml.Name{Local: "container1"},
				Info:    CDATA{"Please choose your meats. The * items are extra.\nNote you can only have 10 toppings."},
				Code:    CDATA{0},
				Type:    CDATA{1},
				Name:    CDATA{"Meats"},
				List: KVFieldWChildren{
					XMLName: xml.Name{Local: "list"},
				},
			},
			{
				XMLName: xml.Name{Local: "container2"},
				Info:    CDATA{"Please choose your non-meats. The * items are extra.\nNote you can only have 10 toppings."},
				Code:    CDATA{0},
				Type:    CDATA{1},
				Name:    CDATA{"Non-Meats"},
				List: KVFieldWChildren{
					XMLName: xml.Name{Local: "list"},
				},
			},
		}

		for _, topping := range toppingData {
			currentSelection := &(categories[topping.Group])
			currentSelection.List.Value = append(currentSelection.List.Value, Item{
				MenuCode:  CDATA{topping.Group},
				ItemCode:  CDATA{topping.Code},
				Name:      CDATA{topping.Name},
				Price:     CDATA{"--"},
				Info:      CDATA{""},
				Size:      nil,
				Image:     CDATA{"non"},
				IsSoldout: CDATA{BoolToInt(false)},
				SizeList:  nil,
			})
		}

		// Ensure toppings actually exist within the slice
		categoryLength := len(categories)
		for i := 0; i < categoryLength; i++ {
			if len(categories[i].List.Value) == 0 {
				categories = append(categories[:i], categories[i+1:]...)
				i -= 1
				categoryLength -= 1
			}
		}
	}

	if sideData != nil {
		sidesStruct := ItemOne{
			XMLName: xml.Name{Local: "container4"},
			Info:    CDATA{"Choose your sides. Note all sides are extra."},
			Code:    CDATA{0},
			Type:    CDATA{"1"},
			Name:    CDATA{"Sides"},
			List: KVFieldWChildren{
				XMLName: xml.Name{Local: "list"},
			},
		}

		for _, side := range sideData {
			sidesStruct.List.Value = append(sidesStruct.List.Value, Item{
				MenuCode:  CDATA{"side"},
				ItemCode:  CDATA{side.Code},
				Name:      CDATA{side.Name},
				Price:     CDATA{"--"},
				Info:      CDATA{""},
				Size:      nil,
				Image:     CDATA{"F_" + side.Code},
				IsSoldout: CDATA{BoolToInt(false)},
				SizeList:  nil,
			})
		}

		categories = append(categories, sidesStruct)
	}

	if sideData == nil && toppingData == nil {
		// This item has no toppings
		categories = append(categories, ItemOne{
			Info: CDATA{},
			Code: CDATA{},
			Type: CDATA{},
			Name: CDATA{},
			List: KVFieldWChildren{},
		})
	}

	return []any{categories[:]}, nil
}
