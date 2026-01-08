#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage:
  cf-dlp-decode.sh [--try-json] <input.log.gz>

What it does:
  1) Decompresses <input.log.gz> -> <input.log.json>
  2) Pretty-prints <input.log.json> (jq)
  3) Decodes .Payload based on:
       - .Headers.content-type starts with "application/json" => base64 -> JSON
       - .Headers.content-encoding equals "gzip"              => base64 -> gzip -> JSON (or other format)
       - .Headers.content-type starts with "text/plain"       => base64 -> plain text
  4) Optionally: with --try-json, attempts JSON decode even when content-type is not supported

Outputs:
  <input.log.json>    (pretty-printed)
  <input.json>        (decoded payload; name may differ if non-json)
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
try_json=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --try-json)
      try_json=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    --)
      shift
      break
      ;;
    -*)
      die "Unknown option: $1"
      ;;
    *)
      break
      ;;
  esac
done

[[ $# -eq 1 ]] || { usage; exit 2; }

in="$1"
[[ -f "$in" ]] || die "Input file not found: $in"
[[ "$in" == *.log.gz ]] || die "Input must end with .log.gz (got: $in)"

# ---- derive filenames ----
base="${in%.log.gz}"
log_json="${base}.log.json"
payload_out="${base}.json"
payload_txt="${base}.txt"
payload_b64_tmp=""
payload_bin_tmp=""
payload_unzipped_tmp=""
payload_decode_tmp=""

# ---- step 1: decompress log ----
gzip -dc -- "$in" > "$log_json"

# ---- step 3: format log json (in-place via temp) ----
tmp="$(mktemp)"
jq '.' -- "$log_json" > "$tmp" && mv "$tmp" "$log_json"

# ---- read headers ----
content_type="$(jq -r '.Headers["content-type"] // empty' -- "$log_json" 2>/dev/null || true)"
content_encoding="$(jq -r '.Headers["content-encoding"] // empty' -- "$log_json" 2>/dev/null || true)"

decode_base64_payload() {
  local strict="${1:-1}"

  fail() {
    if [[ "$strict" -eq 1 ]]; then
      die "$*"
    fi
    rm -f "$payload_b64_tmp" "$payload_bin_tmp" "$payload_unzipped_tmp"
    return 1
  }

  # Extract base64 payload (as text) -> decode to bytes.
  # Use a temp file because payload might be large.
  payload_b64_tmp="$(mktemp)"
  payload_bin_tmp="$(mktemp)"

  jq -r '.Payload // empty' -- "$log_json" > "$payload_b64_tmp"
  [[ -s "$payload_b64_tmp" ]] || fail "No .Payload found (or it was empty) in $log_json"

  if ! base64 -d < "$payload_b64_tmp" > "$payload_bin_tmp" 2>/dev/null; then
    fail "base64 decode failed (is .Payload valid base64?)"
  fi

  return 0
}

maybe_gunzip_payload() {
  local strict="${1:-1}"
  local in_path="$2"

  fail() {
    if [[ "$strict" -eq 1 ]]; then
      die "$*"
    fi
    rm -f "$payload_b64_tmp" "$payload_bin_tmp" "$payload_unzipped_tmp"
    return 1
  }

  if [[ "$content_encoding" == gzip ]]; then
    payload_unzipped_tmp="$(mktemp)"
    if ! gzip -dc < "$in_path" > "$payload_unzipped_tmp"; then
      fail "gzip decode failed (is payload really gzipped?)"
    fi
    payload_decode_tmp="$payload_unzipped_tmp"
  else
    payload_decode_tmp="$in_path"
  fi

  return 0
}

decode_json_payload() {
  local strict="${1:-1}"

  fail() {
    if [[ "$strict" -eq 1 ]]; then
      die "$*"
    fi
    rm -f "$payload_b64_tmp" "$payload_bin_tmp" "$payload_unzipped_tmp"
    return 1
  }

  if ! decode_base64_payload "$strict"; then
    return 1
  fi
  if ! maybe_gunzip_payload "$strict" "$payload_bin_tmp"; then
    return 1
  fi

  if ! jq '.' < "$payload_decode_tmp" > "$payload_out"; then
    fail "json decode failed (decoded payload wasn't valid JSON?)"
  fi

  rm -f "$payload_b64_tmp" "$payload_bin_tmp" "$payload_unzipped_tmp"
  echo "Wrote:"
  echo "  $log_json"
  echo "  $payload_out"
}

decode_text_payload() {
  local strict="${1:-1}"

  fail() {
    if [[ "$strict" -eq 1 ]]; then
      die "$*"
    fi
    rm -f "$payload_b64_tmp" "$payload_bin_tmp" "$payload_unzipped_tmp"
    return 1
  }

  if ! decode_base64_payload "$strict"; then
    return 1
  fi
  if ! maybe_gunzip_payload "$strict" "$payload_bin_tmp"; then
    return 1
  fi

  if ! cat "$payload_decode_tmp" > "$payload_txt"; then
    fail "text write failed"
  fi

  rm -f "$payload_b64_tmp" "$payload_bin_tmp" "$payload_unzipped_tmp"
  echo "Wrote:"
  echo "  $log_json"
  echo "  $payload_txt"
}

# ---- step 4-8: decode payload with options ----
# We only decode if it's JSON-ish by content-type. If not, we still *can* dump raw decoded bytes,
# but the request focuses on JSON payloads, so we gate on application/json unless --try-json is set.
if [[ "$content_type" == application/json* ]]; then
  decode_json_payload 1
elif [[ "$content_type" == text/plain* ]]; then
  decode_text_payload 1
elif [[ "$try_json" -eq 1 ]]; then
  if ! decode_json_payload 0; then
    echo "Wrote:"
    echo "  $log_json"
    echo "Tried payload decode (--try-json) but it failed; leaving payload undecoded." >&2
  fi
else
  echo "Wrote:"
  echo "  $log_json"
  echo "Skipped payload decode: .Headers.content-type is not supported (got: ${content_type:-<empty>})" >&2
  if [[ -t 0 ]]; then
    reply=""
    if read -r -p "Try JSON decode anyway? [y/N] " reply; then
      case "$reply" in
        y|Y|yes|YES)
          if ! decode_json_payload 0; then
            echo "Tried payload decode (prompt) but it failed; leaving payload undecoded." >&2
          fi
          ;;
      esac
    fi
  fi
fi
