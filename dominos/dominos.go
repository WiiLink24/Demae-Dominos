package dominos

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/exp/slices"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (d *Dominos) StoreLookup(zipCode, address string) ([]Store, error) {
	queryParams := url.Values{
		"type": {"Delivery"},
		"c":    {zipCode},
		"s":    {address},
	}

	respChan := make(chan *http.Response)
	go d.sendAsyncGET(fmt.Sprintf("%s/power/store-locator?%s", d.apiURL, queryParams.Encode()), respChan)

	addressResponse := <-respChan
	defer addressResponse.Body.Close()
	respBytes, _ := io.ReadAll(addressResponse.Body)

	jsonData := map[string]any{}
	err := json.Unmarshal(respBytes, &jsonData)
	if err != nil {
		return nil, err
	}

	var stores []Store
	for i, storeData := range jsonData["Stores"].([]any) {
		if i == 5 {
			break
		}

		if !storeData.(map[string]any)["IsDeliveryStore"].(bool) {
			continue
		}

		storeAddress := storeData.(map[string]any)["AddressDescription"].(string)
		if storeData.(map[string]any)["LocationInfo"].(string) != "" {
			storeAddress = strings.Split(storeAddress, storeData.(map[string]any)["LocationInfo"].(string))[0]
		}
		store := Store{
			StoreID:  storeData.(map[string]any)["StoreID"].(string),
			Address:  strings.Replace(storeAddress, "\n", " ", -1),
			WaitTime: storeData.(map[string]any)["ServiceMethodEstimatedWaitMinutes"].(map[string]any)["Delivery"].(map[string]any)["Min"].(float64),
			IsOpen:   storeData.(map[string]any)["IsOpen"].(bool),
		}

		stores = append(stores, store)
	}

	return stores, nil
}

func (d *Dominos) AddressLookup(zipCode, address string) (*User, error) {
	queryParams := url.Values{
		"type": {"Delivery"},
		"c":    {zipCode},
		"s":    {address},
	}

	respChan := make(chan *http.Response)
	go d.sendAsyncGET(fmt.Sprintf("%s/power/store-locator?%s", d.apiURL, queryParams.Encode()), respChan)

	addressResponse := <-respChan
	defer addressResponse.Body.Close()
	respBytes, _ := io.ReadAll(addressResponse.Body)

	jsonData := map[string]any{}
	err := json.Unmarshal(respBytes, &jsonData)
	if err != nil {
		return nil, err
	}

	if jsonData["Status"].(float64) != 0 && jsonData["Status"].(float64) != 1 {
		return nil, fmt.Errorf("domino's returned a status code of %.0f", jsonData["Status"].(float64))
	}

	locationType := "House"
	apartmentNumber := ""
	if jsonData["Address"].(map[string]any)["UnitNumber"].(string) != "" {
		locationType = "Apartment"
		apartmentNumber = jsonData["Address"].(map[string]any)["UnitNumber"].(string)
	}

	return &User{
		Street:          jsonData["Address"].(map[string]any)["Street"].(string),
		ApartmentNumber: apartmentNumber,
		City:            jsonData["Address"].(map[string]any)["City"].(string),
		Region:          jsonData["Address"].(map[string]any)["Region"].(string),
		PostalCode:      jsonData["Address"].(map[string]any)["PostalCode"].(string),
		LocationType:    locationType,
		StreetName:      jsonData["Address"].(map[string]any)["StreetName"].(string),
		StreetNumber:    jsonData["Address"].(map[string]any)["StreetNumber"].(string),
	}, nil
}

func (d *Dominos) GetStoreInfo(storeId string) (*Store, error) {
	respChan := make(chan *http.Response)
	go d.sendAsyncGET(fmt.Sprintf("%s/power/store/%s/profile", d.apiURL, storeId), respChan)

	storeResponse := <-respChan
	defer storeResponse.Body.Close()
	respBytes, _ := io.ReadAll(storeResponse.Body)

	jsonData := map[string]any{}
	err := json.Unmarshal(respBytes, &jsonData)
	if err != nil {
		return nil, err
	}

	information := ""
	if jsonData["LocationInfo"] != nil {
		information = jsonData["LocationInfo"].(string)
	}

	return &Store{
		StoreID:      jsonData["StoreID"].(string),
		Address:      strings.Replace(jsonData["AddressDescription"].(string), "\n", " ", -1),
		WaitTime:     jsonData["ServiceMethodEstimatedWaitMinutes"].(map[string]any)["Delivery"].(map[string]any)["Min"].(float64),
		MinPrice:     jsonData["MinimumDeliveryOrderAmount"].(float64),
		IsOpen:       jsonData["IsOpen"].(bool),
		DetailedWait: jsonData["EstimatedWaitMinutes"].(string),
		Phone:        jsonData["Phone"].(string),
		ServiceHours: ServiceHours{},
		Information:  information,
	}, nil
}

func (d *Dominos) GetMenu(storeId string) ([]MenuCategory, error) {
	respChan := make(chan *http.Response)
	go d.sendAsyncGET(fmt.Sprintf("%s/power/store/%s/menu?lang=en&structured=true", d.apiURL, storeId), respChan)

	storeResponse := <-respChan
	defer storeResponse.Body.Close()
	respBytes, _ := io.ReadAll(storeResponse.Body)

	jsonData := map[string]any{}
	err := json.Unmarshal(respBytes, &jsonData)
	if err != nil {
		return nil, err
	}

	var menus []MenuCategory
	for _, menuData := range jsonData["Categorization"].(map[string]any)["Food"].(map[string]any)["Categories"].([]any) {
		if menuData.(map[string]any)["Name"].(string) == "" {
			// Ignore categories with no products
			continue
		}

		category := MenuCategory{
			Name: menuData.(map[string]any)["Name"].(string),
			Code: menuData.(map[string]any)["Code"].(string),
		}

		subcategories, ok := menuData.(map[string]any)["Categories"].([]any)
		if ok {
			// We have subcategories.
			for _, subcategory := range subcategories {
				if len(subcategory.(map[string]any)["Products"].([]any)) == 0 {
					// Ignore categories with no products
					continue
				}

				category.Categories = append(category.Categories, MenuCategory{
					Name:       subcategory.(map[string]any)["Name"].(string),
					Code:       subcategory.(map[string]any)["Code"].(string),
					Categories: nil,
				})
			}
		}

		menus = append(menus, category)
	}

	return menus, nil
}

func (d *Dominos) GetItemList(storeId string, menuCode string) ([]Item, error) {
	respChan := make(chan *http.Response)
	go d.sendAsyncGET(fmt.Sprintf("%s/power/store/%s/menu?lang=en&structured=true", d.apiURL, storeId), respChan)

	storeResponse := <-respChan
	defer storeResponse.Body.Close()
	respBytes, _ := io.ReadAll(storeResponse.Body)

	jsonData := map[string]any{}
	err := json.Unmarshal(respBytes, &jsonData)
	if err != nil {
		return nil, err
	}

	var itemNames []string
	for _, menuData := range jsonData["Categorization"].(map[string]any)["Food"].(map[string]any)["Categories"].([]any) {
		// It is possible that in a past loop we already got the items.
		if len(itemNames) != 0 {
			break
		}

		if menuData.(map[string]any)["Code"].(string) == menuCode {
			for _, a := range menuData.(map[string]any)["Products"].([]any) {
				itemNames = append(itemNames, a.(string))
			}

			if len(itemNames) == 0 {
				// For some reason Domino's has a parent code the same as a child code.
				// Regardless, no items mean this is a parent category.
				subcategories, ok := menuData.(map[string]any)["Categories"].([]any)
				if ok {
					for _, subcategory := range subcategories {
						if subcategory.(map[string]any)["Code"].(string) == menuCode {
							for _, a := range subcategory.(map[string]any)["Products"].([]any) {
								itemNames = append(itemNames, a.(string))
							}
							break
						}
					}
				}
			}
			break
		} else {
			// Here is where things get tricky. It is possible we have a menu that is nested. Unfortunately we need to iterate over the subcategory.
			// First check for the existence of subcategories.
			if len(menuData.(map[string]any)["Categories"].([]any)) == 0 {
				continue
			}

			subcategories, ok := menuData.(map[string]any)["Categories"].([]any)
			if ok {
				for _, subcategory := range subcategories {
					if subcategory.(map[string]any)["Code"].(string) == menuCode {
						for _, a := range subcategory.(map[string]any)["Products"].([]any) {
							itemNames = append(itemNames, a.(string))
						}
						break
					}
				}
			}
		}
	}

	var items []Item
	for name, itemData := range jsonData["Products"].(map[string]any) {
		if slices.Contains(itemNames, name) {
			item := Item{
				Name:        itemData.(map[string]any)["Name"].(string),
				Description: itemData.(map[string]any)["Description"].(string),
				Img:         itemData.(map[string]any)["Code"].(string),
				Items:       nil,
			}

			// Iterate over the variants to get the sizes
			var variants []string
			for _, a := range itemData.(map[string]any)["Variants"].([]any) {
				variants = append(variants, a.(string))
			}
			for variantName, variant := range jsonData["Variants"].(map[string]any) {
				if slices.Contains(variants, variantName) {
					item.Items = append(item.Items, ItemSize{
						Code:  variant.(map[string]any)["Code"].(string),
						Name:  variant.(map[string]any)["Name"].(string),
						Price: variant.(map[string]any)["Price"].(string),
					})
				}
			}

			items = append(items, item)
		}
	}

	return items, nil
}

func (d *Dominos) GetFoodPrice(storeId string, itemId string) (string, error) {
	respChan := make(chan *http.Response)
	go d.sendAsyncGET(fmt.Sprintf("%s/power/store/%s/menu?lang=en&structured=true", d.apiURL, storeId), respChan)

	storeResponse := <-respChan
	defer storeResponse.Body.Close()
	respBytes, _ := io.ReadAll(storeResponse.Body)

	jsonData := map[string]any{}
	err := json.Unmarshal(respBytes, &jsonData)
	if err != nil {
		return "", err
	}

	return jsonData["Variants"].(map[string]any)[itemId].(map[string]any)["Price"].(string), nil
}

func (d *Dominos) GetToppings(storeId string, itemId string) ([]Topping, error) {
	respChan := make(chan *http.Response)
	go d.sendAsyncGET(fmt.Sprintf("%s/power/store/%s/menu?lang=en&structured=true", d.apiURL, storeId), respChan)

	storeResponse := <-respChan
	defer storeResponse.Body.Close()
	respBytes, _ := io.ReadAll(storeResponse.Body)

	jsonData := map[string]any{}
	err := json.Unmarshal(respBytes, &jsonData)
	if err != nil {
		return nil, err
	}

	d.jsonResponse = &jsonData

	// First we retrieve the metadata for the item.
	productCode := jsonData["Variants"].(map[string]any)[itemId].(map[string]any)["ProductCode"].(string)
	productType := jsonData["Products"].(map[string]any)[productCode].(map[string]any)["ProductType"].(string)
	availableToppingsStr := jsonData["Products"].(map[string]any)[productCode].(map[string]any)["AvailableToppings"].(string)

	if availableToppingsStr == "" {
		return nil, nil
	}

	// Convert the available toppings into a list
	availableToppings := strings.Split(availableToppingsStr, ",")
	for i, topping := range availableToppings {
		if strings.Contains(topping, "=") {
			availableToppings[i] = strings.Split(topping, "=")[0]
		}
	}

	var toppings []Topping
	for name, toppingData := range jsonData["Toppings"].(map[string]any)[productType].(map[string]any) {
		if !slices.Contains(availableToppings, name) {
			// The current topping is not available for this product.
			continue
		}

		topping := Topping{
			Code:  name,
			Name:  toppingData.(map[string]any)["Name"].(string),
			Group: 0,
		}

		if toppingData.(map[string]any)["Tags"].(map[string]any)["Sauce"] != nil {
			topping.Group = Sauce
		} else if toppingData.(map[string]any)["Tags"].(map[string]any)["NonMeat"] != nil && toppingData.(map[string]any)["Tags"].(map[string]any)["NonMeat"].(any).(bool) {
			topping.Group = NonMeat
		} else {
			topping.Group = Meat
		}

		toppings = append(toppings, topping)
	}

	return toppings, nil
}

func (d *Dominos) GetSides(storeId string, itemId string) ([]Topping, error) {
	var jsonData map[string]any
	if d.jsonResponse != nil {
		jsonData = *d.jsonResponse
	} else {
		respChan := make(chan *http.Response)
		go d.sendAsyncGET(fmt.Sprintf("%s/power/store/%s/menu?lang=en&structured=true", d.apiURL, storeId), respChan)

		storeResponse := <-respChan
		defer storeResponse.Body.Close()
		respBytes, _ := io.ReadAll(storeResponse.Body)

		jsonData = map[string]any{}
		err := json.Unmarshal(respBytes, &jsonData)
		if err != nil {
			return nil, err
		}
	}

	// First we retrieve the metadata for the item.
	productCode := jsonData["Variants"].(map[string]any)[itemId].(map[string]any)["ProductCode"].(string)
	productType := jsonData["Products"].(map[string]any)[productCode].(map[string]any)["ProductType"].(string)
	availableSidesStr := jsonData["Products"].(map[string]any)[productCode].(map[string]any)["AvailableSides"].(string)

	if availableSidesStr == "" {
		return nil, nil
	}

	// Convert the available sides into a list
	availableSides := strings.Split(availableSidesStr, ",")

	if jsonData["Sides"].(map[string]any)[productType] == nil {
		return nil, nil
	}

	var sides []Topping
	for name, sideData := range jsonData["Sides"].(map[string]any)[productType].(map[string]any) {
		if !slices.Contains(availableSides, name) {
			// The current topping is not available for this product.
			continue
		}

		sides = append(sides, Topping{
			Code:  name,
			Name:  sideData.(map[string]any)["Name"].(string),
			Group: 0,
		})
	}

	return sides, nil
}

func (d *Dominos) AddItem(storeId, itemId, quantity string, extraToppings []Topping) (map[string]any, error) {
	respChan := make(chan *http.Response)
	go d.sendAsyncGET(fmt.Sprintf("%s/power/store/%s/menu?lang=en&structured=true", d.apiURL, storeId), respChan)

	storeResponse := <-respChan
	defer storeResponse.Body.Close()
	respBytes, _ := io.ReadAll(storeResponse.Body)
	jsonData := map[string]any{}
	err := json.Unmarshal(respBytes, &jsonData)
	if err != nil {
		return nil, err
	}

	defaultToppings := jsonData["Variants"].(map[string]any)[itemId].(map[string]any)["Tags"].(map[string]any)["DefaultToppings"].(string)
	defaultSides := jsonData["Variants"].(map[string]any)[itemId].(map[string]any)["Tags"].(map[string]any)["DefaultSides"].(string)

	options := make(map[string]any)
	if defaultToppings != "" {
		toppings := strings.Split(defaultToppings, ",")
		for _, topping := range toppings {
			toppingArray := strings.Split(topping, "=")
			options[toppingArray[0]] = make(map[string]string)
			options[toppingArray[0]] = map[string]string{
				"1/1": toppingArray[1],
			}
		}
	}
	if defaultSides != "" {
		sides := strings.Split(defaultSides, ",")
		for _, side := range sides {
			sideArray := strings.Split(side, "=")
			options[sideArray[0]] = sideArray[1]
		}
	}

	// Now we add extra toppings if need be
	for _, topping := range extraToppings {
		if topping.Group == Side {
			if _, ok := options[topping.Code]; ok {
				// Ignore the added side as this item already contains it
				continue
			} else {
				// Side does not exist on item, add it
				options[topping.Code] = "1"
			}
		} else {
			if _, ok := options[topping.Code]; ok {
				if topping.Group != Sauce {
					// We do not maximize sauce.
					// Maximize the amount of topping since this item already contains it
					options[topping.Code] = map[string]string{
						"1/1": "1.5",
					}
				}
			} else {
				// Topping does not exist on item, add it
				options[topping.Code] = map[string]string{
					"1/1": "1",
				}
			}
		}
	}

	return map[string]any{
		"Code":    itemId,
		"Qty":     quantity,
		"ID":      "1",
		"isNew":   true,
		"Options": options,
	}, nil
}

func (d *Dominos) GetItemPrice(storeId, itemID string) (string, string, error) {
	respChan := make(chan *http.Response)
	go d.sendAsyncGET(fmt.Sprintf("%s/power/store/%s/menu?lang=en&structured=true", d.apiURL, storeId), respChan)

	storeResponse := <-respChan
	defer storeResponse.Body.Close()
	respBytes, _ := io.ReadAll(storeResponse.Body)

	jsonData := map[string]any{}
	err := json.Unmarshal(respBytes, &jsonData)
	if err != nil {
		return "", "", err
	}

	return jsonData["Variants"].(map[string]any)[itemID].(map[string]any)["Name"].(string), jsonData["Variants"].(map[string]any)[itemID].(map[string]any)["Price"].(string), nil
}

func (d *Dominos) GetPrice(user *User) (*Basket, error) {
	payload := map[string]any{}
	payload["Order"] = map[string]any{
		"Address": map[string]any{
			"Street":       user.Street,
			"City":         user.City,
			"Region":       user.Region,
			"PostalCode":   user.PostalCode,
			"Type":         user.LocationType,
			"StreetName":   user.StreetName,
			"StreetNumber": user.StreetNumber,
		},
		"Coupons":               []any{},
		"CustomerID":            "",
		"Email":                 "",
		"Extension":             "",
		"FirstName":             "",
		"LastName":              "",
		"LanguageCode":          "en",
		"metaData":              nil,
		"OrderChannel":          "OLO",
		"OrderID":               "",
		"OrderMethod":           "Web",
		"OrderTaker":            nil,
		"Payments":              []any{},
		"Phone":                 "",
		"PhonePrefix":           "",
		"Products":              user.Products,
		"ServiceMethod":         "Delivery",
		"SourceOrganizationURI": d.apiURL,
		"StoreID":               user.StoreId,
		"Tags":                  map[string]any{},
		"Version":               "1.0",
		"NoCombine":             true,
		"Partners":              map[string]any{},
		"HotspotsLite":          false,
		"OrderInfoCollection":   []any{},
	}

	// Marshal JSON then send to Domino's
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// First we validate the order
	validateOrder1Chan := make(chan *http.Response)
	go d.sendAsyncPOST(fmt.Sprintf("%s/power/validate-order", d.apiURL), data, validateOrder1Chan)

	validateOrder1Response := <-validateOrder1Chan
	defer validateOrder1Response.Body.Close()
	respBytes, _ := io.ReadAll(validateOrder1Response.Body)

	// Retrieve order id from response
	jsonData := map[string]any{}
	err = json.Unmarshal(respBytes, &jsonData)
	if err != nil {
		return nil, err
	}

	payload["Order"].(map[string]any)["OrderID"] = jsonData["Order"].(map[string]any)["OrderID"].(string)
	data, err = json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Register order with Domino's
	validateOrder2Chan := make(chan *http.Response)
	go d.sendAsyncPOST(fmt.Sprintf("%s/power/validate-order", d.apiURL), data, validateOrder2Chan)

	validateOrder2Response := <-validateOrder2Chan
	defer validateOrder2Response.Body.Close()
	respBytes, _ = io.ReadAll(validateOrder2Response.Body)

	jsonData = map[string]any{}
	err = json.Unmarshal(respBytes, &jsonData)
	if err != nil {
		return nil, err
	}

	// Finally, retrieve order price.
	payload["Order"].(map[string]any)["OrderID"] = jsonData["Order"].(map[string]any)["OrderID"].(string)
	payload["Order"].(map[string]any)["metaData"] = map[string]string{"orderFunnel": "payments"}
	data, err = json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	priceOrderChan := make(chan *http.Response)
	go d.sendAsyncPOST(fmt.Sprintf("%s/power/price-order", d.apiURL), data, priceOrderChan)

	priceOrderResponse := <-priceOrderChan
	defer priceOrderResponse.Body.Close()
	respBytes, _ = io.ReadAll(priceOrderResponse.Body)

	// Now we construct what we will be sending to the channel
	jsonData = map[string]any{}
	err = json.Unmarshal(respBytes, &jsonData)
	if err != nil {
		return nil, err
	}

	if jsonData["Status"].(float64) != 0 && jsonData["Status"].(float64) != 1 {
		return nil, fmt.Errorf("domino's returned a status code of %.0f\nError: %s", jsonData["Status"].(float64), jsonData["StatusItems"].([]any)[0].(map[string]any)["Code"].(string))
	}

	var items []BasketItem
	for _, rawArray := range jsonData["Order"].(map[string]any)["Products"].([]any) {
		itemData := rawArray.(map[string]any)
		name := itemData["Name"].(string)
		price := itemData["Price"].(float64)
		amount := itemData["Amount"].(float64)

		var options []string
		if _, ok := itemData["descriptions"]; ok {
			options = strings.Split(itemData["descriptions"].([]any)[0].(map[string]any)["value"].(string), ",")
		}

		items = append(items, BasketItem{
			Code:     itemData["Code"].(string),
			Name:     &name,
			Price:    &price,
			Amount:   &amount,
			Quantity: int(itemData["Qty"].(float64)),
			Options:  options,
		})
	}

	return &Basket{
		Items:       items,
		BasketPrice: jsonData["Order"].(map[string]any)["Amounts"].(map[string]any)["Menu"].(float64),
		ChargePrice: jsonData["Order"].(map[string]any)["Amounts"].(map[string]any)["Tax"].(float64),
		TotalPrice:  jsonData["Order"].(map[string]any)["Amounts"].(map[string]any)["Customer"].(float64),
		OrderId:     jsonData["Order"].(map[string]any)["OrderID"].(string),
	}, nil
}

func (d *Dominos) PlaceOrder(info *User) error {
	payload := map[string]any{}
	payload["Order"] = map[string]any{
		"Address": map[string]any{
			"Street":               info.Street,
			"City":                 info.City,
			"Region":               info.Region,
			"PostalCode":           info.PostalCode,
			"Type":                 info.LocationType,
			"StreetName":           info.StreetName,
			"StreetNumber":         info.StreetNumber,
			"DeliveryInstructions": "None",
		},
		"Coupons":      []any{},
		"CustomerID":   "",
		"Email":        info.Email,
		"Extension":    "",
		"FirstName":    info.FirstName,
		"LastName":     info.LastName,
		"LanguageCode": "en",
		"OrderChannel": "OLO",
		"OrderID":      info.OrderId,
		"OrderMethod":  "Web",
		"OrderTaker":   nil,
		"Payments": []map[string]any{
			{
				"Type":            "Cash",
				"Amount":          info.Price,
				"Number":          "",
				"CardType":        "",
				"Expiration":      "",
				"SecurityCode":    "",
				"PostalCode":      "",
				"ProviderID":      "",
				"PaymentMethodID": "",
				"OTP":             "",
				"gpmPaymentType":  "",
			},
		},
		"Phone":                 info.PhoneNumber,
		"PhonePrefix":           "",
		"Products":              info.Products,
		"ServiceMethod":         "Delivery",
		"SourceOrganizationURI": "order.dominos.com",
		"StoreID":               info.StoreId,
		"Tags":                  map[string]any{},
		"Version":               "1.0",
		"NoCombine":             true,
		"Partners":              map[string]any{},
		"HotspotsLite":          true,
		"OrderInfoCollection":   []any{},
		"metaData": map[string]any{
			"orderFunnel":        "payments",
			"calculateNutrition": "true",
			"isDomChat":          0,
			"ABTests":            []any{},
			"contactless":        true,
		},
	}

	if info.ApartmentNumber != "" {
		payload["Order"].(map[string]any)["Address"].(map[string]any)["AddressLine2"] = info.ApartmentNumber
	}

	data, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}

	placeOrderChan := make(chan *http.Response)
	go d.sendAsyncPOST(fmt.Sprintf("%s/power/place-order", d.apiURL), data, placeOrderChan)

	placeOrderResponse := <-placeOrderChan
	defer placeOrderResponse.Body.Close()
	respBytes, _ := io.ReadAll(placeOrderResponse.Body)

	jsonData := map[string]any{}
	err = json.Unmarshal(respBytes, &jsonData)
	if err != nil {
		return err
	}

	if jsonData["Status"].(float64) != 0 && jsonData["Status"].(float64) != 1 {
		return fmt.Errorf("domino's returned a status code of %.0f", jsonData["Status"].(float64))
	}

	return nil
}

func (d *Dominos) sendAsyncGET(url string, rc chan *http.Response) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	d.setHeaders(req)
	response, _ := client.Do(req)
	rc <- response
}

func (d *Dominos) sendAsyncPOST(url string, data []byte, rc chan *http.Response) {
	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(data))
	d.setPostHeaders(req)
	response, _ := client.Do(req)
	rc <- response
}

func (d *Dominos) setHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Market", string(d.country))
	req.Header.Set("DPZ-Language", "en")
	req.Header.Set("DPZ-Market", string(d.country))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Safari/605.1.15")
}

func (d *Dominos) setPostHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Origin", d.apiURL)
	req.Header.Set("Referer", fmt.Sprintf("%s/assets/build/xdomain/proxy.html", d.apiURL))
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Market", string(d.country))
	req.Header.Set("DPZ-Language", "en")
	req.Header.Set("DPZ-Market", string(d.country))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Safari/605.1.15")
}
