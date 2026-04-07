#!/bin/sh
set -e

VALKEY_HOST="${VALKEY_HOST:-valkey}"
VALKEY_PORT="${VALKEY_PORT:-6379}"
VALKEY_DB="${VALKEY_DB:-0}"

CLI="valkey-cli -h ${VALKEY_HOST} -p ${VALKEY_PORT} -n ${VALKEY_DB}"

echo "Waiting for Valkey at ${VALKEY_HOST}:${VALKEY_PORT}..."
until $CLI PING >/dev/null 2>&1; do
  sleep 1
done

echo "Valkey is up. Seeding articles into DB ${VALKEY_DB}..."

seed_article() {
  category="$1"
  score="$2"
  title="$3"
  text="$4"

  json=$(printf '{"category":"%s","title":"%s","text":"%s"}' \
    "$category" "$title" "$text")

  $CLI ZADD "$category" "$score" "$json" >/dev/null
}

base_ts=$(date +%s)

# politics
seed_article "politics"    "$((base_ts + 1))"  "US Congress discusses new budget deal" "Lawmakers in the United States are negotiating a new budget agreement. Markets are watching political stability and fiscal policy signals that may affect investor sentiment toward risk assets including Bitcoin."
seed_article "politics"    "$((base_ts + 2))"  "European elections increase market uncertainty" "The latest election campaign in Europe has increased uncertainty around future regulation and fiscal priorities. Political risk often affects demand for alternative assets and speculative instruments."
seed_article "politics"    "$((base_ts + 3))"  "New sanctions package raises geopolitical tension" "A new international sanctions package has raised geopolitical tension. Traders are reassessing cross-border capital flows and safe-haven behavior, with some attention shifting toward digital assets."
seed_article "politics"    "$((base_ts + 4))"  "Government shutdown fears return to headlines" "Concerns over a possible government shutdown are back in the headlines. Short-term uncertainty may increase volatility in traditional markets and spill over into crypto markets."
seed_article "politics"    "$((base_ts + 5))"  "Central bank independence debated by officials" "Public debate around central bank independence has intensified after comments from senior officials. The discussion may influence expectations for monetary policy credibility and market confidence."

# environment
seed_article "environment" "$((base_ts + 6))"  "Heatwave pressures power grids across regions" "A prolonged heatwave is putting pressure on several regional power grids. Higher energy costs can affect mining profitability and create short-term pressure on Bitcoin-related infrastructure."
seed_article "environment" "$((base_ts + 7))"  "Flooding disrupts industrial infrastructure" "Severe flooding has disrupted transport and industrial infrastructure in multiple areas. Supply chain stress and elevated uncertainty may reduce appetite for risky assets in the near term."
seed_article "environment" "$((base_ts + 8))"  "Renewable energy expansion supports mining relocation" "The expansion of renewable energy projects is improving conditions for relocation of mining capacity. Better energy access can strengthen the long-term operational outlook for crypto mining."
seed_article "environment" "$((base_ts + 9))"  "Storm damage causes temporary data center outages" "Storm damage has caused temporary outages in several data centers. The event highlights infrastructure fragility and may increase risk perception in digital asset markets."
seed_article "environment" "$((base_ts + 10))" "Energy regulators discuss emergency conservation measures" "Energy regulators are discussing emergency conservation measures due to unstable seasonal demand. The possibility of higher electricity prices may weaken margins for energy-intensive blockchain operations."

# economy
seed_article "economy"     "$((base_ts + 11))" "Inflation data cools more than expected" "Fresh inflation data came in below analyst expectations, improving hopes for future monetary easing. Lower rate pressure tends to support risk appetite and may be positive for Bitcoin."
seed_article "economy"     "$((base_ts + 12))" "Labor market remains resilient in new report" "The latest labor market report shows continued resilience in employment and wages. Strong macro data may delay rate cuts, creating mixed implications for crypto markets."
seed_article "economy"     "$((base_ts + 13))" "Manufacturing index rebounds after weak quarter" "A key manufacturing index rebounded after a weak quarter, suggesting improving economic momentum. Stronger growth expectations can increase demand for higher-risk assets."
seed_article "economy"     "$((base_ts + 14))" "Bond yields fall on dovish central bank signals" "Government bond yields moved lower after dovish comments from central bank officials. Easier financial conditions are often supportive for speculative assets including Bitcoin."
seed_article "economy"     "$((base_ts + 15))" "Retail sales miss expectations in major economy" "Retail sales missed expectations in one of the largest global economies. Slower consumer activity may strengthen expectations for policy support and influence investor positioning."

# technology
seed_article "technology"  "$((base_ts + 16))" "New blockchain scaling solution enters testing" "A new blockchain scaling solution has entered public testing, promising lower fees and faster settlement. Infrastructure improvements can strengthen confidence in the broader crypto ecosystem."
seed_article "technology"  "$((base_ts + 17))" "Major cloud provider expands Web3 tooling" "A major cloud provider announced expanded tooling for Web3 developers. Easier enterprise integration may improve adoption prospects for blockchain-based products."
seed_article "technology"  "$((base_ts + 18))" "Security update reduces wallet vulnerability risks" "A widely used wallet platform released a security update addressing several vulnerability risks. Better security standards can support user trust in digital asset infrastructure."
seed_article "technology"  "$((base_ts + 19))" "Institutional custody platform adds new features" "An institutional crypto custody platform added automation and compliance features for large clients. Improved infrastructure may attract additional institutional participation."
seed_article "technology"  "$((base_ts + 20))" "Open-source developers publish node performance upgrade" "Open-source contributors published a major node performance upgrade. Better network efficiency may improve confidence in the long-term reliability of blockchain systems."

# crypto
seed_article "crypto"      "$((base_ts + 21))" "Bitcoin ETF inflows continue for third session" "Spot Bitcoin investment products recorded another session of net inflows. Sustained institutional demand is often interpreted as a bullish signal for the market."
seed_article "crypto"      "$((base_ts + 22))" "Large holders accumulate during consolidation" "On-chain data suggests that large holders continue accumulating Bitcoin during a period of price consolidation. This behavior is often associated with positive medium-term expectations."
seed_article "crypto"      "$((base_ts + 23))" "Exchange reserves decline as withdrawals increase" "Bitcoin reserves on centralized exchanges have declined while withdrawals increased. Lower available supply on trading venues may support price strength if demand remains stable."
seed_article "crypto"      "$((base_ts + 24))" "Altcoin rally boosts sentiment across crypto market" "A broad rally in major altcoins has improved overall market sentiment. Positive momentum across the sector often spills over into Bitcoin trading activity."
seed_article "crypto"      "$((base_ts + 25))" "Mining difficulty reaches new high" "Mining difficulty reached a new high, reflecting continued network participation and competition among miners. Strong network fundamentals can reinforce confidence in Bitcoin."

echo "Seed completed successfully."
echo "Inserted 5 articles into each category: politics, environment, economy, technology, crypto."