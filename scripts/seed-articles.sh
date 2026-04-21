#!/bin/sh
set -e

VALKEY_HOST="${VALKEY_HOST:-valkey}"
VALKEY_PORT="${VALKEY_PORT:-6379}"
VALKEY_DB="${VALKEY_DB:-0}"
VALKEY_NEWS_DB="${VALKEY_NEWS_DB:-1}"
VALKEY_LLM_RESPONSE_DB="${VALKEY_LLM_RESPONSE_DB:-2}"
VALKEY_PASSWORD="${VALKEY_PASSWORD:-}"

AUTH_ARGS=""
if [ -n "$VALKEY_PASSWORD" ]; then
  AUTH_ARGS="--no-auth-warning -a ${VALKEY_PASSWORD}"
fi

CLI="valkey-cli ${AUTH_ARGS} -h ${VALKEY_HOST} -p ${VALKEY_PORT} -n ${VALKEY_DB}"
NEWS_CLI="valkey-cli ${AUTH_ARGS} -h ${VALKEY_HOST} -p ${VALKEY_PORT} -n ${VALKEY_NEWS_DB}"
LLM_RESPONSE_CLI="valkey-cli ${AUTH_ARGS} -h ${VALKEY_HOST} -p ${VALKEY_PORT} -n ${VALKEY_LLM_RESPONSE_DB}"

echo "Waiting for Valkey at ${VALKEY_HOST}:${VALKEY_PORT}..."
until $CLI PING >/dev/null 2>&1; do
  sleep 1
done

echo "Valkey is up. Seeding articles into DB ${VALKEY_DB}, news into DB ${VALKEY_NEWS_DB}, and llm responses into DB ${VALKEY_LLM_RESPONSE_DB}..."

seed_article() {
  category="$1"
  score="$2"
  title="$3"
  text="$4"

  json=$(printf '{"category":"%s","title":"%s","text":"%s"}' \
    "$category" "$title" "$text")

  $CLI ZADD "$category" "$score" "$json" >/dev/null
}

seed_news() {
  category="$1"
  score="$2"
  title="$3"
  url="$4"
  social_image="$5"

  json=$(printf '{"category":"%s","title":"%s","url":"%s","socialimage":"%s"}' \
    "$category" "$title" "$url" "$social_image")

  $NEWS_CLI ZADD "$category" "$score" "$json" >/dev/null
}

seed_llm_response() {
  category="$1"
  summarization="$2"
  signal_direction="$3"
  signal_strength="$4"
  uncertainty="$5"
  event_urgency_hours="$6"
  numbers_density="$7"
  entity_density="$8"

  json=$(printf '{"category":"%s","summarization":"%s","features":{"signal_direction":"%s","signal_strength":%s,"uncertainty":%s,"event_urgency_hours":%s,"numbers_density":%s,"entity_density":%s}}' \
    "$category" "$summarization" "$signal_direction" "$signal_strength" "$uncertainty" "$event_urgency_hours" "$numbers_density" "$entity_density")

  $LLM_RESPONSE_CLI SET "$category" "$json" >/dev/null
}

base_ts=$(date +%s)

# politics
seed_article "politics"    "$((base_ts + 1))"  "US Congress discusses new budget deal" "US lawmakers are negotiating a new budget deal. Markets are watching political stability."
seed_article "politics"    "$((base_ts + 2))"  "European elections increase market uncertainty" "European elections increased uncertainty around regulation and fiscal policy."
seed_article "politics"    "$((base_ts + 3))"  "New sanctions package raises geopolitical tension" "A new sanctions package raised geopolitical tension and shifted market attention to digital assets."
seed_article "politics"    "$((base_ts + 4))"  "Government shutdown fears return to headlines" "Government shutdown fears returned and may raise market volatility."
seed_article "politics"    "$((base_ts + 5))"  "Central bank independence debated by officials" "Officials renewed debate over central bank independence and policy credibility."
seed_news "politics"       "$((base_ts + 1))"  "US Congress discusses new budget deal" "https://example.com/politics/us-congress-budget-deal" "https://images.example.com/politics/us-congress-budget-deal.jpg"
seed_news "politics"       "$((base_ts + 2))"  "European elections increase market uncertainty" "https://example.com/politics/european-elections-uncertainty" "https://images.example.com/politics/european-elections-uncertainty.jpg"
seed_news "politics"       "$((base_ts + 3))"  "New sanctions package raises geopolitical tension" "https://example.com/politics/sanctions-geopolitical-tension" "https://images.example.com/politics/sanctions-geopolitical-tension.jpg"
seed_news "politics"       "$((base_ts + 4))"  "Government shutdown fears return to headlines" "https://example.com/politics/government-shutdown-fears" "https://images.example.com/politics/government-shutdown-fears.jpg"
seed_news "politics"       "$((base_ts + 5))"  "Central bank independence debated by officials" "https://example.com/politics/central-bank-independence-debate" "https://images.example.com/politics/central-bank-independence-debate.jpg"
seed_llm_response "politics" "• Бюджетные переговоры в США усилили внимание к политической стабильности.\n• Европейские выборы повысили регуляторную неопределенность.\n• Санкции усилили геополитическое напряжение.\n• Опасения шатдауна могут повысить волатильность рынков.\n• Дискуссия о независимости центробанка влияет на доверие к политике.\nИтог: политический фон остается напряженным и умеренно негативным для риска." "down" "0.62" "0.34" "24" "0" "6"

# environment
seed_article "environment" "$((base_ts + 6))"  "Heatwave pressures power grids across regions" "A heatwave is pressuring power grids and raising energy costs for mining."
seed_article "environment" "$((base_ts + 7))"  "Flooding disrupts industrial infrastructure" "Flooding disrupted transport and industry, increasing uncertainty for risky assets."
seed_article "environment" "$((base_ts + 8))"  "Renewable energy expansion supports mining relocation" "New renewable projects are improving conditions for mining relocation."
seed_article "environment" "$((base_ts + 9))"  "Storm damage causes temporary data center outages" "Storm damage caused temporary data center outages and highlighted infrastructure risk."
seed_article "environment" "$((base_ts + 10))" "Energy regulators discuss emergency conservation measures" "Energy regulators discussed emergency conservation measures and possible higher power prices."
seed_news "environment"    "$((base_ts + 6))"  "Heatwave pressures power grids across regions" "https://example.com/environment/heatwave-power-grids" "https://images.example.com/environment/heatwave-power-grids.jpg"
seed_news "environment"    "$((base_ts + 7))"  "Flooding disrupts industrial infrastructure" "https://example.com/environment/flooding-industrial-infrastructure" "https://images.example.com/environment/flooding-industrial-infrastructure.jpg"
seed_news "environment"    "$((base_ts + 8))"  "Renewable energy expansion supports mining relocation" "https://example.com/environment/renewable-energy-mining-relocation" "https://images.example.com/environment/renewable-energy-mining-relocation.jpg"
seed_news "environment"    "$((base_ts + 9))"  "Storm damage causes temporary data center outages" "https://example.com/environment/storm-data-center-outages" "https://images.example.com/environment/storm-data-center-outages.jpg"
seed_news "environment"    "$((base_ts + 10))" "Energy regulators discuss emergency conservation measures" "https://example.com/environment/emergency-conservation-measures" "https://images.example.com/environment/emergency-conservation-measures.jpg"
seed_llm_response "environment" "• Жара повысила нагрузку на энергосети.\n• Наводнения нарушили транспорт и промышленную инфраструктуру.\n• Развитие ВИЭ поддерживает перенос майнинга.\n• Штормы вызвали временные сбои дата-центров.\n• Регуляторы обсуждают меры экономии энергии и рост цен.\nИтог: экологические и энергетические факторы создают смешанный, но в целом сдержанный фон." "neutral" "0.48" "0.41" "72" "0" "5"

# economy
seed_article "economy"     "$((base_ts + 11))" "Inflation data cools more than expected" "Inflation slowed more than expected, supporting hopes for easier policy."
seed_article "economy"     "$((base_ts + 12))" "Labor market remains resilient in new report" "The labor market remained resilient, which may delay rate cuts."
seed_article "economy"     "$((base_ts + 13))" "Manufacturing index rebounds after weak quarter" "A manufacturing index rebounded, pointing to stronger economic momentum."
seed_article "economy"     "$((base_ts + 14))" "Bond yields fall on dovish central bank signals" "Bond yields fell after dovish central bank comments."
seed_article "economy"     "$((base_ts + 15))" "Retail sales miss expectations in major economy" "Retail sales missed expectations, raising hopes for policy support."
seed_news "economy"        "$((base_ts + 11))" "Inflation data cools more than expected" "https://example.com/economy/inflation-cools" "https://images.example.com/economy/inflation-cools.jpg"
seed_news "economy"        "$((base_ts + 12))" "Labor market remains resilient in new report" "https://example.com/economy/labor-market-resilient" "https://images.example.com/economy/labor-market-resilient.jpg"
seed_news "economy"        "$((base_ts + 13))" "Manufacturing index rebounds after weak quarter" "https://example.com/economy/manufacturing-index-rebounds" "https://images.example.com/economy/manufacturing-index-rebounds.jpg"
seed_news "economy"        "$((base_ts + 14))" "Bond yields fall on dovish central bank signals" "https://example.com/economy/bond-yields-fall" "https://images.example.com/economy/bond-yields-fall.jpg"
seed_news "economy"        "$((base_ts + 15))" "Retail sales miss expectations in major economy" "https://example.com/economy/retail-sales-miss" "https://images.example.com/economy/retail-sales-miss.jpg"
seed_llm_response "economy" "• Инфляция замедлилась сильнее ожиданий.\n• Устойчивый рынок труда может отсрочить снижение ставок.\n• Производственный индекс восстановился.\n• Доходности облигаций снизились после мягких сигналов ЦБ.\n• Слабые розничные продажи усилили ожидания поддержки экономики.\nИтог: макрофон остается преимущественно позитивным для спроса на риск." "up" "0.66" "0.28" "24" "2" "5"

# technology
seed_article "technology"  "$((base_ts + 16))" "New blockchain scaling solution enters testing" "A new blockchain scaling solution entered testing with lower fees and faster settlement."
seed_article "technology"  "$((base_ts + 17))" "Major cloud provider expands Web3 tooling" "A major cloud provider expanded Web3 tooling for developers."
seed_article "technology"  "$((base_ts + 18))" "Security update reduces wallet vulnerability risks" "A wallet platform released a security update to reduce vulnerability risks."
seed_article "technology"  "$((base_ts + 19))" "Institutional custody platform adds new features" "A crypto custody platform added automation and compliance features."
seed_article "technology"  "$((base_ts + 20))" "Open-source developers publish node performance upgrade" "Open-source developers published a node performance upgrade."
seed_news "technology"     "$((base_ts + 16))" "New blockchain scaling solution enters testing" "https://example.com/technology/blockchain-scaling-testing" "https://images.example.com/technology/blockchain-scaling-testing.jpg"
seed_news "technology"     "$((base_ts + 17))" "Major cloud provider expands Web3 tooling" "https://example.com/technology/cloud-web3-tooling" "https://images.example.com/technology/cloud-web3-tooling.jpg"
seed_news "technology"     "$((base_ts + 18))" "Security update reduces wallet vulnerability risks" "https://example.com/technology/wallet-security-update" "https://images.example.com/technology/wallet-security-update.jpg"
seed_news "technology"     "$((base_ts + 19))" "Institutional custody platform adds new features" "https://example.com/technology/custody-platform-features" "https://images.example.com/technology/custody-platform-features.jpg"
seed_news "technology"     "$((base_ts + 20))" "Open-source developers publish node performance upgrade" "https://example.com/technology/node-performance-upgrade" "https://images.example.com/technology/node-performance-upgrade.jpg"
seed_llm_response "technology" "• Масштабирующее решение блокчейна вышло в тестирование.\n• Облачный провайдер расширил Web3-инструменты.\n• Обновление кошелька снизило риски уязвимостей.\n• Кастодиальная платформа добавила новые функции для институционалов.\n• Улучшение нод повысило эффективность сети.\nИтог: технологический фон выглядит позитивным для инфраструктуры крипторынка." "up" "0.71" "0.22" "72" "0" "7"

# crypto
seed_article "crypto"      "$((base_ts + 21))" "Bitcoin ETF inflows continue for third session" "Bitcoin ETF products saw another session of net inflows."
seed_article "crypto"      "$((base_ts + 22))" "Large holders accumulate during consolidation" "Large holders continued accumulating Bitcoin during consolidation."
seed_article "crypto"      "$((base_ts + 23))" "Exchange reserves decline as withdrawals increase" "Exchange reserves fell as Bitcoin withdrawals increased."
seed_article "crypto"      "$((base_ts + 24))" "Altcoin rally boosts sentiment across crypto market" "A rally in major altcoins improved sentiment across the crypto market."
seed_article "crypto"      "$((base_ts + 25))" "Mining difficulty reaches new high" "Mining difficulty reached a new high, showing strong network activity."
seed_news "crypto"         "$((base_ts + 21))" "Bitcoin ETF inflows continue for third session" "https://example.com/crypto/bitcoin-etf-inflows" "https://images.example.com/crypto/bitcoin-etf-inflows.jpg"
seed_news "crypto"         "$((base_ts + 22))" "Large holders accumulate during consolidation" "https://example.com/crypto/whales-accumulate-bitcoin" "https://images.example.com/crypto/whales-accumulate-bitcoin.jpg"
seed_news "crypto"         "$((base_ts + 23))" "Exchange reserves decline as withdrawals increase" "https://example.com/crypto/exchange-reserves-decline" "https://images.example.com/crypto/exchange-reserves-decline.jpg"
seed_news "crypto"         "$((base_ts + 24))" "Altcoin rally boosts sentiment across crypto market" "https://example.com/crypto/altcoin-rally-sentiment" "https://images.example.com/crypto/altcoin-rally-sentiment.jpg"
seed_news "crypto"         "$((base_ts + 25))" "Mining difficulty reaches new high" "https://example.com/crypto/mining-difficulty-high" "https://images.example.com/crypto/mining-difficulty-high.jpg"
seed_llm_response "crypto" "• Притоки в Bitcoin ETF продолжаются несколько сессий подряд.\n• Крупные держатели накапливают BTC в фазе консолидации.\n• Резервы на биржах снижаются, а вывод средств растет.\n• Рост альткоинов улучшил общий рыночный настрой.\n• Сложность майнинга обновила максимум.\nИтог: совокупный криптосигнал остается уверенно позитивным." "up" "0.82" "0.19" "24" "1" "6"

echo "Seed completed successfully."
echo "Inserted 5 articles into DB ${VALKEY_DB} and 5 news items into DB ${VALKEY_NEWS_DB} for each category: politics, environment, economy, technology, crypto."
echo "Inserted 1 llm response into DB ${VALKEY_LLM_RESPONSE_DB} for each category: politics, environment, economy, technology, crypto."
