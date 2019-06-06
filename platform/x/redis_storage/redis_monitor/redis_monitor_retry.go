package redis_monitor

import (
	"errors"
	"fmt"

	"vcs.taiyouxi.net/platform/planx/redigo/redis"
	"vcs.taiyouxi.net/platform/planx/util/storehelper"
)

func (r *RedisPubMonitor) checkConn() {
	if r.state == MONITOR_STATE_Offline {
		return
	}

	if storehelper.PingRedisConn(r.psc.Conn) != nil {
		r.state = MONITOR_STATE_Offline
	} else {
		r.state = MONITOR_STATE_Monitoring
	}
}

func (r *RedisPubMonitor) retry() error {
	if r.psc != nil {
		r.stopPsc()
	}

	// 因为只有很少的监听链接存在，所以这里实现成固定时长去重试
	conn, ok := storehelper.ReBuildRedisConn(r.pool, 2, 1000)
	if !ok {
		return errors.New("ReBuildRedisConn Err")
	}

	err := r.startPsc(conn.RawConn())
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisPubMonitor) startPsc(conn redis.Conn) error {
	// aws 不支持 config命令
	//config_key := "notify-keyspace-events"
	//_, err := conn.Do("config", "set", config_key, "ghE")
	//if err != nil {
	//	return err
	//}
	if conn == nil {
		return fmt.Errorf("start Psc error, the conn is nil.")
	}

	r.psc = &redis.PubSubConn{Conn: conn}

	//TODO YZH 这里如果%d换成*, 并使用PSubscribe应该可以监听更多redis数据库
	if err := r.psc.Subscribe(fmt.Sprintf("__keyevent@%d__:hset", r.db_num)); err != nil {
		return err
	}
	if err := r.psc.Subscribe(fmt.Sprintf("__keyevent@%d__:hdel", r.db_num)); err != nil {
		return err
	}

	return nil
}

func (r *RedisPubMonitor) stopPsc() {
	if r.psc == nil {
		return
	}
	r.psc.Unsubscribe(fmt.Sprintf("__keyevent@%d__:hset", r.db_num))
	r.psc.Unsubscribe(fmt.Sprintf("__keyevent@%d__:hdel", r.db_num))
}
