#!/usr/bin/env python3

"""
Using the access token obtained in app_access.py, an HTTP request is made to an
endpoint on Twitter's api.
"""

import os
import yaml

from http.client import HTTPSConnection
from app_access import get_access_token

CONFIG_PATH = "app_access.yml"

if not os.path.exists(CONFIG_PATH):
    get_access_token()

with open(CONFIG_PATH) as f:
    config = yaml.load(f)
    access_token = config["access_token"]

print("Loaded access token: {}".format(access_token))

conn = HTTPSConnection("api.twitter.com")
headers = {
    "Authorization":"Bearer {}".format(access_token)
}

query_string = "?count=100&screen_name=twitterapi"
conn.request("GET", "/1.1/statuses/user_timeline.json{}".format(query_string), headers=headers)
resp = conn.getresponse()
resp_bytes = resp.read()
resp_str = str(resp_bytes, "ascii")
print(resp_str)
