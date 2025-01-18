package main

import (
	"DemaeDominos/dominos"
	"encoding/xml"
	"fmt"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"io"
	"strings"
)

func documentTemplate(r *Response) {
	r.AddKVWChildNode("container0", KVField{
		XMLName: xml.Name{Local: "contents"},
		Value:   "By clicking agree, you verify you have read and\nagreed to https://wiilink.ca/demae/privacypolicy\nand https://wiilink.ca/tos.",
	})
	r.AddKVWChildNode("container1", KVField{
		XMLName: xml.Name{Local: "contents"},
		// Delivery success
		Value: "Enjoy your food!",
	})
	r.AddKVWChildNode("container2", KVField{
		XMLName: xml.Name{Local: "contents"},
		// Delivery failed
		Value: "Contact WiiLink Support with your Wii Number",
	})
}

func categoryList(r *Response) {
	var err error
	r.dominos, err = dominos.NewDominos(r.request)
	if err != nil {
		r.ReportError(err)
		return
	}

	postalCode := r.request.Header.Get("X-Postalcode")
	address := r.request.Header.Get("X-Address")
	aptNum := r.request.Header.Get("X-Aptnumber")
	stores, err := r.dominos.StoreLookup(postalCode, address, aptNum)
	if err != nil {
		// This endpoint will never return an error from Dominos, just a JSON decode error
		r.ReportError(err)
		return
	}

	var storesXML []BasicShop
	for _, storeData := range stores {
		// We need to get the actual min price
		shopData, err := r.dominos.GetStoreInfo(storeData.StoreID)
		if err != nil {
			r.ReportError(err)
			return
		}

		store := BasicShop{
			ShopCode:    CDATA{storeData.StoreID},
			HomeCode:    CDATA{1},
			Name:        CDATA{"Domino's Pizza"},
			Catchphrase: CDATA{shopData.Address},
			MinPrice:    CDATA{fmt.Sprintf("%.2f", shopData.MinPrice)},
			Yoyaku:      CDATA{1},
			Activate:    CDATA{"on"},
			WaitTime:    CDATA{storeData.WaitTime},
			PaymentList: KVFieldWChildren{
				XMLName: xml.Name{Local: "paymentList"},
				Value: []any{
					KVField{
						XMLName: xml.Name{Local: "athing"},
						Value:   "Fox Card",
					},
				},
			},
			ShopStatus: KVFieldWChildren{
				XMLName: xml.Name{Local: "shopStatus"},
				Value: []any{
					KVFieldWChildren{
						XMLName: xml.Name{Local: "status"},
						Value: []any{
							KVField{
								XMLName: xml.Name{Local: "isOpen"},
								Value:   BoolToInt(storeData.IsOpen),
							},
						},
					},
				},
			},
		}

		storesXML = append(storesXML, store)
	}

	shops := KVFieldWChildren{
		XMLName: xml.Name{Local: "Pizza"},
		Value: []any{
			KVField{
				XMLName: xml.Name{Local: "LargeCategoryName"},
				Value:   "Meal",
			},
			KVFieldWChildren{
				XMLName: xml.Name{Local: "CategoryList"},
				Value: []any{
					KVFieldWChildren{
						XMLName: xml.Name{Local: "TestingCategory"},
						Value: []any{
							KVField{
								XMLName: xml.Name{Local: "CategoryCode"},
								Value:   "01",
							},
							KVFieldWChildren{
								XMLName: xml.Name{Local: "ShopList"},
								Value: []any{
									storesXML,
								},
							},
						},
					},
				},
			},
		},
	}

	/*container := KVFieldWChildren{
		XMLName: xml.Name{Local: "container"},
		Value: []any{
			KVField{
				XMLName: xml.Name{Local: "CategoryCode"},
				Value:   "02",
			},
			KVFieldWChildren{
				XMLName: xml.Name{Local: "ShopList"},
				Value: []any{
					storesXML,
				},
			},
		},
	}*/

	placeholder := KVFieldWChildren{
		XMLName: xml.Name{Local: "Placeholder"},
		Value: []any{
			KVField{
				XMLName: xml.Name{Local: "LargeCategoryName"},
				Value:   "Meal",
			},
			KVFieldWChildren{
				XMLName: xml.Name{Local: "CategoryList"},
				Value: []any{
					KVFieldWChildren{
						XMLName: xml.Name{Local: "TestingCategory"},
						Value: []any{
							KVField{
								XMLName: xml.Name{Local: "CategoryCode"},
								Value:   "02",
							},
							KVFieldWChildren{
								XMLName: xml.Name{Local: "ShopList"},
								Value: []any{
									storesXML,
								},
							},
						},
					},
				},
			},
		},
	}

	r.AddCustomType(shops)

	// It there is no nearby stores, we do not add the placeholder. This will tell the user there are no stores.
	if storesXML != nil && r.request.URL.Query().Get("action") != "webApi_shop_list" {
		r.AddCustomType(placeholder)
	}
}

func shopInfo(r *Response) {
	var err error
	r.dominos, err = dominos.NewDominos(r.request)
	if err != nil {
		r.ReportError(err)
		return
	}

	postalCode := r.request.Header.Get("X-Postalcode")
	address := r.request.Header.Get("X-Address")
	aptNum := r.request.Header.Get("X-Aptnumber")
	stores, err := r.dominos.StoreLookup(postalCode, address, aptNum)
	if err != nil {
		// This endpoint will never return an error from Dominos, just a JSON decode error
		r.ReportError(err)
		return
	}

	var storesXML []BasicShop
	for _, storeData := range stores {
		// We need to get the actual min price
		shopData, err := r.dominos.GetStoreInfo(storeData.StoreID)
		if err != nil {
			r.ReportError(err)
			return
		}

		store := BasicShop{
			ShopCode:    CDATA{storeData.StoreID},
			HomeCode:    CDATA{1},
			Name:        CDATA{"Domino's Pizza"},
			Catchphrase: CDATA{storeData.Address},
			MinPrice:    CDATA{fmt.Sprintf("%.2f", shopData.MinPrice)},
			Yoyaku:      CDATA{1},
			Activate:    CDATA{"on"},
			WaitTime:    CDATA{storeData.WaitTime},
			PaymentList: KVFieldWChildren{
				XMLName: xml.Name{Local: "paymentList"},
				Value: []any{
					KVField{
						XMLName: xml.Name{Local: "athing"},
						Value:   "Fox Card",
					},
				},
			},
			ShopStatus: KVFieldWChildren{
				XMLName: xml.Name{Local: "shopStatus"},
				Value: []any{
					KVFieldWChildren{
						XMLName: xml.Name{Local: "status"},
						Value: []any{
							KVField{
								XMLName: xml.Name{Local: "isOpen"},
								Value:   BoolToInt(storeData.IsOpen),
							},
						},
					},
				},
			},
		}

		storesXML = append(storesXML, store)
	}
	container := KVFieldWChildren{
		XMLName: xml.Name{Local: "container"},
		Value: []any{
			KVField{
				XMLName: xml.Name{Local: "CategoryCode"},
				Value:   "01",
			},
			KVFieldWChildren{
				XMLName: xml.Name{Local: "ShopList"},
				Value: []any{
					storesXML,
				},
			},
		},
	}

	placeholder := KVFieldWChildren{
		XMLName: xml.Name{Local: "container"},
		Value: []any{
			KVField{
				XMLName: xml.Name{Local: "CategoryCode"},
				Value:   "02",
			},
			KVFieldWChildren{
				XMLName: xml.Name{Local: "ShopList"},
				Value: []any{
					storesXML,
				},
			},
		},
	}

	r.AddCustomType(container)

	// It there is no nearby stores, we do not add the placeholder. This will tell the user there are no stores.
	if storesXML != nil {
		r.AddCustomType(placeholder)
	}
}

func inquiryDone(r *Response) {
	// For our purposes, we will not be handling any restaurant requests.
	// However, the error endpoint uses this, so we must deal with that.
	// An error will never send a phone number, check for that first.
	if r.request.PostForm.Get("tel") != "" {
		return
	}

	shiftJisDecoder := func(str string) string {
		ret, _ := io.ReadAll(transform.NewReader(strings.NewReader(str), japanese.ShiftJIS.NewDecoder()))
		return string(ret)
	}

	// Now handle error.
	errorString := fmt.Sprintf(
		"An error has occured at on request %s\nError message: %s",
		shiftJisDecoder(r.request.PostForm.Get("requestType")),
		shiftJisDecoder(r.request.PostForm.Get("message")),
	)

	r.ReportError(fmt.Errorf(errorString))
}
