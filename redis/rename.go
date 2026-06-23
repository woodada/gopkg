package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

// RenameKey 实现 Pika 中键的改名功能（模拟RENAME）
// 参数：client- Redis客户端，oldKey-原键名，newKey-新键名
// 返回：error-操作过程中的错误
func RenameKey(client RedisClient, oldKey, newKey string) error {
	if oldKey == newKey {
		return nil
	}
	ctx := context.Background()

	// 1. 检查原键是否存在
	exists, err := client.Exists(ctx, oldKey).Result()
	if err != nil {
		return fmt.Errorf("检查原键是否存在失败: %v", err)
	}
	if exists == 0 {
		return nil
	}

	// 2. 获取原键的类型
	keyType, err := client.Type(ctx, oldKey).Result()
	if err != nil {
		return fmt.Errorf("获取键类型失败: %v", err)
	}

	// 3. 根据键类型执行复制操作
	switch keyType {
	case "string":
		if err := copyString(client, ctx, oldKey, newKey); err != nil {
			return err
		}
	case "hash":
		if err := copyHash(client, ctx, oldKey, newKey); err != nil {
			return err
		}
	case "list":
		if err := copyList(client, ctx, oldKey, newKey); err != nil {
			return err
		}
	case "set":
		if err := copySet(client, ctx, oldKey, newKey); err != nil {
			return err
		}
	case "zset":
		if err := copyZSet(client, ctx, oldKey, newKey); err != nil {
			return err
		}
	case "none": // key not found
		return nil
	default:
		return fmt.Errorf("不支持的键类型: %s", keyType)
	}

	ttl, err := client.TTL(ctx, oldKey).Result()
	if err != nil {
		if IsKeyNotFound(err) {
			return nil
		}
		return fmt.Errorf("获取原键TTL %s 失败: %v", oldKey, err)
	}
	if ttl > 0 {
		client.Expire(ctx, newKey, ttl)
	}

	// 4. 复制成功后删除原键
	if err := client.Del(ctx, oldKey).Err(); err != nil {
		return fmt.Errorf("删除原键 %s 失败: %v", oldKey, err)
	}

	return nil
}

// 复制字符串类型键
func copyString(client RedisClient, ctx context.Context, oldKey, newKey string) error {
	val, err := client.Get(ctx, oldKey).Result()
	if err != nil {
		if IsKeyNotFound(err) {
			return nil
		}
		return fmt.Errorf("读取字符串键 %s 失败: %v", oldKey, err)
	}
	// 写入新键（覆盖已有键）
	if err := client.Set(ctx, newKey, val, 0).Err(); err != nil {
		return fmt.Errorf("写入字符串键 %s 失败: %v", newKey, err)
	}
	return nil
}

// 复制哈希类型键
func copyHash(client RedisClient, ctx context.Context, oldKey, newKey string) error {
	// 获取哈希所有字段和值
	fields, err := client.HGetAll(ctx, oldKey).Result()
	if err != nil {
		if IsKeyNotFound(err) {
			return nil
		}
		return fmt.Errorf("读取哈希键 %s 失败: %v", oldKey, err)
	}
	if len(fields) == 0 {
		return nil
	}

	for k, v := range fields {
		// 批量写入新哈希（覆盖已有键）
		if err := client.HSet(ctx, newKey, k, v).Err(); err != nil {
			return fmt.Errorf("写入哈希键 %s 失败: %v", newKey, err)
		}
	}
	return nil
}

// 复制列表类型键（保持元素顺序）
func copyList(client RedisClient, ctx context.Context, oldKey, newKey string) error {
	// 先删除新键（避免与原有元素冲突）
	if err := client.Del(ctx, newKey).Err(); err != nil {
		return fmt.Errorf("清理列表新键 %s 失败: %v", newKey, err)
	}
	// 读取列表长度
	len, err := client.LLen(ctx, oldKey).Result()
	if err != nil {
		return fmt.Errorf("获取列表长度失败: %v", err)
	}
	if len == 0 {
		return nil // 空列表无需额外操作
	}
	// 读取所有元素（0到-1表示全部）
	elements, err := client.LRange(ctx, oldKey, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("读取列表元素失败: %v", err)
	}
	// 批量写入新列表（保持原有顺序）
	if err := client.RPush(ctx, newKey, elements).Err(); err != nil {
		return fmt.Errorf("写入列表新键 %s 失败: %v", newKey, err)
	}
	return nil
}

// 复制集合类型键
func copySet(client RedisClient, ctx context.Context, oldKey, newKey string) error {
	// 获取集合所有元素
	members, err := client.SMembers(ctx, oldKey).Result()
	if err != nil {
		return fmt.Errorf("读取集合键 %s 失败: %v", oldKey, err)
	}
	if len(members) == 0 {
		// 空集合也需要创建新键
		if err := client.SAdd(ctx, newKey, "").Err(); err != nil && err != redis.Nil {
			return fmt.Errorf("创建空集合键 %s 失败: %v", newKey, err)
		}
		return nil
	}
	// 批量写入新集合（覆盖已有键）
	if err := client.SAdd(ctx, newKey, members).Err(); err != nil {
		return fmt.Errorf("写入集合键 %s 失败: %v", newKey, err)
	}
	return nil
}

// 复制有序集合类型键（保持分数和顺序）
func copyZSet(client RedisClient, ctx context.Context, oldKey, newKey string) error {
	// 获取有序集合所有元素及分数（WITHSCORES参数）
	// 返回格式: [member1, score1, member2, score2, ...]
	zMembers, err := client.ZRangeWithScores(ctx, oldKey, 0, -1).Result()
	if err != nil {
		return fmt.Errorf("读取有序集合 %s 失败: %v", oldKey, err)
	}
	if len(zMembers) == 0 {
		// 空有序集合也需要创建新键
		if err := client.ZAdd(ctx, newKey, redis.Z{Score: 0, Member: ""}).Err(); err != nil && err != redis.Nil {
			return fmt.Errorf("创建空有序集合 %s 失败: %v", newKey, err)
		}
		return nil
	}
	// 转换为ZAdd所需的参数格式
	var zs []redis.Z
	for _, zm := range zMembers {
		zs = append(zs, redis.Z{
			Score:  zm.Score,
			Member: zm.Member,
		})
	}
	// 批量写入新有序集合（覆盖已有键）
	if err := client.ZAdd(ctx, newKey, zs...).Err(); err != nil {
		return fmt.Errorf("写入有序集合 %s 失败: %v", newKey, err)
	}
	return nil
}
