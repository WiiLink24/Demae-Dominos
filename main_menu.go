package main

import (
	"DemaeDominos/dominos"
	"encoding/xml"
	"fmt"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"io"
	"net/http"
	"strings"
)

func documentTemplate(r *Response) {
	r.AddKVWChildNode("container0", KVField{
		XMLName: xml.Name{Local: "contents"},
		Value:   "By clicking agree, you verify you have read and agree to https://demae.wiilink24.com/privacypolicy and https://demae.wiilink24.com/tos",
	})
	r.AddKVWChildNode("container1", KVField{
		XMLName: xml.Name{Local: "contents"},
		Value:   "Among Us",
	})
	r.AddKVWChildNode("container2", KVField{
		XMLName: xml.Name{Local: "contents"},
		Value:   "Among Us",
	})
}

func categoryList(r *Response) {
	dom, err := dominos.NewDominos(pool, r.request)
	if err != nil {
		r.ReportError(err, http.StatusUnauthorized)
		return
	}

	stores, err := dom.StoreLookup(r.request.Header.Get("X-Postalcode"), r.request.Header.Get("X-Address"))
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	storesXML := make([]BasicShop, 5)

	for i, storeData := range stores {
		store := BasicShop{
			ShopCode:    CDATA{storeData.StoreID},
			HomeCode:    CDATA{1},
			Name:        CDATA{"Domino's Pizza"},
			Catchphrase: CDATA{storeData.Address},
			// Min Price has been observed to be an average of $17 at most stores.
			MinPrice: CDATA{17},
			Yoyaku:   CDATA{1},
			Activate: CDATA{"on"},
			WaitTime: CDATA{storeData.WaitTime},
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

		storesXML[i] = store
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

	r.AddCustomType(shops)
	shops.XMLName = xml.Name{Local: "Placeholder"}
	r.AddCustomType(shops)
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

	r.ReportError(fmt.Errorf(errorString), http.StatusInternalServerError)
}
