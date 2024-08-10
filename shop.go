package main

import (
	"DemaeDominos/dominos"
	"encoding/xml"
	"net/http"
)

func shopOne(r *Response) {
	var err error
	r.dominos, err = dominos.NewDominos(r.request)
	if err != nil {
		r.ReportError(err, http.StatusUnauthorized)
		return
	}

	shopData, err := r.dominos.GetStoreInfo(r.request.URL.Query().Get("shopCode"))
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	shop := ShopOne{
		CategoryCode:  CDATA{"01"},
		Address:       CDATA{"Nope"},
		Information:   CDATA{shopData.Information},
		Attention:     CDATA{"why"},
		Amenity:       CDATA{"Domino's Pizza"},
		MenuListCode:  CDATA{1},
		Activate:      CDATA{"on"},
		WaitTime:      CDATA{shopData.WaitTime},
		TimeOrder:     CDATA{"y"},
		Tel:           CDATA{shopData.Phone},
		YoyakuMinDate: CDATA{1},
		YoyakuMaxDate: CDATA{30},
		PaymentList: KVFieldWChildren{
			XMLName: xml.Name{Local: "paymentList"},
			Value: []any{
				KVField{
					XMLName: xml.Name{Local: "athing"},
					Value:   "Fox Card",
				},
			},
		},
		ShopStatus: ShopStatus{
			Hours: KVFieldWChildren{
				XMLName: xml.Name{Local: "hours"},
				Value: []any{
					KVFieldWChildren{
						XMLName: xml.Name{Local: "all"},
						Value: []any{
							KVField{
								XMLName: xml.Name{Local: "message"},
								Value:   shopData.DetailedWait,
							},
						},
					},
					KVFieldWChildren{
						XMLName: xml.Name{Local: "today"},
						Value: []any{
							KVFieldWChildren{
								XMLName: xml.Name{Local: "values"},
								Value: []any{
									KVField{
										XMLName: xml.Name{Local: "start"},
										Value:   "01:00:00",
									},
									KVField{
										XMLName: xml.Name{Local: "end"},
										Value:   "23:45:00",
									},
									KVField{
										XMLName: xml.Name{Local: "holiday"},
										Value:   "n",
									},
								},
							},
							KVFieldWChildren{
								XMLName: xml.Name{Local: "values1"},
								Value: []any{
									KVField{
										XMLName: xml.Name{Local: "start"},
										Value:   "01:00:00",
									},
									KVField{
										XMLName: xml.Name{Local: "end"},
										Value:   "23:45:00",
									},
									KVField{
										XMLName: xml.Name{Local: "holiday"},
										Value:   "n",
									},
								},
							},
						},
					},
					KVFieldWChildren{
						XMLName: xml.Name{Local: "delivery"},
						Value: []any{
							KVFieldWChildren{
								XMLName: xml.Name{Local: "values"},
								Value: []any{
									KVField{
										XMLName: xml.Name{Local: "start"},
										Value:   "01:00:00",
									},
									KVField{
										XMLName: xml.Name{Local: "end"},
										Value:   "23:45:00",
									},
									KVField{
										XMLName: xml.Name{Local: "holiday"},
										Value:   "n",
									},
								},
							},
						},
					},
					KVFieldWChildren{
						XMLName: xml.Name{Local: "selList"},
						Value:   []any{},
					},
					KVFieldWChildren{
						XMLName: xml.Name{Local: "status"},
						Value: []any{
							KVField{
								XMLName: xml.Name{Local: "isOpen"},
								Value:   BoolToInt(shopData.IsOpen),
							},
						},
					},
				},
			},
			Interval: CDATA{5},
			Holiday:  CDATA{"No ordering on Canada Day"},
		},
		RecommendedItemList: KVFieldWChildren{
			Value: []any{
				Item{
					XMLName:   xml.Name{Local: "container1"},
					MenuCode:  CDATA{10},
					ItemCode:  CDATA{1},
					Name:      CDATA{"Pizza"},
					Price:     CDATA{10},
					Info:      CDATA{"Fresh"},
					Size:      &CDATA{1},
					Image:     CDATA{"PIZZA"},
					IsSoldout: CDATA{0},
					SizeList: &KVFieldWChildren{
						XMLName: xml.Name{Local: "sizeList"},
						Value: []any{
							ItemSize{
								XMLName:   xml.Name{Local: "item1"},
								ItemCode:  CDATA{1},
								Size:      CDATA{1},
								Price:     CDATA{10},
								IsSoldout: CDATA{0},
							},
						},
					},
				},
			},
		},
	}

	// Strip the parent response tag
	r.ResponseFields = shop
}
