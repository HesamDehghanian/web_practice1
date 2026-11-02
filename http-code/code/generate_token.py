#!/usr/bin/env python3
"""Create a short-lived HS256 JWT for local testing only."""
import base64
import hashlib
import hmac
import json
import secrets
import sys
import time


def b64url(data: bytes) -> str:
    return base64.urlsafe_b64encode(data).rstrip(b"=").decode("ascii")


secret = sys.argv[1] if len(sys.argv) > 1 else "replace-this-with-a-long-random-secret-value"
if len(secret) < 32:
    raise SystemExit("Secret must contain at least 32 characters.")

now = int(time.time())
header = {"alg": "HS256", "typ": "JWT"}
payload = {
    "sub": "local-test-user",
    "iat": now,
    "nbf": now,
    "exp": now + 900,
    "jti": secrets.token_hex(8),
}
signing_input = (
    b64url(json.dumps(header, separators=(",", ":")).encode())
    + "."
    + b64url(json.dumps(payload, separators=(",", ":")).encode())
)
signature = hmac.new(secret.encode(), signing_input.encode(), hashlib.sha256).digest()
print(signing_input + "." + b64url(signature))
