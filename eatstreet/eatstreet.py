import requests
import json
import datetime

from eatstreet.country import country, Country

eatstreet_api_url = "https://eatstreet.com/api/v2"

image_urls = {
    Country.USA: "https://cache.dominos.com/olo/6_92_1/assets/build/market/US/_en/images/img/products/larges",
    Country.CANADA: "https://cache.dominos.com/nolo/ca/en/6_88_3/assets/build/market/CA/_en/images/img/products/larges",
}


def get_store_info_eatstreet(store_id):
    jsondata = {}

    response = requests.get(
        f"{eatstreet_api_url}/restaurant-cards/nearby?latitude=30.40049&longitude=-97.70703",
        verify=True,
        headers=get_headers(),
    )

    s1 = response.json()

    for store in s1["cards"]:
        if str(store["id"]) == str(store_id):
            store_id = store["id"]
            response = requests.get(
                f"{eatstreet_api_url}/restaurants/{store_id}",
                verify=True,
                headers=get_headers(),
            )
            s2 = response.json()

            wait_time = store["maxDeliveryEta"]
            min_price = store["pickupMin"]
            address = s2["fullAddress"]
            detailed_wait = int((store["minDeliveryEta"] + store["maxDeliveryEta"]) / 2)
            phone = "911"
            is_open = True
            service_hours = {}
            service_hours["OpenTime"] = "0"
            service_hours["CloseTime"] = "0"
            name = store["name"]

            jsondata.update(
                {
                    "wait_time": wait_time,
                    "min_price": min_price,
                    "address": address,
                    "detailed_wait": detailed_wait,
                    "phone": phone,
                    "is_open": is_open,
                    "service_hours": service_hours,
                    "name": name,
                }
            )
            return jsondata


def lookup_category(categories, category):
    cat = {}

    cat["American Food"] = 6
    cat["Burgers"] = 7
    cat["Catering"] = 9
    cat["Chinese Food"] = 5
    cat["Coffee & Tea"] = 10
    cat["Dessert"] = 10
    cat["Fast Food"] = 7
    cat["Indian Food"] = 8
    cat["Pizza"] = 1
    cat["Smoothies & Juices"] = 10
    cat["Subs & Sandwiches"] = 7

    for k, v in cat.items():
        if k in categories:
            return v


def address_lookup_eatstreet(zipcode, address, category):
    jsondata = {"user": {}, "restaurants": []}
    response = requests.get(
        f"{eatstreet_api_url}/restaurant-cards/nearby?latitude=30.40049&longitude=-97.70703",
        headers=get_headers(),
    )
    s1 = response.json()

    count = 0

    street = "hoge"
    streetnum = "hoge"
    streetname = "hoge"
    city = "hoge"
    region = "hoge"
    postal = "hoge"

    """street = s2["address"]
    streetnum = s2["address"]  # change me
    streetname = s2["address"]
    city = s2["city"]
    region = s2["state"]
    postal = s2["zip"]"""

    jsondata["user"].update(
        {
            "street": street,
            "city": city,
            "region": region,
            "postalcode": postal,
            "type": "House",
            "streetname": streetname,
            "streetnumber": streetnum,
        }
    )

    for store in s1["cards"]:
        if lookup_category(store["cuisines"], int(category)) == int(category):
            store_id = store["id"]
            response = requests.get(
                f"{eatstreet_api_url}/restaurants/{store_id}",
                verify=True,
                headers=get_headers(),
            )
            s2 = response.json()

            count = count + 1
            print(count)
            if count == 1:
                if 1 + 1 == 2:
                    """if (
                        not store["orderingAvailability"]["delivery"] == "OPEN_NOW"
                        and store["futureOrdering"]
                    ):"""
                    if 1 + 1 == 3:
                        jsondata["restaurants"].append(
                            {
                                "id": store_id,
                                "address": street,
                                "time": store["minDeliveryEta"],
                                "isOpen": False,
                                "isDelivery": store["IsDeliveryStore"],
                            }
                        )
                    else:
                        jsondata["restaurants"].append(
                            {
                                "id": store_id,
                                "address": "Hell",
                                "time": store["minDeliveryEta"],
                                "isOpen": True,
                                "isDelivery": True,
                            }
                        )

    x = json.dumps(jsondata)
    return x


def get_menu(store_id):
    food = {"categories": {}}
    response = requests.get(
        f"{eatstreet_api_url}/restaurants/{store_id}/menu",
        verify=True,
        headers=get_headers(),
    )
    s2 = response.json()
    i = 0
    for cat in s2["categories"]:
        i += 1
        if i > 0:
            food["categories"].update({cat["id"]: {}})
            for prods in s2["products"]:
                if s2["categoryId"] == cat["id"]:
                    food["categories"][cat["id"]].update(
                        {
                            prods: {
                                "name": s2["products"][prods]["name"].split(" - ")[0],
                                "desc": s2["products"][prods]["description"],
                                "items": [],
                            }
                        }
                    )
                    food["categories"][cat["id"]][s2["products"][prods]["id"]][
                        "items"
                    ].append(
                        {
                            "code": s2["products"][prods]["id"],
                            "name": s2["products"][prods]["name"].split(" - ")[0],
                            "price": s2["products"][prods]["price"],
                            "img": "",
                            "size": s2["products"][prods]["name"].split(" - ")[1],
                        }
                    )
    x = json.dumps(food)
    return x


def get_recommended_eatstreet(store_id):
    count = 0
    food = {"recommended": {}}
    response = requests.get(
        f"{eatstreet_api_url}/restaurants/{store_id}/menu",
        verify=True,
        headers=get_headers(),
    )
    s2 = response.json()
    cat = s2["categories"][0]
    for prods in s2["products"]:
        if str(prods["categoryId"]) == str(cat["id"]):
            count = count + 1
            if count <= 4:
                food["recommended"].update({prods["id"]: {}})
                food["recommended"][prods["id"]].update(
                    {
                        "code": cat,
                        "name": prods["name"].split(" - ")[0],
                        "code": prods["id"],
                    }
                )

    x = json.dumps(food)
    return x


def get_food_price(store_id, item_id):
    food = {"food": {}}
    response = requests.get(
        f"{eatstreet_api_url}/power/store/{store_id}/menu?lang=en&structured=true",
        verify=True,
        headers=get_headers(),
    )
    s2 = response.json()
    for cat in s2["Variants"][item_id]:
        food["food"].update(
            {
                "name": s2["Variants"][item_id]["Name"],
                "price": s2["Variants"][item_id]["Price"],
            }
        )

    x = json.dumps(food)
    return x


def add_food_item(store_id, code, qty, num):
    tagok = json.loads("{}")
    response = requests.get(
        f"{eatstreet_api_url}/power/store/{store_id}/menu?lang=en&structured=true",
        verify=True,
        headers=get_headers(),
    )
    s2 = response.json()
    itemtags = s2["Variants"][code]["Tags"]["DefaultToppings"]
    itemsides = s2["Variants"][code]["Tags"]["DefaultSides"]
    if itemtags != "":
        tags = itemtags.split(",")
        for tag in tags:
            bruh = tag.split("=")
            tagok.update({bruh[0]: {"1/1": bruh[1]}})
    if itemsides != "":
        tags2 = itemsides.split(",")
        for tag2 in tags2:
            bruh = tag2.split("=")
            tagok.update({bruh[0]: bruh[1]})

    itemer = {
        "Code": code,
        "Qty": int(qty),
        "ID": int(num),
        "isNew": True,
        "Options": tagok,
    }
    return itemer


def get_price(
    street,
    city,
    region,
    postal,
    location_type,
    street_name,
    street_num,
    store_id,
    prods,
):
    bigorder = {
        "Order": {
            "Address": {
                "Street": street,
                "City": city,
                "Region": region,
                "PostalCode": postal,
                "Type": location_type,
                "StreetName": street_name,
                "StreetNumber": street_num,
            },
            "Coupons": [],
            "CustomerID": "",
            "Email": "",
            "Extension": "",
            "FirstName": "",
            "LastName": "",
            "LanguageCode": "en",
            "OrderChannel": "OLO",
            "OrderID": "",
            "OrderMethod": "Web",
            "OrderTaker": None,
            "Payments": [],
            "Phone": "",
            "PhonePrefix": "",
            "Products": prods,
            "ServiceMethod": "Delivery",
            "SourceOrganizationURI": eatstreet_api_url,
            "StoreID": store_id,
            "Tags": {},
            "Version": "1.0",
            "NoCombine": True,
            "Partners": {},
            "HotspotsLite": False,
            "OrderInfoCollection": [],
        }
    }
    headers2 = {
        "Accept": "application/json, text/javascript, */*; q=0.01",
        "Market": str(country.value),
        "DPZ-Language": "en",
        "DPZ-Market": str(country.value),
        "User-Agent": "Mozilla/5.0 (iPhone; CPU OS 14_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) FxiOS/31.0 Mobile/15E148 Safari/605.1.15",
        "Accept-Language": "en-US,en;q=0.5",
        "Content-Type": "application/json; charset=utf-8",
        "Origin": f"{eatstreet_api_url}",
        "Referer": f"{eatstreet_api_url}/assets/build/xdomain/proxy.html",
    }
    response = requests.post(
        f"{eatstreet_api_url}/power/validate-order",
        verify=True,
        headers=headers2,
        data=json.dumps(bigorder),
    )
    orderiddi = json.loads(response.text)["Order"]["OrderID"]
    bigorder["Order"]["OrderID"] = orderiddi
    response = requests.post(
        f"{eatstreet_api_url}/power/validate-order",
        verify=True,
        headers=headers2,
        data=json.dumps(bigorder),
    )
    orderiddi = json.loads(response.text)["Order"]["OrderID"]
    bigorder["Order"]["OrderID"] = orderiddi
    bigorder["Order"].update({"metaData": {"orderFunnel": "payments"}})
    response = requests.post(
        f"{eatstreet_api_url}/power/price-order",
        verify=True,
        headers=headers2,
        data=json.dumps(bigorder),
    )
    return response.text


def place_order(
    street,
    city,
    region,
    postal,
    location_type,
    street_name,
    street_num,
    store_id,
    prods,
    firstname,
    lastname,
    email,
    phone_num,
    delivery_instructions,
    order_id,
    money,
):
    headers2 = {
        "Accept": "application/json, text/javascript, */*; q=0.01",
        "Market": str(country.value),
        "DPZ-Language": "en",
        "DPZ-Market": str(country.value),
        "User-Agent": "Mozilla/5.0 (iPhone; CPU OS 14_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) FxiOS/31.0 Mobile/15E148 Safari/605.1.15",
        "Accept-Language": "en-US,en;q=0.5",
        "Content-Type": "application/json; charset=utf-8",
        "Origin": f"{eatstreet_api_url}",
        "Referer": f"{eatstreet_api_url}/assets/build/xdomain/proxy.html",
    }
    bigorder = {
        "Order": {
            "Address": {
                "Street": street,
                "City": city,
                "Region": region,
                "PostalCode": postal,
                "Type": location_type,
                "StreetName": street_name,
                "StreetNumber": street_num,
            },
            "Coupons": [],
            "CustomerID": "",
            "Email": email,
            "Extension": "",
            "FirstName": firstname,
            "LastName": lastname,
            "LanguageCode": "en",
            "OrderChannel": "OLO",
            "OrderID": order_id,
            "OrderMethod": "Web",
            "OrderTaker": None,
            "Payments": [],
            "Phone": phone_num,
            "PhonePrefix": "",
            "Products": prods,
            "ServiceMethod": "Delivery",
            "SourceOrganizationURI": eatstreet_api_url,
            "StoreID": store_id,
            "Tags": {},
            "Version": "1.0",
            "NoCombine": True,
            "Partners": {},
            "HotspotsLite": False,
            "OrderInfoCollection": [],
        }
    }
    bigorder["Order"]["Address"].update({"DeliveryInstructions": delivery_instructions})
    bigorder["Order"].update(
        {
            "Payments": [
                {
                    "Type": "Cash",
                    "Amount": float(money),
                    "Number": "",
                    "CardType": "",
                    "Expiration": "",
                    "SecurityCode": "",
                    "PostalCode": "",
                    "ProviderID": "",
                    "PaymentMethodID": "",
                    "OTP": "",
                    "gpmPaymentType": "",
                }
            ]
        }
    )
    bigorder["Order"].update({"metaData": {"orderFunnel": "payments"}})
    bigorder["Order"]["metaData"].update(
        {
            "calculateNutrition": "true",
            "isDomChat": 0,
            "ABTests": [],
            "contactless": True,
        }
    )
    response = requests.post(
        f"{eatstreet_api_url}/power/place-order",
        verify=True,
        headers=headers2,
        data=json.dumps(bigorder),
    )
    return response.text


def get_headers() -> dict:
    return {
        "Accept": "application/json, text/javascript, */*; q=0.01",
        "Market": str(country.value),
        "DPZ-Language": "en",
        "DPZ-Market": str(country.value),
        "User-Agent": "Mozilla/5.0 (iPhone; CPU OS 14_4 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) "
        "FxiOS/31.0 Mobile/15E148 Safari/605.1.15 ",
    }


def convert_datetime_today_to_date() -> str:
    """datetime.datetime.today().weekday() returns an int with
    Monday being 0 and Sunday being 6. Dominos uses the first 3 letters of each day rather than an int."""

    current_day = datetime.datetime.today().weekday()
    match current_day:
        case 0:
            return "Mon"
        case 1:
            return "Tue"
        case 2:
            return "Wed"
        case 3:
            return "Thu"
        case 4:
            return "Fri"
        case 5:
            return "Sat"
        case 6:
            return "Sun"
