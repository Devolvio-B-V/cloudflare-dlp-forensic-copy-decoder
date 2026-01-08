#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  cf-dlp-decode.sh <input.log.gz>

What it does:
  1) Decompresses <input.log.gz> -> <input.log.json>
  2) Pretty-prints <input.log.json> (jq)
  3) Decodes .Payload based on:
       - .Headers.content-type starts with "application/json" => base64 -> JSON
       - .Headers.content-encoding equals "gzip"              => base64 -> gzip -> JSON

Outputs:
  <input.log.json>    (pretty-printed)
  <input.json>        (decoded payload JSON; name may differ if non-json)
USAGE
}

die() { echo "ERROR: $*" >&2; exit 1; }

need() {
  command -v "$1" >/dev/null 2>&1 || die "Missing required dependecy: $1"
}

# ---- deps ----
need gzip
need jq
need base64

# ---- args ----
[[ $# -eq 1 ]] || { usage; exit 2; }

in="$1"
[[ -f "$in" ]] || die "Input file not found: $in"
[[ "$in" == *.log.gz ]] || die "Input must end with .log.gz (got: $in)"

# ---- derive filenames ----
base="${in%.log.gz}"
log_json="${base}.log.json"
payload_out="${base}.json"

# ---- step 1: decompress log ----
gzip -dc -- "$in" > "$log_json"

# ---- step 3: format log json (in-place via temp) ----
tmp="$(mktemp)"
jq '.' -- "$log_json" > "$tmp" && mv "$tmp" "$log_json"

# ---- read headers ----
content_type="$(jq -r '.Headers["content-type"] // empty' -- "$log_json" 2>/dev/null || true)"
content_encoding="$(jq -r '.Headers["content-encoding"] // empty' -- "$log_json" 2>/dev/null || true)"

# ---- step 4-8: decode payload with options ----
# We only decode if it's JSON-ish by content-type. If not, we still *can* dump raw decoded bytes,
# but the request focuses on JSON payloads, so we gate on application/json.
if [[ "$content_type" == application/json* ]]; then
  # Extract base64 payload (as text) -> decode to bytes
  # Use a temp file because payload might be large.
  payload_b64_tmp="$(mktemp)"
  payload_bin_tmp="$(mktemp)"

  jq -r '.Payload // empty' -- "$log_json" > "$payload_b64_tmp"
  [[ -s "$payload_b64_tmp" ]] || die "No .Payload found (or it was empty) in $log_json"

  # base64 decode
  if ! base64 -d < "$payload_b64_tmp" > "$payload_bin_tmp" 2>/dev/null; then
    die "base64 decode failed (is .Payload valid base64?)"
  fi

  # JSON + gzip decoding sequence if content-encoding is gzip
  if [[ "$content_encoding" == gzip ]]; then
    # decoded bytes are gzipped json
    if ! gzip -dc < "$payload_bin_tmp" | jq '.' > "$payload_out"; then
      die "gzip+json decode failed (is payload really gzipped JSON?)"
    fi
  else
    # JSON decoding sequence (decoded bytes are JSON)
    if ! jq '.' < "$payload_bin_tmp" > "$payload_out"; then
      die "json decode failed (decoded payload wasn't valid JSON?)"
    fi
  fi

  rm -f "$payload_b64_tmp" "$payload_bin_tmp"
  echo "Wrote:"
  echo "  $log_json"
  echo "  $payload_out"
else
  echo "Wrote:"
  echo "  $log_json"
  echo "Skipped payload decode: .Headers.content-type is not application/json (got: ${content_type:-<empty>})" >&2
fi
