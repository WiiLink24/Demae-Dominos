import functools
import enum

from werkzeug.local import LocalProxy
from flask import request, g
from models import User


class Country(enum.Enum):
    USA = "UNITED_STATES"
    CANADA = "CANADA"


def get_current_country():
    if "country" not in g:
        # Default to the United States
        return None
    else:
        return g.country


country = LocalProxy(get_current_country)


def store_country():
    def decorator(func):
        @functools.wraps(func)
        def country_wrapper(*args, **kwargs):
            area_code = request.args.get("areaCode")

            if not area_code:
                # We always got the MAC Address.
                user: User = User.query.filter_by(
                    mac_address=request.headers["X-WiiMAC"][:12]
                ).first()
                area_code = user.area_code

            if area_code[0] == "0":
                g.country = Country.USA
            else:
                g.country = Country.CANADA

            return func(*args, **kwargs)

        return country_wrapper

    return decorator
