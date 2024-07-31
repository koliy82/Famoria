<-- GENERATE replica.key -->
https://www.mongodb.com/docs/manual/tutorial/enforce-keyfile-access-control-in-existing-replica-set/

<-- BUILD INFO -->
docker compose build
docker compose up --wait

<-- CONNECT TO DATABASE WITH MONGOSH AND PASTE -->
try { rs.status() } catch (err) { rs.initiate({_id:'rs0',members:[{_id:0,host:'host.docker.internal:27017'}]}) }

