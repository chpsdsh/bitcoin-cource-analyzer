#!/bin/sh
set -eu

if [ "$#" -lt 2 ]; then
  echo "usage: $0 <plan.jmx> <output-dir> [jmeter args...]" >&2
  exit 1
fi

PLAN_PATH="$1"
OUTPUT_DIR="$2"
shift 2

PROJECT_NAME="${COMPOSE_PROJECT_NAME:-bitcoin-cource-analyzer}"
NETWORK_NAME="${COMPOSE_NETWORK_NAME:-${PROJECT_NAME}_default}"
JMETER_IMAGE="${JMETER_IMAGE:-justb4/jmeter:5.6.3}"

mkdir -p "$OUTPUT_DIR"
rm -rf "$OUTPUT_DIR/report"

docker run --rm \
  --network "$NETWORK_NAME" \
  -v "$PWD:/work" \
  -w /work \
  "$JMETER_IMAGE" \
  -n \
  -t "$PLAN_PATH" \
  -l "$OUTPUT_DIR/results.jtl" \
  -j "$OUTPUT_DIR/jmeter.log" \
  -e \
  -o "$OUTPUT_DIR/report" \
  -Jjmeter.save.saveservice.output_format=csv \
  -Jjmeter.save.saveservice.default_delimiter='|' \
  -Jjmeter.save.saveservice.print_field_names=true \
  -Jjmeter.save.saveservice.assertion_results=none \
  -Jjmeter.save.saveservice.bytes=true \
  -Jjmeter.save.saveservice.label=true \
  -Jjmeter.save.saveservice.latency=true \
  -Jjmeter.save.saveservice.response_code=true \
  -Jjmeter.save.saveservice.response_message=true \
  -Jjmeter.save.saveservice.successful=true \
  -Jjmeter.save.saveservice.thread_counts=true \
  -Jjmeter.save.saveservice.thread_name=true \
  -Jjmeter.save.saveservice.time=true \
  -Jjmeter.save.saveservice.timestamp_format=ms \
  "$@"
