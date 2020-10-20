package cache

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tal-tech/go-zero/core/errorx"
	"github.com/tal-tech/go-zero/core/hash"
	"github.com/tal-tech/go-zero/core/stores/redis"
	"github.com/tal-tech/go-zero/core/syncx"
)

type mockedNode struct {
	vals        map[string][]byte
	errNotFound error
}

func (mc *mockedNode) DelCache(keys ...string) error {
	var be errorx.BatchError
	for _, key := range keys {
		if _, ok := mc.vals[key]; !ok {
			be.Add(mc.errNotFound)
		} else {
			delete(mc.vals, key)
		}
	}
	return be.Err()
}

func (mc *mockedNode) GetCache(key string, v interface{}) error {
	bs, ok := mc.vals[key]
	if ok {
		return json.Unmarshal(bs, v)
	}

	return mc.errNotFound
}

func (mc *mockedNode) SetCache(key string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	mc.vals[key] = data
	return nil
}

func (mc *mockedNode) SetCacheWithExpire(key string, v interface{}, expire time.Duration) error {
	return mc.SetCache(key, v)
}

func (mc *mockedNode) Take(v interface{}, key string, query func(v interface{}) error) error {
	if _, ok := mc.vals[key]; ok {
		return mc.GetCache(key, v)
	}

	if err := query(v); err != nil {
		return err
	}

	return mc.SetCache(key, v)
}

func (mc *mockedNode) TakeWithExpire(v interface{}, key string, query func(v interface{}, expire time.Duration) error) error {
	return mc.Take(v, key, func(v interface{}) error {
		return query(v, 0)
	})
}

func (mc *mockedNode) Inc(key string, count *int64) error {
	var newVal []byte
	val, ok := mc.vals[key]
	if ok {
		c, err := strconv.ParseInt(string(val), 10 ,64)
		if err != nil {
			return err
		}
		*count = c + 1
		newVal = []byte(strconv.FormatInt(*count, 10))
	} else {
		newVal	= []byte("1")
	}
	mc.vals[key] = newVal
	return nil
}

func (mc *mockedNode) IncBy(key string, increment int64, count *int64) error {
	var newVal []byte
	val, ok := mc.vals[key]
	if ok {
		c, err := strconv.ParseInt(string(val), 10 ,64)
		if err != nil {
			return err
		}
		*count = c + increment
		newVal = []byte(strconv.FormatInt(*count, 10))
	} else {
		newVal	= []byte("1")
	}
	mc.vals[key] = newVal
	return nil
}

func (mc *mockedNode) Count(key string) (int64, error) {
	val, ok := mc.vals[key]
	if ok {
		return strconv.ParseInt(string(val), 10, 64)
	}
	return 0, mc.errNotFound
}

func (mc *mockedNode) Counts(keys ...string) ([]int64, error) {
	sz := len(keys)
	if sz == 0 {
		return []int64{}, nil
	}

	ret := make([]int64, sz)
	for index, key := range keys {
		val, ok := mc.vals[key]
		if ok {
			c, err := strconv.ParseInt(string(val), 10, 64)
			if err == nil {
				ret[index] = c
				continue
			}
		}
		ret[index] = 0
	}
	return ret, nil
}

func TestCache_SetDel(t *testing.T) {
	const total = 1000
	r1, clean1, err := createMiniRedis()
	assert.Nil(t, err)
	defer clean1()
	r2, clean2, err := createMiniRedis()
	assert.Nil(t, err)
	defer clean2()
	conf := ClusterConf{
		{
			RedisConf: redis.RedisConf{
				Host: r1.Addr(),
				Type: redis.NodeType,
			},
			Weight: 100,
		},
		{
			RedisConf: redis.RedisConf{
				Host: r2.Addr(),
				Type: redis.NodeType,
			},
			Weight: 100,
		},
	}
	c := NewCache(conf, syncx.NewSharedCalls(), NewCacheStat("mock"), errPlaceholder)
	for i := 0; i < total; i++ {
		if i%2 == 0 {
			assert.Nil(t, c.SetCache(fmt.Sprintf("key/%d", i), i))
		} else {
			assert.Nil(t, c.SetCacheWithExpire(fmt.Sprintf("key/%d", i), i, 0))
		}
	}
	for i := 0; i < total; i++ {
		var v int
		assert.Nil(t, c.GetCache(fmt.Sprintf("key/%d", i), &v))
		assert.Equal(t, i, v)
	}
	assert.Nil(t, c.DelCache())
	for i := 0; i < total; i++ {
		assert.Nil(t, c.DelCache(fmt.Sprintf("key/%d", i)))
	}
	for i := 0; i < total; i++ {
		var v int
		assert.Equal(t, errPlaceholder, c.GetCache(fmt.Sprintf("key/%d", i), &v))
		assert.Equal(t, 0, v)
	}
}

func TestCache_OneNode(t *testing.T) {
	const total = 1000
	r, clean, err := createMiniRedis()
	assert.Nil(t, err)
	defer clean()
	conf := ClusterConf{
		{
			RedisConf: redis.RedisConf{
				Host: r.Addr(),
				Type: redis.NodeType,
			},
			Weight: 100,
		},
	}
	c := NewCache(conf, syncx.NewSharedCalls(), NewCacheStat("mock"), errPlaceholder)
	for i := 0; i < total; i++ {
		if i%2 == 0 {
			assert.Nil(t, c.SetCache(fmt.Sprintf("key/%d", i), i))
		} else {
			assert.Nil(t, c.SetCacheWithExpire(fmt.Sprintf("key/%d", i), i, 0))
		}
	}
	for i := 0; i < total; i++ {
		var v int
		assert.Nil(t, c.GetCache(fmt.Sprintf("key/%d", i), &v))
		assert.Equal(t, i, v)
	}
	assert.Nil(t, c.DelCache())
	for i := 0; i < total; i++ {
		assert.Nil(t, c.DelCache(fmt.Sprintf("key/%d", i)))
	}
	for i := 0; i < total; i++ {
		var v int
		assert.Equal(t, errPlaceholder, c.GetCache(fmt.Sprintf("key/%d", i), &v))
		assert.Equal(t, 0, v)
	}
}

func TestCache_Balance(t *testing.T) {
	const (
		numNodes = 100
		total    = 10000
	)
	dispatcher := hash.NewConsistentHash()
	maps := make([]map[string][]byte, numNodes)
	for i := 0; i < numNodes; i++ {
		maps[i] = map[string][]byte{
			strconv.Itoa(i): []byte(strconv.Itoa(i)),
		}
	}
	for i := 0; i < numNodes; i++ {
		dispatcher.AddWithWeight(&mockedNode{
			vals:        maps[i],
			errNotFound: errPlaceholder,
		}, 100)
	}

	c := cacheCluster{
		dispatcher:  dispatcher,
		errNotFound: errPlaceholder,
	}
	for i := 0; i < total; i++ {
		assert.Nil(t, c.SetCache(strconv.Itoa(i), i))
	}

	counts := make(map[int]int)
	for i, m := range maps {
		counts[i] = len(m)
	}
	entropy := calcEntropy(counts, total)
	assert.True(t, len(counts) > 1)
	assert.True(t, entropy > .95, fmt.Sprintf("entropy should be greater than 0.95, but got %.2f", entropy))

	for i := 0; i < total; i++ {
		var v int
		assert.Nil(t, c.GetCache(strconv.Itoa(i), &v))
		assert.Equal(t, i, v)
	}

	for i := 0; i < total/10; i++ {
		assert.Nil(t, c.DelCache(strconv.Itoa(i*10), strconv.Itoa(i*10+1), strconv.Itoa(i*10+2)))
		assert.Nil(t, c.DelCache(strconv.Itoa(i*10+9)))
	}

	var count int
	for i := 0; i < total/10; i++ {
		var val int
		if i%2 == 0 {
			assert.Nil(t, c.Take(&val, strconv.Itoa(i*10), func(v interface{}) error {
				*v.(*int) = i
				count++
				return nil
			}))
		} else {
			assert.Nil(t, c.TakeWithExpire(&val, strconv.Itoa(i*10), func(v interface{}, expire time.Duration) error {
				*v.(*int) = i
				count++
				return nil
			}))
		}
		assert.Equal(t, i, val)
	}
	assert.Equal(t, total/10, count)
}

func TestCacheNoNode(t *testing.T) {
	dispatcher := hash.NewConsistentHash()
	c := cacheCluster{
		dispatcher:  dispatcher,
		errNotFound: errPlaceholder,
	}
	assert.NotNil(t, c.DelCache("foo"))
	assert.NotNil(t, c.DelCache("foo", "bar", "any"))
	assert.NotNil(t, c.GetCache("foo", nil))
	assert.NotNil(t, c.SetCache("foo", nil))
	assert.NotNil(t, c.SetCacheWithExpire("foo", nil, time.Second))
	assert.NotNil(t, c.Take(nil, "foo", func(v interface{}) error {
		return nil
	}))
	assert.NotNil(t, c.TakeWithExpire(nil, "foo", func(v interface{}, duration time.Duration) error {
		return nil
	}))
}

func calcEntropy(m map[int]int, total int) float64 {
	var entropy float64

	for _, v := range m {
		proba := float64(v) / float64(total)
		entropy -= proba * math.Log2(proba)
	}

	return entropy / math.Log2(float64(len(m)))
}

func TestCache_Counts(t *testing.T) {
	r1, clean1, err := createMiniRedis()
	assert.Nil(t, err)
	defer clean1()
	r2, clean2, err := createMiniRedis()
	assert.Nil(t, err)
	defer clean2()
	conf := ClusterConf{
		{
			RedisConf: redis.RedisConf{
				Host: r1.Addr(),
				Type: redis.NodeType,
			},
			Weight: 100,
		},
		{
			RedisConf: redis.RedisConf{
				Host: r2.Addr(),
				Type: redis.NodeType,
			},
			Weight: 100,
		},
	}
	c := NewCache(conf, syncx.NewSharedCalls(), NewCacheStat("mock"), errPlaceholder)

	testKeys := []string{
		"TestCacheNode_Inc",
		"TestCacheNode_IncBy",
		"TestCacheNode_Count",
		"TestCacheNode_Counts",
	}

	// before
	counts, e := c.Counts(testKeys...)
	assert.Nil(t, e)
	assert.NotNil(t, counts)
	assert.Equal(t, len(testKeys), len(counts))
	for _, r := range counts {
		assert.Equal(t, r, int64(0))
	}

	// add increment
	var result int64
	for _, key := range testKeys {
		err = c.Inc(key, &result)
		assert.Nil(t, err)
		assert.Equal(t, result, int64(1))
	}

	// add increment
	for index, key := range testKeys {
		err = c.IncBy(key, int64(index), &result)
		assert.Nil(t, err)
		assert.Equal(t, result, int64(1+index))
	}

	// after
	counts, e = c.Counts(testKeys...)
	assert.Nil(t, e)
	assert.NotNil(t, counts)
	assert.Equal(t, len(testKeys), len(counts))

	for index, r := range counts {
		assert.Equal(t, r, int64(1+index))
	}
}
