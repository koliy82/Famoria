
<div align="center">

# Famoria
**is an open-source telegram bot with the function of creating marriages and earning in-game currency.**

![bots](resources/images/leonardo.jpg)

<br>
</div>

## Commands
- `/help` or `/start` - Get help.
- `/menu` - Open bot menu.
- `/gobrak` - Invite someone to create family. (Use as a reply to a user's message)
- `/endbrak` - End the marriage.
- `/profile` - Get your profile with family (if exist).
- `/braks` - Get a list of families in current chat.
- `/braksglobal` - Get a list of global families.
- `/gokids` - Give birth to a child. (Use as a reply to a user's message)
- `/detdom` - Disown of marriage.
- `/kidannihilate` - Annihilate a child.
- `/tree` - Get a family tree. (Without number - text format, with number from 1 to 5 - picture format)
- `/deposit` or `/dep` - Transfer in-game currency from the user balance to the family balance. (Use `/dep 1`)
- `/withdraw` or `/with` - Transfer in-game currency from your family's balance to your own balance. (Use `/with 1`)
- `/inventory` - Reviewing purchased family items.
- `/shop` - Purchase items from the secret shop to enhance the rewards of in-game events
- `/subscribe` - Take out a family subscription.
- `/settings` - Chat settings (administrators only): language, video converter, marriage and gacha features.

# Self-hosting
### Clone the Repository:
```bash
git clone https://github.com/koliy82/famoria.git
```
### Set Up Environment Variables:
Create a .env file in the root directory and add your Telegram bot token. 
- TELEGRAM_TOKEN - Your Telegram bot token. [(How to get a token)](https://core.telegram.org/bots#6-botfather)
- APP_ENV [Optional] - Application environment. (default: dev)
- AppTimeZone [Optional] - Application timezone. (default: Europe/Moscow)
- CLICKHOUSE_URL - ClickHouse URL.
- CLICKHOUSE_PORT - ClickHouse port.
- CLICKHOUSE_USER - ClickHouse user.
- CLICKHOUSE_PASSWORD - ClickHouse password.
- CLICKHOUSE_DATABASE - ClickHouse database.
- MONGO_URI - MongoDB URI.
- MONGO_DATABASE - MongoDB database.
- ERRORS_CHAT_ID [Optional] - Telegram Chat ID for error messages.

### Proxies for blocked networks:
Where Telegram or the database host is blocked at the network level, either side
can be routed through an HTTP(S) CONNECT proxy. The two are independent, and both
are off by default.

- BOT_PROXY [Optional] - `true` to route all Telegram Bot API traffic through a proxy. (default: false)
- BOT_PROXY_URL - The proxy, as `login:pass@host:port`. Required when BOT_PROXY is true.
- DB_PROXY [Optional] - `true` to route MongoDB traffic through a proxy. (default: false)
- DB_PROXY_URL - The proxy, as `login:pass@host:port`. Required when DB_PROXY is true.

```bash
BOT_PROXY=true
BOT_PROXY_URL=login:pass@host:port
DB_PROXY=true
DB_PROXY_URL=login:pass@host:port
```

Notes:
- The scheme is optional. A bare `login:pass@host:port` is treated as `http://`,
  which is what proxy providers normally serve. Prefix `https://` only when the
  proxy itself speaks TLS.
- Special characters in the password must be percent-encoded (`@` becomes `%40`),
  because the value is parsed as a URL.
- Enabling a flag without a usable URL fails at startup. That is deliberate: the
  alternative is a silent direct connection, which on a blocked network looks
  like a hang rather than a misconfiguration.
- Keys may be written in lower or upper case; both are accepted.
- `PROXY_ENABLE` / `PROXY_URL` are separate and apply to yt-dlp (video
  downloads) only. See below.

### Video link converter:
When enabled per chat via `/settings`, the bot turns a posted video link
(YouTube, Shorts, TikTok, Reels) into the video itself, and can delete the
original link message.

- YTDLP_COOKIES_FILE [Optional] - Path to a Netscape-format cookies file. YouTube
  blocks unauthenticated access from datacenter IPs, so this is effectively
  required there.
- PROXY_ENABLE [Optional] - When yt-dlp uses `PROXY_URL`: `false` (default),
  `youtube`, or `true` for all links.
- PROXY_URL [Optional] - The proxy for yt-dlp, as `http://login:pass@host:port`.

## Set Up the Database:
- Famoria uses ClickHouse and MongoDB. You can use Docker to run them.
- Generate replica.key for MongoDB. [(How to generate a key)](https://www.mongodb.com/docs/manual/tutorial/enforce-keyfile-access-control-in-existing-replica-set/)
```bash
cd database-compose
cd clickhouse
docker-compose up -d
cd ../mongo
docker compose build
docker compose up --wait
```

- Connect to MongoDB with the mongo shell and run the following command to initiate the replica set:
```
try { rs.status() } catch (err) { rs.initiate({_id:'rs0',members:[{_id:0,host:'host.docker.internal:27017'}]}) }
```

### Set Up the Bot in root directory:
```bash
docker compose up -d
```

# Contact
If you have any questions or suggestions, feel free to open an issue or contact [Koliy82]([Koliy82](https://t.me/koliy822)).
