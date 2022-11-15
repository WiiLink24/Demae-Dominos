import requests
import json
import datetime

from dominos.country import country, Country


api_urls = {
    Country.USA: "https://order.dominos.com",
    Country.CANADA: "https://order.dominos.ca",
}

image_urls = {
    Country.USA: "https://cache.dominos.com/olo/6_92_1/assets/build/market/US/_en/images/img/products/larges",
    Country.CANADA: "https://cache.dominos.com/nolo/ca/en/6_88_3/assets/build/market/CA/_en/images/img/products/larges",
}


def get_store_info(store_id):
    jsondata = {}

    response = requests.get(
        f"{api_urls[country]}/power/store/{store_id}/profile",
        verify=True,
        headers=get_headers(),
    )

    s1 = response.json()
    wait_time = s1["ServiceMethodEstimatedWaitMinutes"]["Delivery"]["Max"]
    min_price = s1["MinimumDeliveryOrderAmount"]
    address = s1["AddressDescription"]
    detailed_wait = s1["EstimatedWaitMinutes"]
    phone = s1["Phone"]
    is_open = s1["IsOpen"]
    service_hours = s1["ServiceHours"]["Delivery"][convert_datetime_today_to_date()]

    # Some Dominos stores may close the next day, specifically on weekends. We will only handle ordering from the
    # current day.
    if len(service_hours) == 2:
        service_hours = service_hours[1]
    else:
        service_hours = service_hours[0]

    jsondata.update(
        {
            "wait_time": wait_time,
            "min_price": min_price,
            "address": address,
            "detailed_wait": detailed_wait,
            "phone": phone,
            "is_open": is_open,
            "service_hours": service_hours,
        }
    )
    return jsondata


def address_lookup(zipcode, address):
    jsondata = {"user": {}, "restaurants": []}
    response = requests.get(
        f"{api_urls[country]}/power/store-locator?type=Delivery&c={zipcode}&s={address}",
        verify=True,
        headers=get_headers(),
    )
    s1 = response.json()
    print(response.json())
    street = s1["Address"]["Street"]
    streetnum = s1["Address"]["StreetNumber"]
    streetname = s1["Address"]["StreetName"]
    city = s1["Address"]["City"]
    region = s1["Address"]["Region"]
    postal = s1["Address"]["PostalCode"]

    count = 0
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
    for store in s1["Stores"]:
        count = count + 1
        if count <= 5:
            if not store["IsOpen"]:
                jsondata["restaurants"].append(
                    {
                        "id": store["StoreID"],
                        "address": store["AddressDescription"].replace("\n", " "),
                        "time": store["ServiceHoursDescription"]["Delivery"].replace(
                            "\n", " "
                        ),
                        "isOpen": False,
                        "isDelivery": store["IsDeliveryStore"],
                    }
                )
            else:
                jsondata["restaurants"].append(
                    {
                        "id": store["StoreID"],
                        "address": store["AddressDescription"].replace("\n", " "),
                        "time": store["ServiceHoursDescription"]["Delivery"].replace(
                            "\n", " "
                        ),
                        "isOpen": True,
                        "isDelivery": store["IsDeliveryStore"],
                    }
                )
    x = json.dumps(jsondata)
    return x


def get_menu(store_id):
    food = {"categories": {}}
    response = requests.get(
        f"{api_urls[country]}/power/store/{store_id}/menu?lang=en&structured=true",
        verify=True,
        headers=get_headers(),
    )
    s2 = response.json()
    for cat in s2["Categorization"]["Food"]["Categories"]:
        food["categories"].update({cat["Code"]: {}})
        for prods in s2["Products"]:
            if s2["Products"][prods]["ProductType"] == cat["Code"]:
                food["categories"][cat["Code"]].update(
                    {
                        prods: {
                            "name": s2["Products"][prods]["Name"],
                            "desc": s2["Products"][prods]["Description"],
                            "items": [],
                        }
                    }
                )

    for cat2 in s2["Categorization"]["Food"]["Categories"]:
        for cazzo in s2["Products"]:
            if cat2["Code"] == s2["Products"][cazzo]["ProductType"]:
                for item in s2["Variants"]:
                    if s2["Variants"][item]["ProductCode"] == cazzo:
                        food["categories"][cat2["Code"]][cazzo]["items"].append(
                            {
                                "code": item,
                                "name": s2["Variants"][item]["Name"],
                                "price": s2["Variants"][item]["Price"],
                                "img": s2["Variants"][item]["ProductCode"],
                                "size": s2["Variants"][item]["SizeCode"],
                            }
                        )
    x = json.dumps(food)
    return x


def get_recommended(store_id):
    count = 0
    food = {"recommended": {}}
    response = requests.get(
        f"{api_urls[country]}/power/store/{store_id}/menu?lang=en&structured=true",
        verify=True,
        headers=get_headers(),
    )
    s2 = response.json()
    for cat in s2["PreconfiguredProducts"]:
        count = count + 1
        if count <= 4:
            food["recommended"].update({cat: {}})
            food["recommended"][cat].update(
                {
                    "code": cat,
                    "name": s2["PreconfiguredProducts"][cat]["Name"],
                    "code": s2["PreconfiguredProducts"][cat]["Code"],
                }
            )

    x = json.dumps(food)
    return x


def get_food_price(store_id, item_id):
    food = {"food": {}}
    response = requests.get(
        f"{api_urls[country]}/power/store/{store_id}/menu?lang=en&structured=true",
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
        f"{api_urls[country]}/power/store/{store_id}/menu?lang=en&structured=true",
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
            "SourceOrganizationURI": api_urls[country],
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
        "Origin": f"{api_urls[country]}",
        "Referer": f"{api_urls[country]}/assets/build/xdomain/proxy.html",
    }
    response = requests.post(
        f"{api_urls[country]}/power/validate-order",
        verify=True,
        headers=headers2,
        data=json.dumps(bigorder),
    )
    orderiddi = json.loads(response.text)["Order"]["OrderID"]
    bigorder["Order"]["OrderID"] = orderiddi
    response = requests.post(
        f"{api_urls[country]}/power/validate-order",
        verify=True,
        headers=headers2,
        data=json.dumps(bigorder),
    )
    orderiddi = json.loads(response.text)["Order"]["OrderID"]
    bigorder["Order"]["OrderID"] = orderiddi
    bigorder["Order"].update({"metaData": {"orderFunnel": "payments"}})
    response = requests.post(
        f"{api_urls[country]}/power/price-order",
        verify=True,
        headers=headers2,
        data=json.dumps(bigorder),
    )
    return response.text


def place_order(
    street,
    apartment_number,
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
        "Origin": f"{api_urls[country]}",
        "Referer": f"{api_urls[country]}/assets/build/xdomain/proxy.html",
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
            "SourceOrganizationURI": "order.dominos.com",
            "StoreID": store_id,
            "Tags": {},
            "Version": "1.0",
            "NoCombine": True,
            "Partners": {},
            "HotspotsLite": False,
            "OrderInfoCollection": [],
        }
    }

    if apartment_number:
        bigorder["Order"]["Address"].update(
            {"AddressLine2": apartment_number}
        )

    bigorder["Order"]["Address"].update(
        {"DeliveryInstructions": "None"}
    )
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
        f"{api_urls[country]}/power/place-order",
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
