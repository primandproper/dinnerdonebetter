-- Convert stored plaintext OAuth2 client secrets to their SHA-256 hex digests.
-- From this migration on, the application stores only the digest; the plaintext
-- secret is returned to the creator exactly once at creation time.
UPDATE oauth2_clients
SET client_secret = encode(sha256(convert_to(client_secret, 'UTF8')), 'hex');
