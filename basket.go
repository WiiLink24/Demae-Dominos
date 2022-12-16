package main

import (
	"DemaeDominos/dominos"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/gofrs/uuid"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	QueryUserBasket   = `SELECT "user".basket FROM "user" WHERE "user".wii_id = $1 LIMIT 1`
	QueryUserForOrder = `SELECT "user".basket, "user".price, "user".order_id FROM "user" WHERE "user".wii_id = $1 LIMIT 1`
)

func authKey(r *Response) {
	var areaCode string
	row := pool.QueryRow(context.Background(), dominos.QueryUserAreaCode, r.request.Header.Get("X-WiiID"))
	err := row.Scan(&areaCode)
	if err != nil {
		r.ReportError(err, http.StatusUnauthorized)
		return
	}

	// TODO: Append authKey to database
	authKeyValue, err := uuid.DefaultGenerator.NewV1()
	if err != nil {
		r.ReportError(err, http.StatusUnauthorized)
		return
	}

	r.ResponseFields = []any{
		KVField{
			XMLName: xml.Name{Local: "authKey"},
			Value:   authKeyValue.String(),
		},
	}
}

func basketReset(r *Response) {
	_, err := pool.Exec(context.Background(), `UPDATE "user" SET order_id = $1, price = $2, basket = $3 WHERE wii_id = $4`, "", "", "[]", r.request.Header.Get("X-WiiID"))
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}
}

func basketDelete(r *Response) {
	basketNumber, err := strconv.ParseInt(r.request.URL.Query().Get("basketNo"), 10, 64)
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	var lastBasket string
	row := pool.QueryRow(context.Background(), QueryUserBasket, r.request.Header.Get("X-WiiID"))
	err = row.Scan(&lastBasket)
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	var actualBasket []map[string]any
	err = json.Unmarshal([]byte(lastBasket), &actualBasket)
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	actualBasket = append(actualBasket[:basketNumber-1], actualBasket[basketNumber:]...)

	// Convert basket to JSON then insert to database
	jsonStr, err := json.Marshal(actualBasket)
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	_, err = pool.Exec(context.Background(), `UPDATE "user" SET basket = $1 WHERE wii_id = $2`, jsonStr, r.request.Header.Get("X-WiiID"))
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}
}

func basketAdd(r *Response) {
	areaCode := r.request.PostForm.Get("areaCode")
	shopCode := r.request.PostForm.Get("shopCode")
	itemCode := r.request.PostForm.Get("itemCode")
	quantity := r.request.PostForm.Get("quantity")

	var lastBasket string
	dom, err := dominos.NewDominos(pool, r.request)
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	// First we must parse the sides and toppings
	var toppings []dominos.Topping
	for items, _ := range r.request.PostForm {
		if strings.Contains(items, "option") {
			// Extract the topping type and code
			var toppingType dominos.ToppingGroup
			var code string
			for i, s := range strings.Split(items, "[") {
				switch i {
				case 0:
					continue
				case 1:
					toppingType.New(strings.Split(s, "]")[0])
				case 2:
					code = strings.Split(s, "]")[0]
				}
			}

			toppings = append(toppings, dominos.Topping{
				Code:  code,
				Name:  "",
				Group: toppingType,
			})
		}
	}

	// Create our basket
	basket, err := dom.AddItem(shopCode, itemCode, quantity, toppings)
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	row := pool.QueryRow(context.Background(), QueryUserBasket, r.request.Header.Get("X-WiiID"))
	err = row.Scan(&lastBasket)
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	var actualBasket []map[string]any
	err = json.Unmarshal([]byte(lastBasket), &actualBasket)
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	actualBasket = append(actualBasket, basket)

	// Convert basket to JSON then insert to database
	jsonStr, err := json.Marshal(actualBasket)
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	_, err = pool.Exec(context.Background(), `UPDATE "user" SET basket = $1 WHERE area_code = $2`, jsonStr, areaCode)
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}
}

func basketList(r *Response) {
	areaCode := r.request.URL.Query().Get("areaCode")
	address := r.request.Header.Get("X-Address")
	postalCode := r.request.Header.Get("X-Postalcode")

	var basketStr string
	row := pool.QueryRow(context.Background(), QueryUserBasket, r.request.Header.Get("X-WiiID"))
	err := row.Scan(&basketStr)
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	var basket []map[string]any
	err = json.Unmarshal([]byte(basketStr), &basket)
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	dom, err := dominos.NewDominos(pool, r.request)
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	user, err := dom.AddressLookup(postalCode, address)
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	user.StoreId = r.request.URL.Query().Get("shopCode")
	user.Products = basket

	items, err := dom.GetPrice(user)
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	// Add order ID and price to database
	_, err = pool.Exec(context.Background(), `UPDATE "user" SET order_id = $1, price = $2 WHERE area_code = $3`, items.OrderId, fmt.Sprintf("%.2f", items.TotalPrice), areaCode)
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	basketPrice := KVField{
		XMLName: xml.Name{Local: "basketPrice"},
		Value:   items.BasketPrice,
	}

	chargePrice := KVField{
		XMLName: xml.Name{Local: "chargePrice"},
		Value:   items.ChargePrice,
	}

	totalPrice := KVField{
		XMLName: xml.Name{Local: "totalPrice"},
		Value:   items.TotalPrice,
	}

	status := KVFieldWChildren{
		XMLName: xml.Name{Local: "Status"},
		Value: []any{
			KVField{
				XMLName: xml.Name{Local: "isOrder"},
				Value:   BoolToInt(true),
			},
			KVFieldWChildren{
				XMLName: xml.Name{Local: "messages"},
				Value: []any{KVField{
					XMLName: xml.Name{Local: "hey"},
					Value:   "how are you?",
				}},
			},
		},
	}

	var basketItems []BasketItem
	for i, item := range items.Items {
		options := ItemOne{
			XMLName: xml.Name{Local: "container0"},
			Info:    CDATA{""},
			Code:    CDATA{0},
			Type:    CDATA{0},
			Name:    CDATA{"Add-ons"},
			List:    KVFieldWChildren{},
		}
		for _, option := range item.Options {
			options.List.Value = append(options.List.Value, Item{
				MenuCode:   CDATA{0},
				ItemCode:   CDATA{0},
				Name:       CDATA{option},
				Price:      CDATA{0},
				Info:       CDATA{0},
				IsSelected: &CDATA{BoolToInt(true)},
				Image:      CDATA{0},
				IsSoldout:  CDATA{BoolToInt(false)},
			})
		}

		basketItems = append(basketItems, BasketItem{
			XMLName:       xml.Name{Local: fmt.Sprintf("container%d", i)},
			BasketNo:      CDATA{i + 1},
			MenuCode:      CDATA{1},
			ItemCode:      CDATA{item.Code},
			Name:          CDATA{item.Name},
			Price:         CDATA{item.Price},
			Size:          CDATA{""},
			IsSoldout:     CDATA{BoolToInt(false)},
			Quantity:      CDATA{item.Quantity},
			SubTotalPrice: CDATA{item.Amount},
			Menu: KVFieldWChildren{
				XMLName: xml.Name{Local: "Menu"},
				Value: []any{
					KVField{
						XMLName: xml.Name{Local: "name"},
						Value:   "Menu",
					},
					KVFieldWChildren{
						XMLName: xml.Name{Local: "lunchMenuList"},
						Value: []any{
							KVField{
								XMLName: xml.Name{Local: "isLunchTimeMenu"},
								Value:   BoolToInt(false),
							},
							KVField{
								XMLName: xml.Name{Local: "isOpen"},
								Value:   BoolToInt(true),
							},
						},
					},
				},
			},
			OptionList: KVFieldWChildren{
				XMLName: xml.Name{Local: ""},
				Value: []any{
					options,
				},
			},
		})
	}

	cart := KVFieldWChildren{
		XMLName: xml.Name{Local: "List"},
		Value:   []any{basketItems[:]},
	}

	r.ResponseFields = []any{
		basketPrice,
		chargePrice,
		totalPrice,
		status,
		cart,
	}
}

func orderDone(r *Response) {
	var basketStr string
	var price string
	var orderId string
	row := pool.QueryRow(context.Background(), QueryUserForOrder, r.request.Header.Get("X-WiiID"))
	err := row.Scan(&basketStr, &price, &orderId)
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	var basket []map[string]any
	err = json.Unmarshal([]byte(basketStr), &basket)
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	dom, err := dominos.NewDominos(pool, r.request)
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	user, err := dom.AddressLookup(r.request.PostForm.Get("member[PostNo]"), r.request.PostForm.Get("member[Address5]"))
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	fmt.Println(user.StreetNumber)
	fmt.Println(user.LocationType)
	fmt.Println(user.Region)
	user.StoreId = r.request.PostForm.Get("shop[ShopCode]")
	user.FirstName = r.request.PostForm.Get("member[Name1]")
	user.LastName = r.request.PostForm.Get("member[Name2]")
	user.PhoneNumber = r.request.PostForm.Get("member[TelNo]")
	user.Email = r.request.PostForm.Get("member[Mail]")
	user.Products = basket
	user.OrderId = orderId
	user.Price = price

	// dom.PlaceOrder(user)

	currentTime := time.Now().Format("200602011504")
	r.AddKVWChildNode("Message", KVField{
		XMLName: xml.Name{Local: "contents"},
		Value:   "Thank you! Your order has been placed!",
	})
	r.AddKVNode("order_id", "1")
	r.AddKVNode("orderDay", currentTime)
	r.AddKVNode("hashKey", "Testing: 1, 2, 3")
	r.AddKVNode("hour", currentTime)

	// Remove the order data from the database
	_, err = pool.Exec(context.Background(), `UPDATE "user" SET order_id = $1, price = $2, basket = $3 WHERE wii_id = $4`, "", "", "[]", r.request.Header.Get("X-WiiID"))
	if err != nil {
		r.ReportError(err, http.StatusInternalServerError)
		return
	}

	// Post and log successful order !
	_ = dataDog.Incr("demae-dominos.orders_placed", nil, 1)
	PostDiscordWebhook(
		"A successful order has been processed!",
		fmt.Sprintf("The order was placed by user id %s", r.request.Header.Get("X-WiiID")),
		config.OrderWebhook,
		65311,
	)
}
