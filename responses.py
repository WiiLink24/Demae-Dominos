from lxml import etree
from werkzeug import exceptions
from models import User
from helpers import (
    response,
    dict_to_etree,
    multiple_root_nodes,
    get_restaurant,
    get_all_address_data
)
from areas import *
import json
from dominos import domino
from dominos.country import country, store_country
import uuid
import textwrap
from datetime import datetime
from food import db


@response()
def inquiry_done(_):
    """The request a restaurant part.
    In the forms, it give us the telephone of the restaurant,
    name of the restaurant, and restaurant type.
    TODO: append to database"""

    return {}


@multiple_root_nodes()
def validate_condition(_):
    return {}


@store_country()
@response()
def order_done(request):
    shop_code = request.form["shop[ShopCode]"]
    first_name = request.form["member[Name1]"]
    last_name = request.form["member[Name2]"]
    phone_number = request.form["member[TelNo]"]

    basket: User = User.query.filter_by(
        mac_address=request.headers["X-WiiMAC"][:12]
    ).first()

    city = ""
    state = ""
    street = ""
    street_number = ""

    #geocoder = GoogleV3(api_key="")
    #location = geocoder.geocode(request.form.get("member[Address5]"))

    #for data in location.raw["address_components"]:
    #    if data["types"] == ["street_number"]:
     #       street_number = data["long_name"]
      #  elif data["types"] == ["route"]:
       #     street = data["long_name"]
        #elif data["types"] == ["locality", "political"]:
         #   city = data["long_name"]
        #elif data["types"] == ["administrative_area_level_1", "political"]:
         #   state = data["long_name"]

    # data = domino.place_order("address", "city", "state", "zip", "House", "street",
    #     "num", shop_code, basket.basket, "name", "last name", "email",
    #      "phone number", "None", basket.order_id, basket.price)

    #    print(json.dumps(data))

    basket.basket = []
    db.session.commit()

    # We require this specific format.
    # With "orderDay", this handles the year [0:4], month [4:6], and day [6:8].
    # Meanwhile, the channel only parses "hour" at [8:10], and minute at [10:12].
    current_time = datetime.utcnow().strftime("%Y%m%d%H%S")

    return {
        "Message": {"contents": "Thank you! Your order has been placed."},
        "order_id": 17,
        "orderDay": current_time,
        "hashKey": "Testing: 1, 2, 3",
        "hour": current_time,
    }


@multiple_root_nodes()
def basket_delete(request):
    num = request.args.get("basketNo")
    area_code = request.args.get("areaCode")

    basket = User.query.filter_by(area_code=area_code).first()

    order: list = basket.basket

    del order[int(num) - 1]

    basket.basket = []
    db.session.commit()
    idk = basket.basket + order
    basket.basket = idk
    db.session.commit()

    return {}


@multiple_root_nodes()
def inquiry_done(_):
    return {}


@multiple_root_nodes()
def basket_reset(request):
    basket: User = User.query.filter_by(
        mac_address=request.headers["X-WiiMAC"][:12]
    ).first()
    basket.basket = []

    db.session.commit()

    return {}


@store_country()
@multiple_root_nodes()
def basket_list(request):
    tk2 = request.args.get("areaCode")
    shopcode = request.args.get("shopCode")
    address = request.headers.get("X-Address")
    postal = request.headers.get("X-Postalcode")
    bask = None
    count = 0
    cart = {
        "basketPrice": "",
        "chargePrice": "",
        "totalPrice": "",
        "Status": {"isOrder": 1, "messages": {"hey": "how are you?"}},
        "List": {},
    }

    queried_data = User.query.filter_by(area_code=tk2).first()

    location = get_all_address_data(address, postal)

    city = ""
    state = ""
    street = ""
    street_number = ""

    for data in location.raw["address_components"]:
        if data["types"] == ["street_number"]:
            street_number = data["long_name"]
        elif data["types"] == ["route"]:
            street = data["long_name"]
        elif data["types"] == ["locality", "political"]:
            city = data["long_name"]
        elif data["types"] == ["administrative_area_level_1", "political"]:
            state = data["long_name"]

    x = json.loads(
        domino.get_price(
            address,
            city,
            state,
            postal,
            "House",
            street,
            street_number,
            shopcode,
            queried_data.basket,
        )
    )

    cart["basketPrice"] = str(x["Order"]["Amounts"]["Menu"])
    cart["chargePrice"] = str(x["Order"]["Amounts"]["Tax"])
    cart["totalPrice"] = str(x["Order"]["Amounts"]["Customer"])
    for food in x["Order"]["Products"]:
        count = count + 1
        conta = "container{}".format(str(count))
        container = {
            "basketNo": count,
            # maybe we need a way to get the menucode here
            "menuCode": 1,
            "itemCode": food["Code"],
            "name": food["Name"],
            "price": str(food["Price"]),
            "size": "L",
            "isSoldout": 0,
            "quantity": food["Qty"],
            "subTotalPrice": str(food["Amount"]),
            "Menu": {
                "name": "Menu",
                "lunchMenuList": {
                    "isLunchTimeMenu": 1,
                    "isOpen": 1,
                },
            },
            "optionList": {
                "testing": {
                    "info": "idk",
                    "code": 1,
                    "type": 1,
                    "name": "domino",
                    "list": {
                        "item_one": {
                            "name": "Item One",
                            "menuCode": 1,
                            "itemCode": 1,
                            "image": 1,
                            "isSoldout": 0,
                            "info": "idk",
                            "price": "5.99",
                        }
                    },
                }
            },
        }
        cart["List"].update({conta: container})
    queried_data.order_id = x["Order"]["OrderID"]
    queried_data.price = str(x["Order"]["Amounts"]["Customer"])
    db.session.commit()
    return cart


@multiple_root_nodes()
def auth_key(request):
    queried_data = User.query.filter_by(
        mac_address=request.headers["X-WiiMac"][:12]
    ).first()
    queried_data.auth_key = str(uuid.uuid1())
    db.session.commit()
    return {"authKey": str(uuid.uuid1())}


@store_country()
@multiple_root_nodes()
def item_one(request):
    name = request.args.get("itemCode")
    shopcode = request.args.get("shopCode")
    x = json.loads(domino.get_food_price(shopcode, name))

    return {
        "price": x["food"]["price"],
        "optionList": {
            "testing": {
                "info": "Pick your crust",
                "code": "1",
                "type": "1",
                "name": x["food"]["name"],
                "list": {
                    "item_one": {
                        "name": "Thin",
                        "menuCode": 1,
                        "itemCode": 1,
                        "image": "non",
                        "isSoldout": 0,
                        "info": "Who wants it?",
                        "price": "5.99",
                    },
                    "item_two": {
                        "name": "Gluten-Free",
                        "menuCode": 1,
                        "itemCode": 1,
                        "image": "non",
                        "isSoldout": 0,
                        "info": "Who wants it?",
                        "price": "5.99",
                    },
                },
            },
            "testing2": {
                "info": "Pick your crust",
                "code": "0",
                "type": "-1",
                "name": "Please Help",
                "list": {
                    "item_one": {
                        "name": "Thin",
                        "menuCode": 1,
                        "itemCode": 1,
                        "image": "non",
                        "isSoldout": 0,
                        "info": "Who wants it?",
                        "price": "5.99",
                    },
                    "item_two": {
                        "name": "Gluten-Free",
                        "menuCode": 1,
                        "itemCode": 1,
                        "image": "non",
                        "isSoldout": 0,
                        "info": "Who wants it?",
                        "price": "5.99",
                    },
                },
            },
        },
    }


@store_country()
@multiple_root_nodes()
def item_list(request):
    menucode = int(request.args.get("menuCode")) + 1
    shopcode = request.args.get("shopCode")
    x = json.loads(domino.get_menu(shopcode))
    fooding = {
        "Count": 0,
        "List": {},
    }
    count = 0
    count2 = 0
    count3 = 0
    itemcode = 0
    ourcat = None
    for cat in x["categories"]:
        count = count + 1
        # playing safe xd
        if str(count) == str(menucode):
            ourcat = cat

    for types in x["categories"][ourcat]:
        count = count + 1
        if count <= 28:
            # Word Wrap out description
            wrapper = textwrap.TextWrapper(width=39)
            description = ""
            i = 40
            idk = x["categories"][ourcat][types]["desc"]

            while i % 40 == 0:
                if i > len(idk):
                    break

                text_array = wrapper.wrap(idk)
                for i, text in enumerate(text_array):
                    if i == 0:
                        description += text
                    elif i == 2:
                        description += "\n"
                        description += text[:22]
                    elif i >= 3:
                        break
                    else:
                        description += "\n"
                        description += text

                i += 40

            fooding["Count"] = fooding["Count"] + 1
            conta = "container{}".format(str(count2))

            container = {
                "name": x["categories"][ourcat][types]["name"],
                "item": {
                    "menuCode": f"{menucode}",
                    "itemCode": types,
                    "price": "666",
                    "info": description,
                    "size": "bizza",
                    "image": types,
                    "isSoldout": 0,
                    "sizeList": {},
                },
            }
            count2 = count2 + 1
            for items in x["categories"][ourcat][types]["items"]:
                iteming = "item{}".format(str(count3))
                sizelist = {
                    "size": items["name"],
                    "itemCode": items["code"],
                    "isSoldout": 0,
                    "price": items["price"],
                }
                count3 = count3 + 1
                container["item"]["sizeList"].update({iteming: sizelist})
            fooding["List"].update({conta: container})

    return fooding


@store_country()
@multiple_root_nodes()
def menu_list(request):
    shop_id = request.args.get("shopCode")
    x = json.loads(domino.get_menu(shop_id))
    menus = {"response": {}}

    for i, cat in enumerate(x["categories"]):
        conta = f"testing_{i}"
        container = {
            "menuCode": i,
            "linkTitle": cat,
            "enabledLink": "1",
            "name": "Yeah! Good food.",
            "info": "Screamingly delightful.",
            "setNum": 1,
            "lunchMenuList": {
                "isLunchTimeMenu": 0,
                "hour": {
                    "start": "00:00:00",
                    "end": "24:59:59",
                },
                "isOpen": 1,
                "message": "Where does this show up?",
            },
        }
        menus["response"].update({conta: container})

    menus["response"].update(
        {
            "z": {
                "menuCode": 11,
                "linkTitle": "Placeholder",
                "enabledLink": 1,
                "name": "Yeah! Good food.",
                "info": "Screamingly delightful.",
                "setNum": 0,
                "lunchMenuList": {
                    "isLunchTimeMenu": 1,
                    "hour": {
                        "start": 1,
                        "end": 1,
                    },
                    "isOpen": 1,
                },
                "message": "Where does this show up?",
            }
        }
    )

    return menus


@store_country()
@multiple_root_nodes()
def shop_one(request):
    activate = "on"

    shop_id = request.args.get("shopCode")
    x = json.loads(domino.get_recommended(shop_id))
    info = domino.get_store_info(shop_id)

    menucode = 0
    menus = {"response": {}}

    if not info["is_open"]:
        activate = "off"

    delivery_start_time = f'{info["service_hours"]["OpenTime"]}:00'
    delivery_end_time = f'{info["service_hours"]["CloseTime"]}:00'

    menus["response"].update(
        {
            "categoryCode": "02",
            "address": info["address"],
            "information": "idk",
            "attention": "why.",
            "amenity": "Dominos Pizza",
            "menuListCode": 1,
            "activate": activate,
            "waitTime": info["wait_time"],
            "timeorder": "y",
            "tel": info["phone"],
            "yoyakuMinDate": 1,
            "yoyakuMaxDate": 30,
            "paymentList": {"athing": "Fox Card"},
            "shopStatus": {
                "hours": {
                    "all": {
                        "message": info["detailed_wait"],
                    },
                    "today": {
                        "values": {
                            "start": delivery_start_time,
                            "end": delivery_end_time,
                            "holiday": "n",
                        }
                    },
                    "delivery": {
                        "values": {
                            "start": delivery_start_time,
                            "end": delivery_end_time,
                            "holiday": "n",
                        }
                    },
                    "status": {
                        "isOpen": info["is_open"],
                    },
                    "selList": {
                        "values": {
                            "one": {"id": 1, "name": "Wii"},
                            "two": {"id": 2, "name": "Link"},
                        }
                    },
                },
                "interval": 5,
                "holiday": "No ordering on Canada Day.",
            },
        }
    )
    menus["response"].update({"recommendItemList": {}})

    for cat in x["recommended"]:
        menucode = menucode + 1
        conta = "container{}".format(str(menucode))

        # Format our name
        name = ""
        product = x["recommended"][cat]["name"]
        i = 15
        wrapper = textwrap.TextWrapper(width=15)

        while i % 15 == 0:
            if i > len(x["recommended"][cat]["name"]):
                print(name)
                break

            text_array = wrapper.wrap(product)
            for i, text in enumerate(text_array):
                if i == 0:
                    name += text
                else:
                    name += "\n"
                    name += text

            i += 15

        container = {
            "menuCode": 10,
            "itemCode": 1,
            # Wrap text
            "name": name,
            "price": 1,
            "info": "Freshly charred",
            "size": 1,
            "image": x["recommended"][cat]["code"],
            "isSoldout": 0,
            "sizeList": {
                "item1": {"itemCode": 1, "size": 1, "price": 1, "isSoldout": 0}
            },
        }

        menus["response"]["recommendItemList"].update({conta: container})

    return menus


@response()
def shop_info(request):
    # Return a blank dict for now
    return {}


@response()
def shop_list(request):
    return category_list(request)


@response()
def document_template(request):
    # Observed to be true in v1 and v512.
    # Actually not, the version is different when making an order request
    if request.args.get("version") != "00000":
        # Dump request dataD
        print(request.args)
        print(request.json)

    return {
        "container0": {
            "contents": "Welcome to the stream! This will be the first pizza   ever ordered on the WiiLink24 Demae "
            "Channel!!!!!!"
        },
        "container1": {"contents": "Congratulations on throwing money into the void"},
        "container2": {"contents": "the heck are you doing"},
    }


@response()
def area_list(request):
    if request.args.get("zipCode"):
        # Nintendo, for whatever reason, require a separate "selectedArea" element
        # as a root node within output.
        # This violates about every XML specification in existence.
        # I am reasonably certain there was a mistake as their function to
        # interpret nodes at levels accepts a parent node, to which they seem to
        # have passed NULL instead of response.
        #
        # We are not going to bother spending time to deal with this.
        @response()
        def area_list_only_segments():
            return {
                "areaList": {
                    "place": {
                        "segment": "segment title",
                        "list": {
                            "areaPlace": {"areaName": "place name", "areaCode": 2}
                        },
                    },
                },
                "areaCount": 1,
            }

        area_list_output = area_list_only_segments()

        selected_area = dict_to_etree("selectedArea", {"areaCode": 1})
        selected_area_output = etree.tostring(selected_area, pretty_print=True)

        return area_list_output + selected_area_output

    area_code = request.args.get("areaCode")
    if not area_code:
        # We expect either a zip code or an area code.
        return exceptions.BadRequest()

    if area_code == "0":
        # An area code of 0 is passed upon first search.
        return {
            "areaList": {
                "place": {
                    "segment": "United States",
                    "list": {"areaPlace": get_american_states()},
                },
                "place2": {
                    "segment": "Canada",
                    "list": {"areaPlace": get_canadian_provinces()},
                },
            },
            "areaCount": 2,
        }

    if area_code != "0":
        # An area code of 0 is passed upon first search. All else is deterministic.
        zip_code = generate_random(10, area_code)

        user = User(
            area_code=zip_code, basket=[], mac_address=request.headers["X-WiiMac"][:12]
        )

        db.session.add(user)
        db.session.commit()

        cities, state_name, count = get_cities_by_state(area_code, zip_code)

        return {
            "areaList": {
                "place": {
                    "container0": "aaaa",
                    "segment": state_name,
                    "list": {"areaPlace": cities},
                },
            },
            "areaCount": count,
        }

    return exceptions.NotFound()


def formulate_restaurant(request, category_id: int, area_code) -> dict:
    return {
        "LargeCategoryName": "Meal",
        "CategoryList": {
            "TestingCategory": {
                "CategoryCode": f"{category_id:02}",
                "ShopList": {"Shop": get_restaurant(request, area_code)},
            }
        },
    }


@store_country()
@multiple_root_nodes()
def category_list(request):
    # TODO: formulate properly
    print(country)
    area_code = request.args.get("areaCode")
    queried_data = User.query.filter_by(area_code=area_code).first()

    return {
        "response": {
            "Pizza": formulate_restaurant(request, 1, area_code),
            "Placeholder": formulate_restaurant(request, 2, area_code),
        }
    }


@store_country()
@multiple_root_nodes()
def basket_add(request):
    auth = request.form.get("areaCode")
    shopcode = request.form.get("shopCode")
    itemcode = request.form.get("itemCode")
    qty = request.form.get("quantity")

    item = domino.add_food_item(shopcode, itemcode, qty, "1")
    query = User.query.filter_by(area_code=auth).first()

    food = query.basket
    idk = food + [item]

    query.basket = idk
    db.session.commit()

    return {"e": True}


action_list = {
    "webApi_document_template": document_template,
    "webApi_area_list": area_list,
    "webApi_category_list": category_list,
    "webApi_area_shopinfo": shop_info,
    "webApi_shop_list": shop_list,
    "webApi_shop_one": shop_one,
    "webApi_menu_list": menu_list,
    "webApi_item_list": item_list,
    "webApi_item_one": item_one,
    "webApi_Authkey": auth_key,
    "webApi_basket_list": basket_list,
    "webApi_basket_reset": basket_reset,
    "webApi_basket_add": basket_add,
    "webApi_validate_condition": validate_condition,
    "webApi_order_done": order_done,
    "webApi_basket_delete": basket_delete,
    "webApi_inquiry_done": inquiry_done,
}
