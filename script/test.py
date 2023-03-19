#!/bin/python3

# Basic test for register/login/create post/list post page by page

import grequests # this one must be imported fisrt!
import requests
import tqdm

PREFIX='http://localhost:8080/api'
CONCURRENT=100 # set it to 1 if using sqlite3
N_USER = 97
N_POST_PER_USER = 3
N_POST_REPLY_PER_USR = 2


def get(url: str, params, token=None):
    return requests.get(f'{PREFIX}{url}', params=params, headers={"Auth": token} if token else None)

def post(url: str, json, token=None):
    return requests.post(f'{PREFIX}{url}', json=json, headers={"Auth": token} if token else None)

def gpost(url: str, json, token=None):
    return grequests.post(f'{PREFIX}{url}', json=json, headers={"Auth": token} if token else None)

def assertOK(res: requests.Response):
    assert res.status_code == 200, "{} {} Got {}".format(res.request.method, res.request.url, res.json())

def registerUsers(n: int):
    username2user = {}
    regReqs = []
    for i in tqdm.trange(n, desc="prepare register"):
        # register
        username = f"187{i:08d}"
        nickname = f"zzxn{i}"
        password = f"password{i}"
        username2user[username] = {"nickname": nickname, "password": password}
        req = gpost('/user/register', {"username": username, "password": password, 
                                      "vcode":"0000", "nickname": nickname})
        regReqs.append(req)



    loginReqs = []
    users = {}
    progress = tqdm.trange(n, desc="register")
    for res in grequests.imap(regReqs, size=CONCURRENT):
        assertOK(res)
        progress.update(1)

        userID = res.json()["user_id"]
        username = res.json()["username"]
        mustUser = username2user[username]
        mustPassword = mustUser["password"]

        users[userID] = {"username": username, "password": "password", **res.json()}
        # login
        loginReqs.append(gpost('/user/login', {"username": username, "password": mustPassword}))  

    progress = tqdm.trange(n, desc="login")
    for res in grequests.imap(loginReqs, size=CONCURRENT):
        assertOK(res)
        progress.update(1)

        userID = res.json()["user_id"]
        mustUser = users[userID]
        mustNickname = mustUser["nickname"]
        
        assert res.json()["username"] == mustUser["username"]
        assert res.json()["nickname"] == mustNickname
        assert res.json()["auth_token"] != ""
    return users


def makePosts(users):
    postList = []
    for authorID, user in tqdm.tqdm(users.items(), desc="create post"):
        for i in range(N_POST_PER_USER):
            p = {"author_id": authorID, "title": f"title{i}{authorID}", "content": f"content{i}{authorID}"}
            res = post('/post/create', p, user['auth_token'])
            assertOK(res)
            p["post_id"] = res.json()["post_id"]
            postList.append(p)
    return postList


def replyPost(postID, users):
    replyList = []
    for authorID, user in tqdm.tqdm(users.items(), desc="create reply"):
        for i in range(N_POST_REPLY_PER_USR):
            p = {"author_id": authorID, "post_id": str(postID), "content": f"content{i}{authorID}"}
            res = post('/post/reply', p, user['auth_token'])
            assertOK(res)
            p["reply_id"] = res.json()["reply_id"]
            replyList.append(p)
    return replyList


def main():
    users = registerUsers(N_USER)
    postList = makePosts(users)

    # query posts according create_time
    expectGen = (p for p in postList[::-1])
    progress = tqdm.trange(len(postList), desc="check post")
    expectCnt = len(postList)
    cursor = ""
    while expectCnt > 0:
        res = get('/post/list', {"order": "create_time", "cursor": cursor})
        assertOK(res)
        cursor, page = res.json()["cursor"], res.json()["list"]
        for realR in page:
            expectCnt -= 1
            assert expectCnt >= 0
            progress.update(1)
            expectR = next(expectGen)
            for field in ["post_id", "author_id", "title", "content"]:
                assert realR[field] == expectR[field], f"expect {expectR}, got {realR}"

    # reply a post
    postID = postList[0]["post_id"]
    replyList = replyPost(postID, users)
   
    # after reply, the replied one should be the first one
    res = get('/post/list', {"order": "reply_time"})
    assertOK(res)
    for field in ["post_id", "author_id", "title", "content"]:
        assert res.json()["list"][0][field] == postList[0][field], f"expect {res.json()['list'][0]}, got {postList[0]}"

    # query replies, check the order is right
    expectGen = (r for r in replyList[::-1])
    progress = tqdm.trange(len(replyList), desc="check reply")
    expectCnt = len(replyList)
    cursor = ""
    while expectCnt > 0:
        res = get('/post/reply/list', {"post_id": postID, "cursor": cursor})
        assertOK(res)
        cursor, page = res.json()["cursor"], res.json()["list"]
        for realR in page:
            expectCnt -= 1
            assert expectCnt >= 0
            progress.update(1)
            expectR = next(expectGen)
            for field in ["reply_id", "author_id", "content"]:
                assert realR[field] == expectR[field], f"expect {expectR}, got {realR}"


main()
