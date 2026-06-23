package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

// 初始化测试用的Redis客户端（连接本地Pika）
func getTestClient() RedisClient {
	r, _ := InitRedis(RedisConfig{
		Host:     "8.135.12.29",
		Port:     9221,
		Password: "",
		DB:       0,
		PoolSize: 1,
	})
	return r
}

// 清理测试数据
func cleanTestData(client RedisClient, keys ...string) {
	ctx := context.Background()
	if len(keys) > 0 {
		client.Del(ctx, keys...)
	}
}

// TestRenameString 测试字符串类型键的改名
func TestRenameString(t *testing.T) {
	client := getTestClient()
	defer cleanTestData(client, "str_old", "str_new")
	ctx := context.Background()

	// 准备测试数据
	err := client.Set(ctx, "str_old", "hello", 0).Err()
	assert.NoError(t, err)

	// 执行改名
	err = RenameKey(client, "str_old", "str_new")
	assert.NoError(t, err)

	// 验证结果
	// 1. 原键应被删除
	exists, _ := client.Exists(ctx, "str_old").Result()
	assert.Equal(t, int64(0), exists)

	// 2. 新键应存在且值正确
	val, err := client.Get(ctx, "str_new").Result()
	assert.NoError(t, err)
	assert.Equal(t, "hello", val)
}

// TestRenameHash 测试哈希类型键的改名
func TestRenameHash1(t *testing.T) {
	client := getTestClient()
	ctx := context.Background()

	k1 := "dev:topic:20:20,杨立光中医保健官方旗舰店:userId=;name=jsj020122;id=jsj020122:messages"
	k2 := k1 + "#new"

	oldKey, newKey := k1, k2
	//oldKey, newKey := k2, k1

	payload, err := client.HGetAll(ctx, oldKey).Result()
	assert.NoError(t, err)

	// 执行改名
	err = RenameKey(client, oldKey, newKey)
	assert.NoError(t, err)

	// 验证结果
	// 1. 原键应被删除
	exists, _ := client.Exists(ctx, oldKey).Result()
	assert.Equal(t, int64(0), exists)

	// 2. 新键应存在且数据正确
	newData, err := client.HGetAll(ctx, newKey).Result()
	assert.NoError(t, err)
	t.Log(newData)
	//assert.Equal(t, "pika", newData["name"])
	//assert.Equal(t, "kv", newData["type"])

	for k, v := range newData {
		assert.Equal(t, v, payload[k])
	}
}

// TestRenameHash 测试哈希类型键的改名
func TestRenameHash(t *testing.T) {
	client := getTestClient()
	defer cleanTestData(client, "hash_old", "hash_new")
	ctx := context.Background()
	// dev:topic:20:20,杨立光中医保健官方旗舰店:userId=;name=jsj020122;id=jsj020122:messages
	// 准备测试数据
	//hashData := map[string]interface{}{
	//	"name": "pika",
	//	"type": "kv",
	//}
	err := client.HSet(ctx, "hash_old", "name", "pika").Err()
	assert.NoError(t, err)

	err = client.HSet(ctx, "hash_old", "type", "kv").Err()
	assert.NoError(t, err)

	// 执行改名
	err = RenameKey(client, "hash_old", "hash_new")
	assert.NoError(t, err)

	// 验证结果
	// 1. 原键应被删除
	exists, _ := client.Exists(ctx, "hash_old").Result()
	assert.Equal(t, int64(0), exists)

	// 2. 新键应存在且数据正确
	newData, err := client.HGetAll(ctx, "hash_new").Result()
	assert.NoError(t, err)
	assert.Equal(t, "pika", newData["name"])
	assert.Equal(t, "kv", newData["type"])
}

// TestRenameList 测试列表类型键的改名（保持顺序）
func TestRenameList(t *testing.T) {
	client := getTestClient()
	defer cleanTestData(client, "list_old", "list_new")
	ctx := context.Background()

	// 准备测试数据（顺序：a -> b -> c）
	err := client.RPush(ctx, "list_old", "a", "b", "c").Err()
	assert.NoError(t, err)

	// 执行改名
	err = RenameKey(client, "list_old", "list_new")
	assert.NoError(t, err)

	// 验证结果
	// 1. 原键应被删除
	exists, _ := client.Exists(ctx, "list_old").Result()
	assert.Equal(t, int64(0), exists)

	// 2. 新键顺序应与原键一致
	elements, err := client.LRange(ctx, "list_new", 0, -1).Result()
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "c"}, elements)
}

// TestRenameSet 测试集合类型键的改名
func TestRenameSet(t *testing.T) {
	client := getTestClient()
	defer cleanTestData(client, "set_old", "set_new")
	ctx := context.Background()

	// 准备测试数据
	err := client.SAdd(ctx, "set_old", "x", "y", "z").Err()
	assert.NoError(t, err)

	// 执行改名
	err = RenameKey(client, "set_old", "set_new")
	assert.NoError(t, err)

	// 验证结果
	// 1. 原键应被删除
	exists, _ := client.Exists(ctx, "set_old").Result()
	assert.Equal(t, int64(0), exists)

	// 2. 新键元素应与原键一致（集合无序，检查包含关系）
	members, err := client.SMembers(ctx, "set_new").Result()
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"x", "y", "z"}, members) // 忽略顺序比较
}

// TestRenameZSet 测试有序集合类型键的改名（保持分数）
func TestRenameZSet(t *testing.T) {
	client := getTestClient()
	defer cleanTestData(client, "zset_old", "zset_new")
	ctx := context.Background()

	// 准备测试数据（成员+分数）
	zsetData := []redis.Z{
		{Score: 90, Member: "math"},
		{Score: 80, Member: "english"},
	}
	err := client.ZAdd(ctx, "zset_old", zsetData...).Err()
	assert.NoError(t, err)

	// 执行改名
	err = RenameKey(client, "zset_old", "zset_new")
	assert.NoError(t, err)

	// 验证结果
	// 1. 原键应被删除
	exists, _ := client.Exists(ctx, "zset_old").Result()
	assert.Equal(t, int64(0), exists)

	// 2. 新键分数应与原键一致
	newZSet, err := client.ZRangeWithScores(ctx, "zset_new", 0, -1).Result()
	assert.NoError(t, err)
	assert.Len(t, newZSet, 2)
	assert.Equal(t, float64(80), newZSet[0].Score) // 按分数升序排列
	assert.Equal(t, "english", newZSet[0].Member)
	assert.Equal(t, float64(90), newZSet[1].Score)
	assert.Equal(t, "math", newZSet[1].Member)
}

// TestRenameNonExistentKey 测试原键不存在的情况
func TestRenameNonExistentKey(t *testing.T) {
	client := getTestClient()
	defer cleanTestData(client, "non_exist_old", "non_exist_new")

	// 执行改名（原键不存在）
	err := RenameKey(client, "non_exist_old", "non_exist_new")
	assert.NoError(t, err)
}

// TestRenameOverwriteExistingKey 测试新键已存在时的覆盖行为
func TestRenameOverwriteExistingKey(t *testing.T) {
	client := getTestClient()
	defer cleanTestData(client, "overwrite_old", "overwrite_new")
	ctx := context.Background()

	// 准备数据：原键（字符串）和新键（哈希，故意不同类型）
	err := client.Set(ctx, "overwrite_old", "original", 0).Err()
	assert.NoError(t, err)
	err = client.HSet(ctx, "overwrite_new", "field", "old_value").Err()
	assert.NoError(t, err)

	// 执行改名（覆盖新键）
	err = RenameKey(client, "overwrite_old", "overwrite_new")
	assert.NoError(t, err)

	// 验证结果：新键应被替换为原键的类型和值
	// 1. 新键类型应为字符串（原键类型）
	keyType, _ := client.Type(ctx, "overwrite_new").Result()
	assert.Equal(t, "string", keyType)

	// 2. 新键值应为原键值
	val, _ := client.Get(ctx, "overwrite_new").Result()
	assert.Equal(t, "original", val)
}

// TestRenameEmptyKey 测试空键（如空列表、空哈希）的改名
func TestRenameEmptyKey(t *testing.T) {
	client := getTestClient()
	defer cleanTestData(client, "empty_hash_old", "empty_hash_new")
	ctx := context.Background()

	// 准备空哈希
	err := client.HSet(ctx, "empty_hash_old", "", "").Err()
	assert.NoError(t, err)

	// 执行改名
	err = RenameKey(client, "empty_hash_old", "empty_hash_new")
	assert.NoError(t, err)

	// 验证：新键应为空哈希，原键被删除
	existsOld, _ := client.Exists(ctx, "empty_hash_old").Result()
	assert.Equal(t, int64(0), existsOld)

	existsNew, _ := client.Exists(ctx, "empty_hash_new").Result()
	assert.Equal(t, int64(1), existsNew)

	hashLen, _ := client.HLen(ctx, "empty_hash_new").Result()
	assert.Equal(t, int64(1), hashLen)
}

// TestRenameLargeList 测试大列表的改名（验证分批处理兼容性）
func TestRenameLargeList(t *testing.T) {
	client := getTestClient()
	oldKey, newKey := "large_list_old", "large_list_new"
	defer cleanTestData(client, oldKey, newKey)
	ctx := context.Background()

	// 生成包含10000个元素的大列表
	const size = 10000
	elements := make([]interface{}, size)
	for i := 0; i < size; i++ {
		elements[i] = fmt.Sprintf("item%d", i)
	}
	err := client.RPush(ctx, oldKey, elements...).Err()
	assert.NoError(t, err)

	// 执行改名
	err = RenameKey(client, oldKey, newKey)
	assert.NoError(t, err)

	// 验证新列表长度与原列表一致
	newLen, err := client.LLen(ctx, newKey).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(size), newLen)

	// 随机验证部分元素（首尾各取10个）
	head, _ := client.LRange(ctx, newKey, 0, 9).Result()
	tail, _ := client.LRange(ctx, newKey, size-10, size-1).Result()
	assert.Equal(t, "item0", head[0])
	assert.Equal(t, "item9", head[9])
	assert.Equal(t, fmt.Sprintf("item%d", size-10), tail[0])
	assert.Equal(t, fmt.Sprintf("item%d", size-1), tail[9])
}

// TestRenameExpiredKey 测试带过期时间的键改名（验证过期时间是否保留）
func TestRenameExpiredKey(t *testing.T) {
	client := getTestClient()
	oldKey, newKey := "expire_old", "expire_new"
	defer cleanTestData(client, oldKey, newKey)
	ctx := context.Background()

	// 设置带过期时间的字符串键（10秒过期）
	err := client.Set(ctx, oldKey, "temp", 10*time.Second).Err()
	assert.NoError(t, err)

	// 执行改名
	err = RenameKey(client, oldKey, newKey)
	assert.NoError(t, err)

	// 验证新键的过期时间（应接近10秒）
	ttl, err := client.TTL(ctx, newKey).Result()
	assert.NoError(t, err)
	fmt.Println(ttl / time.Second)
	assert.True(t, ttl >= 8*time.Second && ttl <= 10*time.Second)
}

// TestRenameConcurrent 测试并发读写时的改名（验证数据一致性）
func TestRenameConcurrent(t *testing.T) {
	client := getTestClient()
	oldKey, newKey := "concurrent_old", "concurrent_new"
	defer cleanTestData(client, oldKey, newKey)
	ctx := context.Background()

	// 初始化原键（哈希类型）
	err := client.HSet(ctx, oldKey, "count", 0).Err()
	assert.NoError(t, err)

	// 启动10个协程并发向原键写入数据
	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			// 每个协程累加100次
			for j := 0; j < 100; j++ {
				client.HIncrBy(ctx, oldKey, "count", 1)
			}
		}()
	}

	// 等待写入完成后执行改名
	wg.Wait()
	err = RenameKey(client, oldKey, newKey)
	assert.NoError(t, err)

	// 验证新键数据正确性（10*100=1000）
	count, err := client.HGet(ctx, newKey, "count").Int64()
	assert.NoError(t, err)
	assert.Equal(t, int64(1000), count)
}

// TestRenameKeyTypeConflict 测试新键与原键类型冲突（验证覆盖行为）
func TestRenameKeyTypeConflict(t *testing.T) {
	client := getTestClient()
	oldKey, newKey := "type_old", "type_new"
	defer cleanTestData(client, oldKey, newKey)
	ctx := context.Background()

	// 原键：列表；新键：集合（类型冲突）
	err := client.RPush(ctx, oldKey, "a", "b").Err() // 列表
	assert.NoError(t, err)
	err = client.SAdd(ctx, newKey, "x", "y").Err() // 集合
	assert.NoError(t, err)

	// 执行改名（覆盖新键）
	err = RenameKey(client, oldKey, newKey)
	assert.NoError(t, err)

	// 验证新键类型变为列表，且数据正确
	keyType, _ := client.Type(ctx, newKey).Result()
	assert.Equal(t, "list", keyType)

	elements, _ := client.LRange(ctx, newKey, 0, -1).Result()
	assert.Equal(t, []string{"a", "b"}, elements)
}

// TestRenameSameKey 测试原键与新键同名（无操作）
func TestRenameSameKey(t *testing.T) {
	client := getTestClient()
	key := "same_key"
	defer cleanTestData(client, key)
	ctx := context.Background()

	// 初始化数据
	err := client.Set(ctx, key, "test", 0).Err()
	assert.NoError(t, err)

	// 执行改名（新旧键相同）
	err = RenameKey(client, key, key)
	assert.NoError(t, err)

	// 验证键未被删除，数据不变
	val, err := client.Get(ctx, key).Result()
	assert.NoError(t, err)
	assert.Equal(t, "test", val)
}
