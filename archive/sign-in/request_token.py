#!/usr/bin/env python3

import re
import hmac
import json
import time
import uuid
import base64
import random

from hashlib import sha1
from urllib.parse import quote
from http.client import HTTPSConnection
from config import consumer_key, consumer_secret

def get_base_url(protocol, host, path):
    return "{}{}{}".format(protocol, host, path)

def get_urlencoded_callback():
    s = "http://127.0.0.1:80/callback"
    return quote(s, safe="")

def get_callback():
    return "http://127.0.0.1:80/callback"

def get_nonce():
    s = str(uuid.uuid4())
    r = re.sub("[^0-9A-Za-z]", "", s)
    return r

def get_timestamp():
    i = int(time.time())
    return str(i)

def get_signature_params(oauth_args, getting_request_token=True):
    params = list()
    exclude = excluded_signature_params(getting_request_token)
    for k in oauth_args:
        if k in exclude:
            continue
        url_k, url_v = quote(k, safe=""), quote(oauth_args[k], safe="")
        params.append((url_k, url_v))
    params.sort(key=lambda t: t[0])
    return params

def get_signature_params_string(params):
    s = ""
    for (k, v) in params:
        s += k + "=" + v + "&"
    s = s[0:-1]
    return s

def get_signature_base_string(verb, base_url, params_string):
    verb = verb.upper()
    base_url = quote(base_url, safe="")
    params_string = quote(params_string, safe="")
    base = "{}&{}&{}".format(verb, base_url, params_string)
    return base

def get_signature_key(consumer_secret, oauth_access_token_secret):
    key = quote(consumer_secret, "") + "&"
    if oauth_access_token_secret:
        key += quote(oauth_access_token_secret,"")
    return key

def get_b64_signature(base_string, signing_key):
    hash_bytes = hmac.new(bytes(signing_key, "UTF-8"), bytes(base_string, "UTF-8"), sha1).digest()
    b64_hash_bytes = base64.b64encode(hash_bytes)
    signature = str(b64_hash_bytes, "ascii")
    return signature

def excluded_signature_params(getting_request_token):
    exclude = ["oauth_signature"]
    if getting_request_token:
        exclude.append("oauth_token")
    return exclude

def excluded_header_params(getting_request_token):
    if getting_request_token:
        exclude = ["oauth_token"]
    else:
        exclude = ["oauth_callback"]
    return exclude

def build_oauth_header(oauth_args, getting_request_token):
    exclude = excluded_header_params(getting_request_token)
    header = "OAuth "
    for k in oauth_args:
        if k in exclude:
            continue
        header += quote(k, safe="") + "=\"" + quote(oauth_args[k], safe="") + "\", "
    header = header[0:-2]
    return header

VERB = "POST"
PROTOCOL = "https://"
HOST = "api.twitter.com"
PATH = "/oauth/request_token"
BASE_URL = get_base_url(PROTOCOL, HOST, PATH)

getting_request_token = True
oauth_access_token_secret = None

oauth_args = {
    "oauth_callback": get_callback(),
    "oauth_token": None,
    "oauth_signature": None,
    "oauth_consumer_key": consumer_key,
    "oauth_nonce": get_nonce(),
    "oauth_timestamp": get_timestamp(),
    "oauth_version": "1.0",
    "oauth_signature_method": "HMAC-SHA1"
}


if __name__ == "__main__":

    signature_params = get_signature_params(oauth_args, getting_request_token=getting_request_token)
    signature_params_string = get_signature_params_string(signature_params)
    print("generated signature parameter string...\n{}\n".format(signature_params_string))

    signature_base_string = get_signature_base_string(VERB, BASE_URL, signature_params_string)
    print("generated signature base string...\n{}\n".format(signature_base_string))

    signature_key = get_signature_key(consumer_secret, oauth_access_token_secret)
    print("generated signature key...\n{}\n".format(signature_key))

    signature = get_b64_signature(signature_base_string, signature_key)
    oauth_args["oauth_signature"] = signature
    print("generated oauth signature...\n{}\n".format(signature))

    header = build_oauth_header(oauth_args, getting_request_token)
    print("built oauth header...\n{}\n".format(header))

    conn = HTTPSConnection(HOST)
    headers = {
        "Authorization":header
    }

    conn.request(VERB, PATH, headers=headers)
    resp = conn.getresponse()
    resp_bytes = resp.read()
    resp_str = str(resp_bytes, "ascii")

    print(resp.status)
    print(resp.getheaders())
    print(resp_str)
