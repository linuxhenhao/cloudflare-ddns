case "$1" in 
start):
  /usr/bin/cloudflare-ddns >/dev/null 2>&1
stop):
 kill -all cloudflare-ddns
restart):
 kill -all cloudflare-ddns
 /usr/bin/cloudflare-ddns >/dev/null 2>&1
