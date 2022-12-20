package main

import (
	"DemaeDominos/dominos"
	"encoding/xml"
	"fmt"
	"github.com/mitchellh/go-wordwrap"
	"net/http"
	"strings"
)

func menuList(r *Response) {
	dom, err := dominos.NewDominos(pool, r.request)
	if err != nil {
		r.ReportError(err, http.StatusUnauthorized)
		return
	}

	menuData, err := dom.GetMenu(r.request.URL.Query().Get("shopCode"))
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	var menus []Menu
	for _, menu := range menuData {
		menus = append(menus, Menu{
			XMLName:     xml.Name{Local: fmt.Sprintf("container_%s", menu.Code)},
			MenuCode:    CDATA{menu.Code},
			LinkTitle:   CDATA{menu.Name},
			EnabledLink: CDATA{0},
			Name:        CDATA{menu.Name},
			Info:        CDATA{menu.Name},
			SetNum:      CDATA{1},
			LunchMenuList: struct {
				IsLunchTimeMenu CDATA `xml:"isLunchTimeMenu"`
				Hour            KVFieldWChildren
				IsOpen          CDATA `xml:"isOpen"`
				Message         CDATA `xml:"message"`
			}{
				IsLunchTimeMenu: CDATA{BoolToInt(false)},
				Hour: KVFieldWChildren{
					XMLName: xml.Name{Local: "hour"},
					Value: []any{
						KVField{
							XMLName: xml.Name{Local: "start"},
							Value:   "00:00:00",
						},
						KVField{
							XMLName: xml.Name{Local: "end"},
							Value:   "24:59:59",
						},
					},
				},
				IsOpen:  CDATA{BoolToInt(true)},
				Message: CDATA{"Where does this show up?"},
			},
		})

		// If this food category does not have subcategories, make it one.
		if menu.Categories == nil {
			menus = append(menus, Menu{
				XMLName:     xml.Name{Local: fmt.Sprintf("container_%s", menu.Code)},
				MenuCode:    CDATA{menu.Code},
				LinkTitle:   CDATA{menu.Name},
				EnabledLink: CDATA{1},
				Name:        CDATA{menu.Name},
				Info:        CDATA{menu.Name},
				SetNum:      CDATA{1},
				LunchMenuList: struct {
					IsLunchTimeMenu CDATA `xml:"isLunchTimeMenu"`
					Hour            KVFieldWChildren
					IsOpen          CDATA `xml:"isOpen"`
					Message         CDATA `xml:"message"`
				}{
					IsLunchTimeMenu: CDATA{BoolToInt(false)},
					Hour: KVFieldWChildren{
						XMLName: xml.Name{Local: "hour"},
						Value: []any{
							KVField{
								XMLName: xml.Name{Local: "start"},
								Value:   "00:00:00",
							},
							KVField{
								XMLName: xml.Name{Local: "end"},
								Value:   "24:59:59",
							},
						},
					},
					IsOpen:  CDATA{BoolToInt(true)},
					Message: CDATA{"Where does this show up?"},
				},
			})
		} else {
			for _, subcategory := range menu.Categories {
				menus = append(menus, Menu{
					XMLName:     xml.Name{Local: fmt.Sprintf("container_%s", subcategory.Code)},
					MenuCode:    CDATA{subcategory.Code},
					LinkTitle:   CDATA{subcategory.Name},
					EnabledLink: CDATA{1},
					Name:        CDATA{subcategory.Name},
					Info:        CDATA{subcategory.Name},
					SetNum:      CDATA{1},
					LunchMenuList: struct {
						IsLunchTimeMenu CDATA `xml:"isLunchTimeMenu"`
						Hour            KVFieldWChildren
						IsOpen          CDATA `xml:"isOpen"`
						Message         CDATA `xml:"message"`
					}{
						IsLunchTimeMenu: CDATA{BoolToInt(false)},
						Hour: KVFieldWChildren{
							XMLName: xml.Name{Local: "hour"},
							Value: []any{
								KVField{
									XMLName: xml.Name{Local: "start"},
									Value:   "00:00:00",
								},
								KVField{
									XMLName: xml.Name{Local: "end"},
									Value:   "24:59:59",
								},
							},
						},
						IsOpen:  CDATA{BoolToInt(true)},
						Message: CDATA{"Where does this show up?"},
					},
				})
			}
		}
	}

	// Append 1 more as a placeholder
	placeholder := menus[0]
	placeholder.XMLName = xml.Name{Local: "placeholder"}
	menus = append(menus, placeholder)
	r.AddCustomType(menus)
}

func itemList(r *Response) {
	var items []NestedItem
	dom, err := dominos.NewDominos(pool, r.request)
	if err != nil {
		r.ReportError(err, http.StatusUnauthorized)
		return
	}

	itemData, err := dom.GetItemList(r.request.URL.Query().Get("shopCode"), r.request.URL.Query().Get("menuCode"))
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	for i, item := range itemData {
		name := wordwrap.WrapString(item.Name, 26)
		for i, s := range strings.Split(name, "\n") {
			switch i {
			case 0:
				name = s
				break
			default:
				name += "\n"
				name += s
				break
			}
		}

		description := wordwrap.WrapString(item.Description, 39)
		for i, s := range strings.Split(description, "\n") {
			switch i {
			case 0:
				description = s
				break
			case 1:
				description += "\n"
				description += s
				break
			case 2:
				description += "\n"
				description += strings.Split(wordwrap.WrapString(s, 22), "\n")[0]
				break
			default:
				break
			}
		}

		nestedItem := NestedItem{
			XMLName: xml.Name{Local: fmt.Sprintf("container%d", i)},
			Name:    CDATA{name},
			Item: Item{
				XMLName:   xml.Name{Local: "item"},
				MenuCode:  CDATA{r.request.URL.Query().Get("menuCode")},
				ItemCode:  CDATA{item.Img},
				Price:     CDATA{"vee"},
				Info:      CDATA{description},
				Size:      &CDATA{"something"},
				Image:     CDATA{item.Img},
				IsSoldout: CDATA{BoolToInt(false)},
				SizeList: &KVFieldWChildren{
					XMLName: xml.Name{Local: "sizeList"},
					Value:   nil,
				},
			},
		}

		for i2, size := range item.Items {
			sizeName := wordwrap.WrapString(size.Name, 21)
			for i, s := range strings.Split(sizeName, "\n") {
				switch i {
				case 0:
					sizeName = s
					break
				case 1:
					sizeName += "\n"
					sizeName += s
					break
				default:
					// If it is too long it becomes ... so we are fine
					sizeName += " " + s
					break
				}
			}

			nestedItem.Item.SizeList.Value = append(nestedItem.Item.SizeList.Value, ItemSize{
				XMLName:   xml.Name{Local: fmt.Sprintf("item%d", i2)},
				ItemCode:  CDATA{size.Code},
				Size:      CDATA{sizeName},
				Price:     CDATA{size.Price},
				IsSoldout: CDATA{BoolToInt(false)},
			})
		}

		items = append(items, nestedItem)
	}

	r.ResponseFields = []any{
		KVField{
			XMLName: xml.Name{Local: "Count"},
			Value:   len(items),
		},
		KVFieldWChildren{
			XMLName: xml.Name{Local: "List"},
			Value:   []any{items[:]},
		},
	}
}

func itemOne(r *Response) {
	dom, err := dominos.NewDominos(pool, r.request)
	options, err := getToppings(r.request)
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	price, err := dom.GetFoodPrice(r.request.URL.Query().Get("shopCode"), r.request.URL.Query().Get("itemCode"))
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	r.ResponseFields = []any{
		KVField{
			XMLName: xml.Name{Local: "price"},
			Value:   price,
		},
		KVFieldWChildren{
			XMLName: xml.Name{Local: "optionList"},
			Value:   options,
		},
	}
}
