#!/bin/python3

PREFIX='http://localhost:8080/api'
SIZE = 5000
BATCH_SIZE = 10000
CONCURRENT=500 # set it to 1 if using sqlite3
N_USER = SIZE
N_ACTIVE_USER = min(SIZE, 10000)
N_POST = SIZE
N_REPLY_PER_POST = (0, 3)

from faker import Faker
from random import Random

Faker.seed(42)
rng = Random(42)


fakeCN = Faker('zh_CN')
fakeUS = Faker()

seq = (i for i in range(10000000000000000))

def fakeUsername():
    # if False:
    #     return fakeUS.username() + "_" + next(seq)
    # else:
    return fakeCN.unique.phone_number()

def fakePassword():
    return "p@ssw0rd"

def fakeNickname():
    if fakeCN.boolean():
        return fakeCN.user_name()[:10] + "_" + str(next(seq))
    else:
        return fakeCN.name()[:10] + "_" + str(next(seq))

def fakeTitle():
    return fakeCN.sentence()[:50]

def fakeContent():
    return fakeCN.paragraph()[:500]

def split_batch(init_list, batch_size):
    groups = zip(*(iter(init_list),) * batch_size)
    end_list = [list(i) for i in groups]
    count = len(init_list) % batch_size
    end_list.append(init_list[-count:]) if count != 0 else end_list
    return end_list

    
import grequests # this one must be imported fisrt!
import requests
import tqdm



def get(url: str, params, token=None):
    return requests.get(f'{PREFIX}{url}', params=params, headers={"Auth": token} if token else None)

def post(url: str, json, token=None):
    return requests.post(f'{PREFIX}{url}', json=json, headers={"Auth": token} if token else None)

def gpost(url: str, json, token=None):
    return grequests.post(f'{PREFIX}{url}', json=json, headers={"Auth": token} if token else None)

def assertOK(res: requests.Response):
    assert res.status_code == 200, "{} {} Got {}".format(res.request.method, res.request.url, res.json())

def registerUsers(n: int):
    users = []

    progress = tqdm.trange(n, desc="prepare register")
    for batch in split_batch(range(n), BATCH_SIZE):
        regReqs = []
        for _ in batch:
            # register
            username = fakeUsername()
            nickname = fakeNickname()
            password = fakePassword()
            req = gpost('/user/register', {"username": username, "password": password, 
                                        "vcode":"0000", "nickname": nickname})
            regReqs.append(req)

        for res in grequests.imap(regReqs, size=CONCURRENT):
            assertOK(res)
            progress.update(1)
            users.append(res.json())
    return users


def makePosts(users):
    postIDs = []
    progress = tqdm.trange(N_POST)
    for batch in split_batch(range(N_POST), BATCH_SIZE):
        postReqs = []
        for _ in batch:
            user = rng.choice(users)
            authorID = user["user_id"]
            p = {"author_id": authorID, "title": fakeTitle(), "content": fakeContent()}
            postReqs.append(gpost('/post/create', p, user['auth_token']))
        for res in grequests.imap(postReqs, size=CONCURRENT):
            assertOK(res)
            progress.update(1)
            postIDs.append(res.json()["post_id"])
    return postIDs


def replyPosts(postIDs, users):
    postIDs = [e for e in postIDs]
    rng.shuffle(postIDs)
    
    progress = tqdm.tqdm(postIDs, desc="prepare reply req")
    for batch in split_batch(postIDs, BATCH_SIZE):
        replyReqs = []
        for postID in batch:
            nPost = rng.randint(*N_REPLY_PER_POST)
            for _ in range(nPost):
                user = rng.choice(users)
                p = {"author_id": user["user_id"], "post_id": str(postID), "content": fakeContent()}
                replyReqs.append(gpost('/post/reply', p, user['auth_token']))
            progress.update(1)
    
        for res in grequests.imap(replyReqs, size=CONCURRENT):
            assertOK(res)
    return



def main():
    users = registerUsers(N_USER)[:N_ACTIVE_USER]
    postIDs = makePosts(users)
    replyPosts(postIDs, users)

main()

