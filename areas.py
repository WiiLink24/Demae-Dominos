import secrets
import string
import geonamescache

from helpers import RepeatedElement


canadian_provinces = {
    "01": "Alberta",
    "02": "British Colombia",
    "03": "Manitoba",
    "04": "New Brunswick",
    "05": "Newfoundland and Labrador",
    "07": "Nova Scotia",
    "08": "Ontario",
    "09": "Prince Edward Island",
    "10": "Quebec",
    "11": "Saskatchewan",
    "12": "Yukon",
    "13": "Northwest Territories",
}


def get_american_states() -> list:
    gc = geonamescache.GeonamesCache()
    results = []

    for state in gc.get_us_states_by_names().values():
        results.append(
            RepeatedElement({"areaName": state["name"], "areaCode": state["code"]})
        )

    return results


def get_canadian_provinces() -> list:
    gc = geonamescache.GeonamesCache()
    results = []

    for provinces in canadian_provinces:
        results.append(
            RepeatedElement(
                {"areaName": canadian_provinces[provinces], "areaCode": provinces}
            )
        )

    return results


def get_cities_by_state(state_code: str, zip_code: str) -> (list, str, int):
    gc = geonamescache.GeonamesCache()
    results = []
    count = 0
    state_name = ""

    try:
        # First see if it is Canadian
        state_name = canadian_provinces[state_code]
    except KeyError:
        # This is American
        state_name = gc.get_us_states()[state_code]["name"]

    for city in gc.get_cities().values():
        if city["countrycode"] == "CA" or city["countrycode"] == "US":
            if city["admin1code"] == state_code:
                if city["population"] > 50000:
                    results.append(
                        RepeatedElement(
                            {
                                "areaName": city["name"],
                                "areaCode": zip_code,
                                "isNextArea": 0,
                                "display": 1,
                                "kanji1": state_name,
                                "kanji2": city["name"],
                                "kanji3": "",
                                "kanji4": "",
                            }
                        )
                    )
                    count += 1

    print(count)
    return results, state_name, count


def generate_random(length: int, area_code: str):
    # We will use this function to generate an area code for the user.
    gc = geonamescache.GeonamesCache()
    code = ""
    letters = string.digits
    random = "".join(secrets.choice(letters) for i in range(length))

    try:
        # First see if it is Canadian
        _ = canadian_provinces[area_code]
        code = "1"
    except KeyError:
        # This is American
        code = "0"

    return code + random
