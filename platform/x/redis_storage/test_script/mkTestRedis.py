# coding=utf8
import redis

# 数据源的链接
cache0 = redis.StrictRedis(host='localhost', port=6379, db = 0)

# 测试Redis的链接
cache1 = redis.StrictRedis(host='localhost', port=6379, db = 8)

# 以哪个id为数据源
form_id = '1003'

# 测试Redis需要的大小
redis_size = 1123


v_profile = cache0.dump('profile:0:0:' + form_id)
v_bag = cache0.dump('bag:0:0:' + form_id)

print "Before DB Size " + str(cache1.dbsize())
cache1.flushdb()

for i in range(redis_size):
    key = 'profile:0:0:' + str(i)
    print 'profile has restore ' + str(i)
    cache1.restore(key, 0, v_profile)

for i in range(redis_size):
    key = 'bag:0:0:' + str(i)
    print 'bag has restore ' + str(i)
    cache1.restore(key, 0, v_bag)

print "Now DB Size " + str(cache1.dbsize())
