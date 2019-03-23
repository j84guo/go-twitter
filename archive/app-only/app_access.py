#!/usr/bin/env python3

"""
This script illustrates OAuth's client credentials grant. The client is also
the resource owner, and uses a client key + secret to obtain an access token.
"""

import sys
import json
import yaml
import base64

from urllib.parse import quote, urlencode
from http.client import HTTPSConnection

# client credentials
from config import c_key, c_secret

# url encode
ue_c_key = quote(c_key, safe="")
ue_c_secret = quote(c_secret, safe="")

# base64 encode
c_cred_str = "{}:{}".format(ue_c_key, ue_c_secret)
b64_c_cred_bytes = base64.b64encode(bytes(c_cred_str, "ascii"))
b64_c_cred_str = str(b64_c_cred_bytes, "ascii")

# request access token
def get_access_token():
    conn = HTTPSConnection("api.twitter.com")
    headers = {
        "Authorization":"Basic {}".format(b64_c_cred_str),
        "Content-Type":"application/x-www-form-urlencoded;charset=UTF-8"
    }
    body = {
        "grant_type":"client_credentials"
    }
    conn.request("POST", "/oauth2/token", urlencode(body), headers)
    resp = conn.getresponse()
    resp_bytes = resp.read()
    resp_str = str(resp_bytes, "ascii")

    # extract token
    resp_body = json.loads(resp_str)
    if resp_body["token_type"] != "bearer":
        print("Error extracting access token from response...\n")
        print(resp_str)
    access_token = resp_body["access_token"]
    print("Obtained access token: {}".format(access_token))

    # write to configuration file
    config = {"access_token":access_token}
    with open("app_access.yml", "w") as f:
        yaml.dump(config, f, default_flow_style=False)

if __name__ == "__main__":
    get_access_token()
