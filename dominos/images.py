from dominos.domino import get_headers, image_urls
from dominos.country import country
from PIL import Image
import requests


def download_item_image(filename):
    response = requests.get(
        f"{image_urls[country]}/{filename}",
        headers=get_headers(),
    )

    if response.status_code != 200:
        return

    # We can't pass the bytes object directly to PIL because of a weird embedded null byte error
    f = open(f"./images/{str(country.value)}/{filename}", "wb")
    f.write(response.content)
    f.close()

    image = Image.open(f"./images/{str(country.value)}/{filename}")
    new_image = image.resize((160, 160))
    new_image.save(f"./images/{str(country.value)}/{filename}")
