
<div align="center">

# Famoria
**is an open-source telegram bot with the function of creating marriages and earning in-game currency.**

![bots](resources/images/leonardo.jpg)

<br>
</div>

## Commands
- `/start` - Start the bot.
- `/help` - Get help.
- `/gobrak` - Invite someone to marry.
- `/endbrak` - End the marriage.
- `/profile` - Get your profile.
- `/braks` - Get a list of marriages in current chat.
- `/braksglobal` - Get a list of global marriages.
- `/gokids` - Give birth to a child.
- `/detdom` - Disown of marriage.
- `/kidannihilate` - Annihilate a child.
- `/tree` - Get a family tree. (in development...)

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