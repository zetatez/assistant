
curl -s -X GET http://127.0.0.1:8080/health | jq

curl -s -X GET http://127.0.0.1:8080/user/list | jq

curl -s -X GET http://127.0.0.1:8080/user/get/1 | jq

curl -s -X DELETE http://127.0.0.1:8080/user/delete/1 | jq

curl -s -X PUT http://127.0.0.1:8080/user/create \
  -H 'Content-Type: application/json' \
  -d '{"name":"Alice 2","email":"alice2@example.com"}' | jq

