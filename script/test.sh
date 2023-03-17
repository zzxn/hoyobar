#!/bin/sh

# Smoke test for login/register/create post/list the first page of posts

curl  -d '{"username": "18702123685", "password": "password", "vcode":"0000", "nickname": "zzxn"}' localhost:8080/api/user/register
curl  -d '{"username": "18702123685", "password": "password"}' localhost:8080/api/user/login

curl  -d '{"title": "t001", "content": "asd", "author_id": "111"}' localhost:8080/api/post/create
curl  -d '{"title": "t002", "content": "asd", "author_id": "111"}' localhost:8080/api/post/create
curl  -d '{"title": "t003", "content": "asd", "author_id": "111"}' localhost:8080/api/post/create
curl  -d '{"title": "t004", "content": "asd", "author_id": "111"}' localhost:8080/api/post/create
curl  -d '{"title": "t005", "content": "asd", "author_id": "111"}' localhost:8080/api/post/create
curl  -d '{"title": "t006", "content": "asd", "author_id": "111"}' localhost:8080/api/post/create
curl  -d '{"title": "t007", "content": "asd", "author_id": "111"}' localhost:8080/api/post/create
curl  -d '{"title": "t008", "content": "asd", "author_id": "111"}' localhost:8080/api/post/create
curl  -d '{"title": "t009", "content": "asd", "author_id": "111"}' localhost:8080/api/post/create
curl  -d '{"title": "t010", "content": "asd", "author_id": "111"}' localhost:8080/api/post/create
curl  -d '{"title": "t011", "content": "asd", "author_id": "111"}' localhost:8080/api/post/create
curl  -d '{"title": "t012", "content": "asd", "author_id": "111"}' localhost:8080/api/post/create
curl  -d '{"title": "t013", "content": "asd", "author_id": "111"}' localhost:8080/api/post/create
curl  -d '{"title": "t014", "content": "asd", "author_id": "111"}' localhost:8080/api/post/create
curl  -d '{"title": "t015", "content": "asd", "author_id": "111"}' localhost:8080/api/post/create
curl  -d '{"title": "t016", "content": "asd", "author_id": "111"}' localhost:8080/api/post/create
curl  -d '{"title": "t017", "content": "asd", "author_id": "111"}' localhost:8080/api/post/create
curl  -d '{"title": "t018", "content": "asd", "author_id": "111"}' localhost:8080/api/post/create
curl  -d '{"title": "t019", "content": "asd", "author_id": "111"}' localhost:8080/api/post/create
curl  -d '{"title": "t020", "content": "asd", "author_id": "111"}' localhost:8080/api/post/create

curl  "localhost:8080/api/post/list?order=create_time"
