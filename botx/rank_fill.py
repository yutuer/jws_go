# coding=utf8
# 用于填充排行榜，为了测试排行榜尾的奖励是否能收到
import redis

cache0 = redis.StrictRedis(host='vpc1-qa1.qzg0t0.0001.cnn1.cache.amazonaws.com.cn', port=6379, db = 2)
cache8 = redis.StrictRedis(host='vpc1-qa1.qzg0t0.0001.cnn1.cache.amazonaws.com.cn', port=6379, db = 8)
form_id = 'bacb8503-bd60-49c3-99b7-3477b45e7360'


redis_size = 980

def cpdb():
    v_profile = cache0.dump('profile:1:14:' + form_id)
    v_bag = cache0.dump('bag:1:14:' + form_id)
    v_tmp = cache0.dump('tmp:1:14:' + form_id)
    v_store = cache0.dump('store:1:14:' + form_id)
    v_general = cache0.dump('general:1:14:' + form_id)

    for i in range(redis_size):
        key = 'profile:1:14:' + str(i)
        print 'profile has restore ' + str(i)
        cache0.restore(key, 0, v_profile)

    for i in range(redis_size):
        key = 'bag:1:14:' + str(i)
        print 'bag has restore ' + str(i)
        cache0.restore(key, 0, v_bag)

    for i in range(redis_size):
        key = 'tmp:1:14:' + str(i)
        print 'tmp has restore ' + str(i)
        cache0.restore(key, 0, v_tmp)

    for i in range(redis_size):
        key = 'store:1:14:' + str(i)
        print 'store has restore ' + str(i)
        cache0.restore(key, 0, v_store)

    for i in range(redis_size):
        key = 'general:1:14:' + str(i)
        print 'general has restore ' + str(i)
        cache0.restore(key, 0, v_general)

def dorank():
    rank1 = "14:RankCorpBossSingleLow"
    rank2 = "14:RankCorpBossTotalLow"
    for i in range(redis_size):
        cache8.zadd(rank1, 100, "1:14:"+str(i))
        cache8.zadd(rank2, 100, "1:14:"+str(i))

dorank()