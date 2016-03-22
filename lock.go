package main

import (
	"fmt"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

const DefaultTimeout = 10 * time.Second

type Lock struct {
	resource string
	token    string
	conn     redis.Conn
	timeout  time.Duration
}

func (lock *Lock) tryLock() (ok bool, err error) {
	_, err = redis.String(lock.conn.Do("SET", lock.key(), lock.token, "EX", int64(lock.timeout/time.Second), "NX"))
	if err == redis.ErrNil {
		// The lock was not successful, it already exists.
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (lock *Lock) Unlock() (err error) {
	_, err = lock.conn.Do("del", lock.key())
	return
}

func (lock *Lock) key() string {
	return fmt.Sprintf("redislock:%s", lock.resource)
}

func (lock *Lock) AddTimeout(ex_time int64) (ok bool, err error) {
	ttl_time, err := redis.Int64(lock.conn.Do("TTL", lock.key()))
	fmt.Println(ttl_time)
	if err != nil {
		log.Fatal("redis get failed:", err)
	}
	if ttl_time > 0 {
		fmt.Println(11)
		_, err := redis.String(lock.conn.Do("SET", lock.key(), lock.token, "EX", int64(ttl_time+ex_time)))
		if err == redis.ErrNil {
			return false, nil
		}
		if err != nil {
			return false, err
		}
	}
	return false, nil
}

func TryLock(conn redis.Conn, resource string, token string) (lock *Lock, ok bool, err error) {
	return TryLockWithTimeout(conn, resource, token, DefaultTimeout)
}

func TryLockWithTimeout(conn redis.Conn, resource string, token string, timeout time.Duration) (lock *Lock, ok bool, err error) {
	lock = &Lock{resource, token, conn, timeout}

	ok, err = lock.tryLock()

	if !ok || err != nil {
		lock = nil
	}

	return
}

func main() {
	fmt.Println("start")
	conn, err := redis.Dial("tcp", "localhost:6379")

	lock, ok, err := TryLock(conn, "xiaoru.cc", "token")
	if err != nil {
		log.Fatal("Error while attempting lock")
	}
	if !ok {
		log.Fatal("bug")
	}
	lock.AddTimeout(100)
	fmt.Println("end")
	time.Sleep(DefaultTimeout + 5)
	defer lock.Unlock()
}
