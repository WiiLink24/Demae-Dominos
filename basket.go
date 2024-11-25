package main

import (
	"DemaeDominos/dominos"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/mitchellh/go-wordwrap"
	"strconv"
	"strings"
	"time"
)

const (
	DoesAuthKeyExist  = `SELECT EXISTS(SELECT 1 FROM "user" WHERE "user".wii_id = $1 AND "user".auth_key IS NOT NULL)`
	QueryUserBasket   = `SELECT "user".basket, "user".auth_key FROM "user" WHERE "user".wii_id = $1 LIMIT 1`
	QueryUserForOrder = `SELECT "user".basket, "user".price, "user".order_id FROM "user" WHERE "user".wii_id = $1 LIMIT 1`
	InsertAuthkey     = `UPDATE "user" SET auth_key = $1 WHERE wii_id = $2`
	ClearBasket       = `UPDATE "user" SET order_id = $1, price = $2, basket = $3 WHERE wii_id = $4`
	UpdateUserBasket  = `UPDATE "user" SET basket = $1 WHERE wii_id = $2`
	UpdateOrderId     = `UPDATE "user" SET order_id = $1, price = $2 WHERE wii_id = $3`
)

func authKey(r *Response) {
	authKeyValue, err := uuid.DefaultGenerator.NewV1()
	if err != nil {
		r.ReportError(err)
		return
	}

	// First we query to determine if the user already has an auth key. If they do, reset the basket.
	var authExists bool
	row := pool.QueryRow(context.Background(), DoesAuthKeyExist, r.request.Header.Get("X-WiiID"))
	err = row.Scan(&authExists)
	if err != nil {
		r.ReportError(err)
		return
	}

	if authExists {
		_, err = pool.Exec(context.Background(), ClearBasket, "", "", "[]", r.request.Header.Get("X-WiiID"))
		if err != nil {
			r.ReportError(err)
			return
		}
	}

	_, err = pool.Exec(context.Background(), InsertAuthkey, authKeyValue.String(), r.request.Header.Get("X-WiiID"))
	if err != nil {
		r.ReportError(err)
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
	_, err := pool.Exec(context.Background(), ClearBasket, "", "", "[]", r.request.Header.Get("X-WiiID"))
	if err != nil {
		r.ReportError(err)
		return
	}
}

func basketDelete(r *Response) {
	basketNumber, err := strconv.Atoi(r.request.URL.Query().Get("basketNo"))
	if err != nil {
		r.ReportError(err)
		return
	}

	var lastBasket string
	row := pool.QueryRow(context.Background(), QueryUserBasket, r.request.Header.Get("X-WiiID"))
	err = row.Scan(&lastBasket, nil)
	if err != nil {
		r.ReportError(err)
		return
	}

	var actualBasket []map[string]any
	err = json.Unmarshal([]byte(lastBasket), &actualBasket)
	if err != nil {
		r.ReportError(err)
		return
	}

	// Pop last element
	actualBasket = append(actualBasket[:basketNumber-1], actualBasket[basketNumber:]...)

	// Convert basket to JSON then insert to database
	jsonStr, err := json.Marshal(actualBasket)
	if err != nil {
		r.ReportError(err)
		return
	}

	_, err = pool.Exec(context.Background(), UpdateUserBasket, jsonStr, r.request.Header.Get("X-WiiID"))
	if err != nil {
		r.ReportError(err)
		return
	}
}

func basketAdd(r *Response) {
	shopCode := r.request.PostForm.Get("shopCode")
	itemCode := r.request.PostForm.Get("itemCode")
	quantity := r.request.PostForm.Get("quantity")

	var lastBasket string
	var err error
	r.dominos, err = dominos.NewDominos(r.request)
	if err != nil {
		r.ReportError(err)
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
	basket, err := r.dominos.AddItem(shopCode, itemCode, quantity, toppings)
	if err != nil {
		r.ReportError(err)
		return
	}

	var _authKey string
	row := pool.QueryRow(context.Background(), QueryUserBasket, r.GetHollywoodId())
	err = row.Scan(&lastBasket, &_authKey)
	if err != nil {
		r.ReportError(err)
		return
	}

	var actualBasket []map[string]any
	err = json.Unmarshal([]byte(lastBasket), &actualBasket)
	if err != nil {
		r.ReportError(err)
		return
	}

	actualBasket = append(actualBasket, basket)

	// Convert basket to JSON then insert to database
	jsonStr, err := json.Marshal(actualBasket)
	if err != nil {
		r.ReportError(err)
		return
	}

	_, err = pool.Exec(context.Background(), UpdateUserBasket, jsonStr, r.GetHollywoodId())
	if err != nil {
		r.ReportError(err)
		return
	}
}

func basketList(r *Response) {
	address := r.request.Header.Get("X-Address")
	postalCode := r.request.Header.Get("X-Postalcode")

	var basketStr string
	row := pool.QueryRow(context.Background(), QueryUserBasket, r.GetHollywoodId())
	err := row.Scan(&basketStr, nil)
	if err != nil {
		r.ReportError(err)
		return
	}

	var basket []map[string]any
	err = json.Unmarshal([]byte(basketStr), &basket)
	if err != nil {
		r.ReportError(err)
		return
	}

	r.dominos, err = dominos.NewDominos(r.request)
	if err != nil {
		r.ReportError(err)
		return
	}

	user, err := r.dominos.AddressLookup(postalCode, address)
	if err != nil {
		r.ReportError(err)
		return
	}

	user.StoreId = r.request.URL.Query().Get("shopCode")
	user.Products = basket

	items, err := r.dominos.GetPrice(user)
	if err != nil {
		r.ReportError(err)
		return
	}

	// Add order ID and price to database
	_, err = pool.Exec(context.Background(), UpdateOrderId, items.OrderId, fmt.Sprintf("%.2f", items.TotalPrice), r.GetHollywoodId())
	if err != nil {
		r.ReportError(err)
		return
	}

	basketPrice := KVField{
		XMLName: xml.Name{Local: "basketPrice"},
		Value:   fmt.Sprintf("$%.2f", items.BasketPrice),
	}

	chargePrice := KVField{
		XMLName: xml.Name{Local: "chargePrice"},
		Value:   fmt.Sprintf("$%.2f", items.ChargePrice),
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
		name := wordwrap.WrapString(item.Name, 29)
		for i, s := range strings.Split(name, "\n") {
			switch i {
			case 0:
				name = s
				break
			case 1:
				name += "\n"
				name += s
				break
			default:
				// If it is too long it becomes ... so we are fine
				name += " " + s
				break
			}
		}

		priceStr := fmt.Sprintf("$%.2f", item.Price)
		amountStr := fmt.Sprintf("$%.2f", item.Amount)
		basketItems = append(basketItems, BasketItem{
			XMLName:       xml.Name{Local: fmt.Sprintf("container%d", i)},
			BasketNo:      CDATA{i + 1},
			MenuCode:      CDATA{1},
			ItemCode:      CDATA{item.Code},
			Name:          CDATA{name},
			Price:         CDATA{priceStr},
			Size:          CDATA{""},
			IsSoldout:     CDATA{BoolToInt(false)},
			Quantity:      CDATA{item.Quantity},
			SubTotalPrice: CDATA{amountStr},
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

	row := pool.QueryRow(context.Background(), QueryUserForOrder, r.GetHollywoodId())
	err := row.Scan(&basketStr, &price, &orderId)
	if err != nil {
		r.ReportError(err)
		return
	}

	var basket []map[string]any
	err = json.Unmarshal([]byte(basketStr), &basket)
	if err != nil {
		r.ReportError(err)
		return
	}

	r.dominos, err = dominos.NewDominos(r.request)
	if err != nil {
		r.ReportError(err)
		return
	}

	user, err := r.dominos.AddressLookup(r.request.PostForm.Get("member[PostNo]"), r.request.PostForm.Get("member[Address5]"))
	if err != nil {
		r.ReportError(err)
		return
	}

	user.StoreId = r.request.PostForm.Get("shop[ShopCode]")
	user.FirstName = r.request.PostForm.Get("member[Name1]")
	user.LastName = r.request.PostForm.Get("member[Name2]")
	user.PhoneNumber = r.request.PostForm.Get("member[TelNo]")
	user.Email = r.request.PostForm.Get("member[Mail]")
	user.Products = basket
	user.OrderId = orderId
	user.Price = price

	// If the error does fail we should alert the user and allow for the basket to be cleared.
	didError := false
	err = r.dominos.PlaceOrder(user)
	if err != nil {
		PostDiscordWebhook(
			"Performing error failed.",
			fmt.Sprintf("The order was placed by user id %s", r.GetHollywoodId()),
			config.OrderWebhook,
			65311,
		)
		didError = true
	}

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
	_, err = pool.Exec(context.Background(), ClearBasket, "", "", "[]", r.GetHollywoodId())
	if err != nil || didError {
		r.ReportError(err)
		return
	}

	// Post and log successful order!
	PostDiscordWebhook(
		"A successful order has been processed!",
		fmt.Sprintf("The order was placed by user id %s", r.request.Header.Get("X-WiiNo")),
		config.OrderWebhook,
		65311,
	)
}
