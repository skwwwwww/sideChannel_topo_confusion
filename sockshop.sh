#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# --- Configuration ---
# Base URL for the application
BASE_URL="http://192.168.88.151"

# Initial md.sid cookie value provided in the first request
# This cookie is typically set by the server on the initial page load (e.g., index.html)
INITIAL_MD_SID="md.sid=s%3AnVzRAoOajN5_s0z6154EQwWYh-0oItzd.O%2Foo8w9qtbuqcMRuISbS2C8g1DE4MhBx3Jo6nbNLkIc"

# Common User-Agent string to mimic a browser
USER_AGENT="Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36"

# Customer session ID extracted from the 'logged_in' cookie value from the /login response
CUSTOMER_SESSION_ID="nVzRAoOajN5_s0z6154EQwWYh-0oItzd"

# Specific item ID used for browsing and adding to cart
ITEM_ID="a0a4f044-b040-410d-8ead-4de0446aec7e"

# Cookie jar file to persist session cookies across requests
COOKIE_JAR="cookiejar.txt"

# Delay between major operations (in seconds)
OPERATION_DELAY=0.5
# Delay between full iterations (in seconds) - IMPORTANT for load control
ITERATION_DELAY=0.5

echo "Starting HTTP request automation script for $BASE_URL..."
echo "--------------------------------------------------"

# Clean up previous cookie jar to start with a fresh session
rm -f "$COOKIE_JAR"
echo "Cleaned up old cookie jar: $COOKIE_JAR"

# --- Perform initial login outside the loop to establish session ---
echo "--- Performing initial Login (GET /login) ---"
# This request sets the 'logged_in' cookie, which is crucial for subsequent authenticated requests.
curl -s -o /dev/null \
  -b "$INITIAL_MD_SID" \
  -c "$COOKIE_JAR" \
  -H "Accept: */*" \
  -H "Accept-Language: zh-CN,zh;q=0.9" \
  -H "Authorization: Basic MTox" \
  -H "Connection: keep-alive" \
  -H "Host: 192.168.88.151" \
  -H "Referer: $BASE_URL/index.html" \
  -A "$USER_AGENT" \
  -H "X-Requested-With: XMLHttpRequest" \
  --compressed \
  "$BASE_URL/login"
echo "Login request sent. Cookies (including logged_in) saved to $COOKIE_JAR."
sleep $OPERATION_DELAY # Simulate user pause

# --- Main loop for repeated user actions ---
ITERATION_COUNT=0
while true; do
  ITERATION_COUNT=$((ITERATION_COUNT + 1))
  echo ""
  echo "##################################################"
  echo "### Starting new automation iteration #$ITERATION_COUNT ###"
  echo "##################################################"
  echo ""

  # --- 1. Fetching common HTML elements and initial data (after login) ---
  echo "--- Fetching topbar.html ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: text/html, */*; q=0.01" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/index.html" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/topbar.html"
  sleep $OPERATION_DELAY

  echo "--- Fetching navbar.html ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: text/html, */*; q=0.01" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/index.html" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/navbar.html"
  sleep $OPERATION_DELAY

  echo "--- Fetching footer.html ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: text/html, */*; q=0.01" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/index.html" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/footer.html"
  sleep $OPERATION_DELAY

  echo "--- Fetching initial catalogue (size=5) data ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: application/json, text/javascript, */*; q=0.01" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/index.html" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/catalogue?size=5"
  sleep $OPERATION_DELAY

  echo "--- Fetching cart initial state ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: application/json, text/javascript, */*; q=0.01" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/index.html" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/cart"
  sleep 1

  # --- 2. 浏览商品 (Browse Product Detail) ---
  echo "--- Navigating to detail.html for item $ITEM_ID ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/index.html" \
    -H "Upgrade-Insecure-Requests: 1" \
    -A "$USER_AGENT" \
    --compressed \
    "$BASE_URL/detail.html?id=$ITEM_ID"
  sleep 1

  echo "--- Fetching catalogue item $ITEM_ID data (product details) ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: application/json, text/javascript, */*; q=0.01" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/detail.html?id=$ITEM_ID" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/catalogue/$ITEM_ID"
  sleep $OPERATION_DELAY

  echo "--- Fetching cart state after viewing detail page ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: application/json, text/javascript, */*; q=0.01" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/detail.html?id=$ITEM_ID" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/cart"
  sleep $OPERATION_DELAY

  echo "--- Fetching customer data for session $CUSTOMER_SESSION_ID ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: */*" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/detail.html?id=$ITEM_ID" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/customers/$CUSTOMER_SESSION_ID"
  sleep $OPERATION_DELAY

  echo "--- Fetching related catalogue items (tags=blue) ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: application/json, text/javascript, */*; q=0.01" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/detail.html?id=$ITEM_ID" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/catalogue?sort=id&size=3&tags=blue"
  sleep 1

  # --- 3. 添加商品到购物车 (Add item to cart) ---
  CART_ADD_PAYLOAD="{\"id\": \"$ITEM_ID\"}"
  echo "--- Adding item $ITEM_ID to cart (POST /cart) ---"
  curl -s -o /dev/null \
    -X POST \
    -b "$COOKIE_JAR" \
    -H "Accept: */*" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Content-Type: application/json; charset=UTF-8" \
    -H "Content-Length: ${#CART_ADD_PAYLOAD}" \
    -H "Host: 192.168.88.151" \
    -H "Origin: $BASE_URL" \
    -H "Referer: $BASE_URL/detail.html?id=$ITEM_ID" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --data "$CART_ADD_PAYLOAD" \
    --compressed \
    "$BASE_URL/cart"
  echo "Item added to cart."
  sleep 1

  # --- Requests to refresh page content after adding to cart (likely 304 Not Modified) ---
  echo "--- Refreshing detail.html page after adding to cart ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Cache-Control: max-age=0" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "If-Modified-Since: Tue, 21 Mar 2017 11:31:47 GMT" \
    -H "If-None-Match: W/\"287c-15af0a320b8\"" \
    -H "Referer: $BASE_URL/detail.html?id=$ITEM_ID" \
    -H "Upgrade-Insecure-Requests: 1" \
    -A "$USER_AGENT" \
    --compressed \
    "$BASE_URL/detail.html?id=$ITEM_ID"
  sleep $OPERATION_DELAY

  echo "--- Fetching catalogue item $ITEM_ID data (refresh) ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: application/json, text/javascript, */*; q=0.01" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/detail.html?id=$ITEM_ID" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/catalogue/$ITEM_ID"
  sleep $OPERATION_DELAY

  echo "--- Fetching cart state (refresh) ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: application/json, text/javascript, */*; q=0.01" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/detail.html?id=$ITEM_ID" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/cart"
  sleep $OPERATION_DELAY

  echo "--- Fetching customer data for session $CUSTOMER_SESSION_ID (refresh) ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: */*" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/detail.html?id=$ITEM_ID" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/customers/$CUSTOMER_SESSION_ID"
  sleep $OPERATION_DELAY

  echo "--- Fetching related catalogue items (refresh) ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: application/json, text/javascript, */*; q=0.01" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/detail.html?id=$ITEM_ID" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/catalogue?sort=id&size=3&tags=blue"
  sleep $OPERATION_DELAY

  echo "--- Fetching catalogue item $ITEM_ID data (redundant refresh from log) ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: application/json, text/javascript, */*; q=0.01" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/detail.html?id=$ITEM_ID" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/catalogue/$ITEM_ID"
  sleep 1

  # --- 4. 查看购物车 (View Cart) ---
  echo "--- Navigating to basket.html ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/detail.html?id=$ITEM_ID" \
    -H "Upgrade-Insecure-Requests: 1" \
    -A "$USER_AGENT" \
    --compressed \
    "$BASE_URL/basket.html"
  sleep 1

  # --- Data fetching requests on basket.html page ---
  echo "--- Fetching cart data in basket.html ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: application/json, text/javascript, */*; q=0.01" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/basket.html" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/cart"
  sleep $OPERATION_DELAY

  echo "--- Fetching card data in basket.html ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: application/json, text/javascript, */*; q=0.01" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/basket.html" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/card"
  sleep $OPERATION_DELAY

  echo "--- Fetching address data in basket.html ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: application/json, text/javascript, */*; q=0.01" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/basket.html" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/address"
  sleep $OPERATION_DELAY

  echo "--- Fetching catalogue (size=3) in basket.html ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: application/json, text/javascript, */*; q=0.01" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/basket.html" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/catalogue?size=3"
  sleep $OPERATION_DELAY

  echo "--- Fetching cart data (second time in basket.html) ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: application/json, text/javascript, */*; q=0.01" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/basket.html" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/cart"
  sleep $OPERATION_DELAY

  echo "--- Fetching customer data (second time in basket.html) ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: */*" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/basket.html" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/customers/$CUSTOMER_SESSION_ID"
  sleep $OPERATION_DELAY

  echo "--- Fetching catalogue item $ITEM_ID data (again in basket.html) ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: application/json, text/javascript, */*; q=0.01" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/basket.html" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/catalogue/$ITEM_ID"
  sleep $OPERATION_DELAY

  echo "--- Fetching catalogue item $ITEM_ID data (yet again in basket.html, redundant from log) ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: application/json, text/javascript, */*; q=0.01" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/basket.html" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/catalogue/$ITEM_ID"
  sleep 1

  # --- 5. 付款 (Payment) ---
  echo "--- Placing order (POST /orders) ---"
  # This request has an empty body, as indicated by Content-Length: 0.
  curl -s -o /dev/null \
    -X POST \
    -b "$COOKIE_JAR" \
    -H "Accept: */*" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Content-Length: 0" \
    -H "Host: 192.168.88.151" \
    -H "Origin: $BASE_URL" \
    -H "Referer: $BASE_URL/basket.html" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/orders"
  echo "Order placement request sent."
  sleep 1

  echo "--- Clearing cart after successful order (DELETE /cart) ---"
  curl -s -o /dev/null \
    -X DELETE \
    -b "$COOKIE_JAR" \
    -H "Accept: */*" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Origin: $BASE_URL" \
    -H "Referer: $BASE_URL/basket.html" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/cart"
  echo "Cart cleared."
  sleep 1

  # --- Post-payment actions: viewing order history ---
  echo "--- Navigating to customer-orders.html ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/basket.html" \
    -H "Upgrade-Insecure-Requests: 1" \
    -A "$USER_AGENT" \
    --compressed \
    "$BASE_URL/customer-orders.html?" # Trailing ? from original request
  echo "Redirected to order history page."
  sleep 1

  echo "--- Fetching orders data for display on customer-orders.html ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: application/json, text/javascript, */*; q=0.01" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/customer-orders.html?" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/orders"
  sleep $OPERATION_DELAY

  echo "--- Fetching customer data for session $CUSTOMER_SESSION_ID (final check) ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: */*" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/customer-orders.html?" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/customers/$CUSTOMER_SESSION_ID"
  sleep $OPERATION_DELAY

  echo "--- Fetching cart data (final check, should be empty) ---"
  curl -s -o /dev/null \
    -b "$COOKIE_JAR" \
    -H "Accept: application/json, text/javascript, */*; q=0.01" \
    -H "Accept-Language: zh-CN,zh;q=0.9" \
    -H "Connection: keep-alive" \
    -H "Host: 192.168.88.151" \
    -H "Referer: $BASE_URL/customer-orders.html?" \
    -A "$USER_AGENT" \
    -H "X-Requested-With: XMLHttpRequest" \
    --compressed \
    "$BASE_URL/cart"
  sleep $OPERATION_DELAY

  echo ""
  echo "--------------------------------------------------"
  echo "Iteration #$ITERATION_COUNT complete. Waiting for $ITERATION_DELAY seconds before next iteration..."
  echo "--------------------------------------------------"
  sleep $ITERATION_DELAY # IMPORTANT: Pause before starting the next full cycle
done
