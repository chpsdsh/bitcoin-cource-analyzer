#!/bin/sh
set -eu

if [ "$#" -ne 5 ]; then
  echo "usage: $0 <results.jtl> <min-samples> <max-error-pct> <max-p95-ms> <max-avg-ms>" >&2
  exit 1
fi

RESULTS_FILE="$1"
MIN_SAMPLES="$2"
MAX_ERROR_PCT="$3"
MAX_P95_MS="$4"
MAX_AVG_MS="$5"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

awk -F'|' 'NR > 1 { print $2 > elapsed; count++; if ($8 != "true") errors++; sum += $2 } END {
  if (count == 0) {
    printf("no samples found\n") > "/dev/stderr"
    exit 2
  }
  printf("%d\n", count) > count_file
  printf("%d\n", errors + 0) > error_file
  printf("%.6f\n", (errors + 0) * 100 / count) > error_pct_file
  printf("%.6f\n", sum / count) > avg_file
}' \
elapsed="$TMP_DIR/elapsed.txt" \
count_file="$TMP_DIR/count.txt" \
error_file="$TMP_DIR/errors.txt" \
error_pct_file="$TMP_DIR/error_pct.txt" \
avg_file="$TMP_DIR/avg.txt" \
"$RESULTS_FILE"

COUNT="$(cat "$TMP_DIR/count.txt")"
ERRORS="$(cat "$TMP_DIR/errors.txt")"
ERROR_PCT="$(cat "$TMP_DIR/error_pct.txt")"
AVG_MS="$(cat "$TMP_DIR/avg.txt")"

sort -n "$TMP_DIR/elapsed.txt" > "$TMP_DIR/elapsed-sorted.txt"
P95_INDEX="$(awk -v count="$COUNT" 'BEGIN {
  idx = int((count * 95 + 99) / 100)
  if (idx < 1) idx = 1
  print idx
}')"
P95_MS="$(awk -v idx="$P95_INDEX" 'NR == idx { print $1; exit }' "$TMP_DIR/elapsed-sorted.txt")"

echo "samples=$COUNT errors=$ERRORS error_pct=$ERROR_PCT avg_ms=$AVG_MS p95_ms=$P95_MS"

if [ "$COUNT" -lt "$MIN_SAMPLES" ]; then
  echo "expected at least $MIN_SAMPLES samples, got $COUNT" >&2
  exit 1
fi

awk -v actual="$ERROR_PCT" -v max="$MAX_ERROR_PCT" 'BEGIN { exit !(actual <= max) }' || {
  echo "error rate $ERROR_PCT exceeded max $MAX_ERROR_PCT" >&2
  exit 1
}

awk -v actual="$P95_MS" -v max="$MAX_P95_MS" 'BEGIN { exit !(actual <= max) }' || {
  echo "p95 $P95_MS exceeded max $MAX_P95_MS" >&2
  exit 1
}

awk -v actual="$AVG_MS" -v max="$MAX_AVG_MS" 'BEGIN { exit !(actual <= max) }' || {
  echo "avg $AVG_MS exceeded max $MAX_AVG_MS" >&2
  exit 1
}
