package db

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

const (
	DefaultMaxActiveSize = 500
	DefaultMaxIdleSize   = 100
	DefaultIdleTimeout   = 600 * time.Second
	DefaultMaxAge        = 2 * 60 * 60 // 2 hours
	DefaultMaxLength     = 4096        // 4K bytes
	DefaultBatchNum      = 10000
)

var Redis_Response_Big_Err = errors.New("ERR redis tempory failure or response big than 50MB")
var Redis_Response_Biger_Err = errors.New("ERR redis tempory failure or response big than 500MB")

type RedisPool interface {
	ActiveCount() int
	Close() error
	Get() redis.Conn
}

type RedisOption struct {
	Pool      RedisPool
	maxAge    int
	maxLength int
}

type RedisDB struct {
	Option *RedisOption
}

func NewRedisOption(network, address, password string) (option *RedisOption) {
	pool := &redis.Pool{
		MaxActive:   DefaultMaxActiveSize,
		MaxIdle:     DefaultMaxIdleSize,
		IdleTimeout: DefaultIdleTimeout,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
		Dial: func() (redis.Conn, error) {
			return dial(network, address, password)
		},
	}
	option = &RedisOption{
		Pool:      pool,
		maxAge:    DefaultMaxAge,
		maxLength: DefaultMaxLength,
	}
	return
}

func (op *RedisOption) Close() (err error) {
	if op.Pool != nil {
		err = op.Pool.Close()
	}
	return
}

func (op *RedisOption) SetMaxAge(age int) {
	if age >= 0 {
		op.maxAge = age
	}
}

func (op *RedisOption) SetMaxLength(length int) {
	if length >= 0 {
		op.maxLength = length
	}
}

func NewRedisDBWithOption(option *RedisOption) (db *RedisDB, err error) {
	db = &RedisDB{
		Option: option,
	}
	_, err = db.ping()
	return db, err
}

// do base function
func (db *RedisDB) do(action string, key string, val ...interface{}) (err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return
	}
	args := []interface{}{key}
	args = append(args, val...)
	if _, err = conn.Do(strings.ToUpper(strings.TrimSpace(action)), args...); err != nil {
		return err
	}
	return nil
}

func (db *RedisDB) Load(key string, val interface{}) (found bool, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		return false, err
	}
	reply, err := conn.Do("GET", key)
	if err != nil {
		return false, err
	}
	if reply == nil {
		return false, nil // no reply was associated with this key
	}
	switch val.(type) {
	case *int, *uint, *int32, *uint32, *int64, *uint64:
		num, err := redis.Int64(reply, err)
		if err != nil {
			return false, err
		}
		rv := reflect.ValueOf(val)
		p := rv.Elem()
		p.SetInt(num)
	default:
		b, err := redis.Bytes(reply, err)
		if err != nil {
			return false, err
		}

		decoder := gob.NewDecoder(bytes.NewBuffer(b))
		err = decoder.Decode(val)
	}

	return true, err
}

func (db *RedisDB) Save(key string, val interface{}) (err error) {
	var storeValue interface{}
	switch val.(type) {
	case int, uint, int32, uint32, int64, uint64:
		storeValue = val
	default:
		buf := new(bytes.Buffer)
		encoder := gob.NewEncoder(buf)
		err = encoder.Encode(val)
		if err != nil {
			return err
		}
		if db.Option.maxLength != 0 && buf.Len() > db.Option.maxLength {
			return errors.New("RedisDB: the value to store is too big")
		}
		storeValue = buf.Bytes()
	}

	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return err
	}
	if db.Option.maxAge == 0 {
		_, err = conn.Do("SET", key, storeValue)
	} else {
		_, err = conn.Do("SET", key, storeValue, "EX", db.Option.maxAge)
	}
	return err
}

func (db *RedisDB) SaveEx(key string, val interface{}, maxAge int) (err error) {
	var storeValue interface{}
	switch val.(type) {
	case int, uint, int32, uint32, int64, uint64, string:
		storeValue = val
	default:
		buf := new(bytes.Buffer)
		encoder := gob.NewEncoder(buf)
		err = encoder.Encode(val)
		if err != nil {
			return err
		}
		if db.Option.maxLength != 0 && buf.Len() > db.Option.maxLength {
			return errors.New("RedisDB: the value to store is too big")
		}
		storeValue = buf.Bytes()
	}

	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return err
	}
	if maxAge == 0 {
		_, err = conn.Do("SET", key, storeValue)
	} else {
		_, err = conn.Do("SET", key, storeValue, "EX", maxAge)
	}
	return err
}

func (db *RedisDB) SetNxEx(key string, val int, expire int) (success bool, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return
	}

	res, err := redis.String(conn.Do("SET", key, val, "EX", expire, "NX"))
	if err != nil {
		return
	}
	fmt.Printf("res: %s, err: %+v\n", res, err)

	return res == "OK", nil
}

func (db *RedisDB) SaveByte(key string, val interface{}, maxAge int) (err error) {

	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return err
	}
	if maxAge == 0 {
		_, err = conn.Do("SET", key, val)
	} else {
		_, err = conn.Do("SET", key, val, "EX", maxAge)
	}
	return err
}

func (db *RedisDB) GetString(key string) (content string, err error) { //refactor
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return
	}
	content, err = redis.String(conn.Do("GET", key))
	return
}

func (db *RedisDB) GetByte(key string) (content []byte, err error) { //refactor
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return
	}
	content, err = redis.Bytes(conn.Do("GET", key))
	return
}

func (db *RedisDB) GetInt64(key string) (content int64, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return
	}
	content, err = redis.Int64(conn.Do("GET", key))
	return
}

func (db *RedisDB) GetInt(key string) (content int, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return
	}
	content, err = redis.Int(conn.Do("GET", key))
	return
}

func (db *RedisDB) GetBool(key string) (val bool, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return
	}

	val, err = redis.Bool(conn.Do("GET", key))
	return
}

func (db *RedisDB) SetString(key string, content string) (err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return
	}
	_, err = conn.Do("SET", key, content)
	return
}

func (db *RedisDB) Increase(key string) (counter int64, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	counter, err = redis.Int64(conn.Do("INCR", key))
	return
}

func (db *RedisDB) IncrBy(key string, counter int64) (result int64, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	result, err = redis.Int64(conn.Do("INCRBY", key, counter))
	return
}

func (db *RedisDB) Decrease(key string) (counter int, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	counter, err = redis.Int(conn.Do("DECR", key))
	return
}

func (db *RedisDB) Delete(key string) (err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if _, err = conn.Do("DEL", key); err != nil {
		return err
	}
	return nil
}

func (db *RedisDB) Keys(pattern string) (keys []string, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	values, err := redis.Values(conn.Do("KEYS", pattern))
	if err != nil {
		return nil, err
	}
	err = redis.ScanSlice(values, &keys)
	return
}

func (db *RedisDB) SAdd(key string, member string) (err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if _, err = conn.Do("SADD", key, member); err != nil {
		return err
	}
	return nil
}

func (db *RedisDB) Spop(key string) (member string, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()

	return redis.String(conn.Do("SPOP", key))
}

func (db *RedisDB) SRem(key string, member string) (err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if _, err = conn.Do("SREM", key, member); err != nil {
		return err
	}
	return nil
}

func (db *RedisDB) SRems(key string, members []string) (err error) {
	args := []interface{}{key}
	for _, member := range members {
		args = append(args, member)
	}

	conn := db.Option.Pool.Get()
	defer conn.Close()
	if _, err = conn.Do("SREM", args...); err != nil {
		return err
	}
	return nil
}

func (db *RedisDB) Expire(key string, maxAge int) (err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if _, err = conn.Do("EXPIRE", key, maxAge); err != nil {
		return err
	}
	return nil
}

func (db *RedisDB) SAddMult(params ...string) (err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	args := redis.Args{}
	if _, err = conn.Do("SADD", args.AddFlat(params)...); err != nil {
		return err
	}
	return nil
}

func (db *RedisDB) SAddMultValues(key string, values []string) (err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	args := []interface{}{key}
	for _, v := range values {
		args = append(args, v)
	}
	if _, err = conn.Do("SADD", args...); err != nil {
		return err
	}
	return nil
}

func (db *RedisDB) Scard(key string) (num int, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	return redis.Int(conn.Do("SCARD", key))
}

func (db *RedisDB) SMembers(key string) (members []string, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	values, err := redis.Values(conn.Do("SMEMBERS", key))
	if err != nil {
		if isResponseBigErr(err) {
			err = nil
			return db.LoadLargeValues(key)
		}
		return nil, err
	}
	err = redis.ScanSlice(values, &members)
	return
}

func (db *RedisDB) SisMember(key string, member interface{}) (has bool, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	value, err := redis.Int(conn.Do("SISMEMBER", key, member))
	if err != nil {
		return
	}

	has = value == 1
	return
}

func (db *RedisDB) SINTER(keys ...string) (members []string, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	args := redis.Args{}
	values, err := redis.Values(conn.Do("SINTER", args.AddFlat(keys)...))
	if err != nil {
		return nil, err
	}
	err = redis.ScanSlice(values, &members)
	return
}

func (db *RedisDB) SDIFF(keys ...string) (members []string, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	args := redis.Args{}
	values, err := redis.Values(conn.Do("SDIFF", args.AddFlat(keys)...))
	if err != nil {
		return nil, err
	}
	err = redis.ScanSlice(values, &members)
	return
}

func (db *RedisDB) SDIFFSTORE(keys ...string) (err error) {
	var val []interface{}
	for _, v := range keys {
		val = append(val, v)
	}
	conn := db.Option.Pool.Get()
	defer conn.Close()

	_, err = conn.Do("SDIFFSTORE", val...)
	return
}

func (db *RedisDB) SRandMember(key string, count int) (members []string, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()

	values, err := redis.Values(conn.Do("SRANDMEMBER", key, count))
	if err != nil {
		return nil, err
	}
	err = redis.ScanSlice(values, &members)
	return
}

func (db *RedisDB) Exists(key string) (found bool, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return
	}

	value, err := redis.Int64(conn.Do("EXISTS", key))
	if err != nil {
		return false, err
	}

	if value == 0 {
		return false, nil
	}

	return true, nil
}

func (db *RedisDB) ZAdd(key string, score int64, member string) (err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if _, err = conn.Do("ZADD", key, score, member); err != nil {
		return err
	}
	return nil
}

// ZAddMulti first element of val must be key
func (db *RedisDB) ZAddMulti(val []interface{}) (err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return
	}

	if _, err = conn.Do("ZADD", val...); err != nil {
		return err
	}
	return nil
}
func (db *RedisDB) ZRem(key string, member string) (err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if _, err = conn.Do("ZREM", key, member); err != nil {
		return err
	}
	return nil
}

func (db *RedisDB) ZRemRangeByRank(key string, start, stop int) (err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if _, err = conn.Do("ZREMRANGEBYRANK", key, start, stop); err != nil {
		return err
	}
	return nil

}

func (db *RedisDB) ZRemRangeByScore(key string, start, stop int) (err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if _, err = conn.Do("ZREMRANGEBYSCORE", key, start, stop); err != nil {
		return err
	}
	return nil

}

func (db *RedisDB) ZRemBatch(key string, members []string) (err error) {
	args := []interface{}{key}
	for _, member := range members {
		args = append(args, member)
	}
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if _, err = conn.Do("ZREM", args...); err != nil {
		return err
	}
	return nil
}

func (db *RedisDB) ZRange(key string, start, stop int) (members []string, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return
	}

	values, err := redis.Values(conn.Do("ZRANGE", key, start, stop))
	if err != nil {
		return nil, err
	}

	err = redis.ScanSlice(values, &members)
	return
}

func (db *RedisDB) ZRangeWithScores(key string, start, stop int) (members []string, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return
	}

	members, err = redis.Strings(conn.Do("ZRANGE", key, start, stop, "WITHSCORES"))

	return

}

func (db *RedisDB) ZIncrby(key string, inc int, member string) (err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return
	}

	_, err = conn.Do("ZINCRBY", key, inc, member)
	if err != nil {
		return err
	}

	return nil
}

func (db *RedisDB) ZUnionstore(targetKey string, originKey string) (err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return
	}

	_, err = conn.Do("ZUNIONSTORE", targetKey, 1, originKey)
	if err != nil {
		return err
	}

	return nil
}

func (db *RedisDB) ZRank(key, member string) (rank int64, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return
	}

	rank, err = redis.Int64(conn.Do("ZRANK", key, member))
	if err != nil {
		return 0, err
	}

	return
}

func (db *RedisDB) MGetBytes(keys []string) (bs [][]byte, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return
	}
	args := redis.Args{}
	var values []interface{}
	values, err = redis.Values(conn.Do("MGET", args.AddFlat(keys)...))
	var b []byte
	for _, v := range values {
		if v != nil {
			if b, err = redis.Bytes(v, err); err != nil {
				continue
			}
			bs = append(bs, b)
		}
	}
	err = nil
	return
}

func (db *RedisDB) MGetOrigin(keys []string) (values []interface{}, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return
	}
	args := redis.Args{}
	values, err = redis.Values(conn.Do("MGET", args.AddFlat(keys)...))
	return values, err
}

func (db *RedisDB) ZRevRange(key string, start, stop int64, val interface{}) (err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		return err
	}
	values, err := redis.Values(conn.Do("ZREVRANGE", key, start, stop, "WITHSCORES"))
	if err != nil {
		return err
	}
	err = redis.ScanSlice(values, val)
	return
}

func (db *RedisDB) ZRangeByScoreWithScores(key string, start, stop int64, val interface{}) (err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		return err
	}
	values, err := redis.Values(conn.Do("ZRANGEBYSCORE", key, start, stop, "WITHSCORES"))
	if err != nil {
		return err
	}
	err = redis.ScanSlice(values, val)
	return
}

func (db *RedisDB) ZRangeByScore(key string, start, stop int64, val interface{}) (err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		return err
	}
	values, err := redis.Values(conn.Do("ZRANGEBYSCORE", key, start, stop))
	if err != nil {
		return err
	}
	err = redis.ScanSlice(values, val)
	return
}

func (db *RedisDB) ZRevRank(key string, member string) (found bool, rank int64, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		return false, 0, err
	}
	found = true
	rank, err = redis.Int64(conn.Do("ZREVRANK", key, member))
	if err == redis.ErrNil {
		return false, 0, nil
	}
	return
}

func (db *RedisDB) ZRevRanks(key string, members []string) (founds []bool, ranks []int64, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		return nil, nil, err
	}
	for _, m := range members {
		conn.Send("ZREVRANK", key, m, "WITHSCORES")
	}
	conn.Flush()
	var rank int64
	for i, _ := range members {
		rank, err = redis.Int64(conn.Receive())
		if err == redis.ErrNil {
			founds[i] = false
			ranks[i] = 0
		} else {
			founds[i] = true
			ranks[i] = rank
		}
	}
	err = nil
	return
}

func (db *RedisDB) ZScore(key string, member string) (found bool, score float64, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err := conn.Err(); err != nil {
		return false, 0, err
	}
	found = true
	score, err = redis.Float64(conn.Do("ZSCORE", key, member))
	if err == redis.ErrNil {
		return false, 0, nil
	}
	return
}

func (db *RedisDB) MExists(keys []string) (found []bool, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return
	}
	for _, k := range keys {
		conn.Send("EXISTS", k)
	}
	conn.Flush()
	for _, _ = range keys {
		v, err := conn.Receive()
		if err != nil {
			log.Println("redis err:", err.Error())
			v = false
		}
		found = append(found, v == int64(1))
	}
	return found, err
}

func (db *RedisDB) LPushBatch(key string, members []string) (err error) {
	args := []interface{}{key}
	for _, member := range members {
		args = append(args, member)
	}
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if _, err = conn.Do("LPUSH", args...); err != nil {
		return err
	}
	return nil
}

func (db *RedisDB) LRange(key string, start, stop int64) (members []string, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return
	}

	values, err := redis.Values(conn.Do("LRANGE", key, start, stop))
	if err != nil {
		return nil, err
	}

	err = redis.ScanSlice(values, &members)
	return
}

func (db *RedisDB) RPush(key string, members []string) (err error) {
	args := []interface{}{key}
	for _, member := range members {
		args = append(args, member)
	}
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if _, err = conn.Do("RPUSH", args...); err != nil {
		return err
	}
	return nil
}

func (db *RedisDB) LPop(key string) (member string, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()

	reply, err := conn.Do("LPOP", key)
	if err != nil {
		return "", err
	}

	return redis.String(reply, err)
}

func (db *RedisDB) BRPop(key string) (rlt []string, err error) {
	args := []interface{}{key}
	args = append(args, "0")
	conn := db.Option.Pool.Get()
	defer conn.Close()
	values, err := redis.Values(conn.Do("BRPOP", args...))
	if err != nil {
		return
	}
	err = redis.ScanSlice(values, &rlt)
	return
}

func (db *RedisDB) Rpop(key string) (member string, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()

	return redis.String(conn.Do("RPOP", key))
}

func dial(network, address, password string) (redis.Conn, error) {
	c, err := redis.Dial(network, address)
	if err != nil {
		return nil, err
	}
	if password != "" {
		if _, err := c.Do("AUTH", password); err != nil {
			c.Close()
			return nil, err
		}
	}
	return c, err
}

func (db *RedisDB) ping() (bool, error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	data, err := conn.Do("PING")
	if err != nil || data == nil {
		return false, err
	}
	return (data == "PONG"), nil
}

func (db *RedisDB) HMSet(key string, kv ...interface{}) (err error) {
	var val = []interface{}{key}
	val = append(val, kv...)
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return
	}

	if _, err = conn.Do("HMSET", val...); err != nil {
		return err
	}
	return nil
}

// PFAdd <=> PFADD key element [element]
func (db *RedisDB) PFAdd(key string, val ...interface{}) (err error) {
	return db.do("PFADD", key, val...)
}

// PFCount <=> PFCOUNT key element and others [others]
func (db *RedisDB) PFCount(key string, others ...interface{}) (count int64, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return
	}
	args := []interface{}{key}
	args = append(args, others...)
	return redis.Int64(conn.Do("PFCOUNT", args...))
}

// PFMerge <=> PFMERGE destkey sourcekey [sourcekey...]
func (db *RedisDB) PFMerge(destkey string, sourcekey string, others ...interface{}) (err error) {
	allkey := []interface{}{sourcekey}
	allkey = append(allkey, others...)
	return db.do("PFMERGE", destkey, allkey...)
}

// HashIncrBy hash increase by...
func (db *RedisDB) HashIncrBy(key string, field string, by int64) (err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	if err = conn.Err(); err != nil {
		return
	}
	if _, err = conn.Do("HINCRBY", key, field, by); err != nil {
		return err
	}
	return nil
}

// HashGetAllInt hash get all [integer]
//func (db *RedisDB) HashGetAllInt(key string) (result map[string]int64, err error) {
//	conn := db.Option.Pool.Get()
//	defer conn.Close()
//	if err = conn.Err(); err != nil {
//		return
//	}
//	return redis.Int64Map(conn.Do("HGETALL", key))//todo 打开
//}

func (db *RedisDB) HashKeys(key string) (keys []string, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()

	values, err := redis.Values(conn.Do("HKEYS", key))
	if err != nil {
		return
	}

	err = redis.ScanSlice(values, &keys)
	return
}

func (db *RedisDB) HGet(key, field string) (member string, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()

	reply, err := conn.Do("HGET", key, field)
	if err != nil {
		return "", err
	}

	return redis.String(reply, err)
}

const Large_Count = 500000
const Cursor_Start_Stop = "0"

func (db *RedisDB) LoadLargeValues(key string) (dau []string, err error) {
	conn := db.Option.Pool.Get()
	defer conn.Close()
	cursor := Cursor_Start_Stop
	for {
		ks := make([]string, 0)
		ks, cursor, err = sscan(key, cursor, conn)
		if err != nil {
			return
		}
		dau = stringSliceMerge(dau, ks)
		if cursor == Cursor_Start_Stop {
			break
		}
	}

	return
}

func sscan(key, cursor string, conn redis.Conn) (ks []string, c string, err error) {
	sscanReply, err := redis.Values(conn.Do("sscan", key, cursor, "count", Large_Count))
	if err != nil {
		return
	}
	c, err = redis.String(sscanReply[0], nil)
	if err != nil {
		return
	}
	err = redis.ScanSlice(sscanReply[1].([]interface{}), &ks)
	return
}

func stringSliceMerge(slice []string, otherSlice []string) (ret []string) {
	mergeMap := make(map[string]string, 0)
	for _, v := range slice {
		mergeMap[v] = v
	}
	for _, v := range otherSlice {
		mergeMap[v] = v
	}
	ret = make([]string, len(mergeMap))
	i := 0 //todo refactor
	for _, s := range mergeMap {
		ret[i] = s
		i++
	}
	return
}

func isResponseBigErr(err error) bool {
	return err.Error() == Redis_Response_Big_Err.Error() || err.Error() == Redis_Response_Biger_Err.Error()
}
