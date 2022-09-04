from models import db
import config
import ssl
import colorama
import os

from flask import request, Flask, send_from_directory
from flask_migrate import Migrate
from werkzeug import exceptions
from werkzeug.serving import WSGIRequestHandler
from debug import request_dump
from werkzeug.datastructures import ImmutableMultiDict

colorama.init()

app = Flask(__name__)
app.config["SQLALCHEMY_DATABASE_URI"] = config.db_url
app.config["SQLALCHEMY_TRACK_MODIFICATIONS"] = False

db.init_app(app)

# Ensure the DB is able to determine migration needs.
migrate = Migrate(app, db, compare_type=True)

import responses
from dominos.images import download_item_image
from dominos.country import country, store_country


@app.route("/nwapi.php", methods=["GET"])
def base_api():
    request_dump(request)
    try:
        # These values should be consistent for both v1 and v512.
        if request.args["platform"] != "wii":
            return exceptions.BadRequest()

        action = request.args["action"]
        return responses.action_list[action](request)
    except KeyError:
        # This is not an action or a format we know of.
        return exceptions.NotFound()


def print_multi(passed_dict):
    if isinstance(passed_dict, ImmutableMultiDict):
        passed_dict = passed_dict.items(multi=True)

    for key, value in passed_dict:
        try:
            # Encode as UTF-8
            value = value.encode("shift-jis").decode("utf-8")
            print(f"{key} -> {value}")
        except Exception as e:
            # If it errors, leave as is with a note.
            print(f"An error occurred while decoding key {e}")
            print(f"Its value is '{value}' (not decoded)")


@app.post("/nwapi.php")
def error_api():
    request_dump(request)
    if request.form.get("action") is None:
        print("Received an error!")

        print_multi(request.args)
        print_multi(request.form)

        return responses.action_list["webApi_document_template"](request)

    try:
        # These values should be consistent for both v1 and v512.
        if request.form["platform"] != "wii":
            return exceptions.BadRequest()

        action = request.form["action"]
        return responses.action_list[action](request)
    except KeyError:
        # This is not an action or a format we know of.
        return exceptions.NotFound()


# Image Routes
@app.route("/logoimg2/<filename>")
def serve_logo(filename):
    return send_from_directory("./images", filename)


@app.route("/itemimg/<_>/<filename>")
@store_country()
def serve_item_image(_, filename):
    if os.path.exists(f"./images/{str(country.value)}/{filename}"):
        return send_from_directory(f"./images/{str(country.value)}", filename)
    else:
        download_item_image(filename)
        return send_from_directory(f"./images/{str(country.value)}", filename)


if __name__ == "__main__":
    WSGIRequestHandler.protocol_version = "HTTP/1.1"
    context = ssl.SSLContext(ssl.PROTOCOL_TLSv1)
    context.set_ciphers("ALL:@SECLEVEL=0")
    context.load_cert_chain("server.pem", "server.key")
    app.run(host="::", port=443, debug=True, ssl_context=context)
