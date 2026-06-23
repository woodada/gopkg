package milvus

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/milvus-io/milvus/client/v2/column"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/woodada/gopkg/logger"
)

const DefaultTopK = 10
const DefaultANNSField = "embedding"
const InnerId = "_id"
const InnerScore = "_score"

type Config struct {
	Address string `json:"address"  mapstructure:"address"  toml:"address"  yaml:"address"` // 必填
	Token   string `json:"token"  mapstructure:"token"  toml:"token"  yaml:"token"`         // 必填
	DB      string `json:"db"      mapstructure:"db"      toml:"db"      yaml:"db"`         // 选择哪个库 必填
}

func (c Config) String() string {
	buf, _ := json.Marshal(c)
	return string(buf)
}

type Client struct {
	cfg Config
	cli *milvusclient.Client
}

func NewClient(config Config) *Client {
	return &Client{cfg: config}
}

func (c *Client) Init() error {
	ctx, cancle := context.WithTimeout(context.Background(), time.Second*10)
	defer cancle()
	cli, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address: c.cfg.Address,
		APIKey:  c.cfg.Token,
		DBName:  c.cfg.DB,
	})

	if err != nil {
		logger.Errorf("[milvus] init client error: %v config: %s]", logger.FileLine(FileWithLineNum()), err, c.cfg)
		return err
	}
	logger.Infof("[milvus] init client success! config: %s]", logger.FileLine(FileWithLineNum()), c.cfg)
	c.cli = cli
	return nil
}

func (c *Client) Close() error {
	if c.cli != nil {
		ctx, cancle := context.WithTimeout(context.Background(), time.Second*10)
		defer cancle()
		err := c.cli.Close(ctx)
		if err != nil {
			logger.Errorf("[milvus] close client error: %v]", err)
			return err
		}
	}
	return nil
}

// GetServerVersion 获取服务器版本
func (c *Client) GetServerVersion(ctx context.Context) (string, error) {
	return c.cli.GetServerVersion(ctx, milvusclient.NewGetServerVersionOption())
}

// DropCollection 删除集合
func (c *Client) DropCollection(ctx context.Context, collName string) error {
	st := time.Now()
	err := c.cli.DropCollection(ctx, milvusclient.NewDropCollectionOption(collName))
	us := time.Since(st)
	if err != nil {
		logger.CtxErrorf(ctx, "[milvus] drop %s error! ustime: %d err: %v", logger.FileLine(FileWithLineNum()), collName, us.Milliseconds(), err)
		return err
	}

	if us > time.Second*3 {
		logger.CtxWarnf(ctx, "[milvus] drop %s success! ustime: %d", logger.FileLine(FileWithLineNum()), collName, us.Milliseconds())
	} else {
		logger.CtxDebugf(ctx, "[milvus] drop %s success! ustime: %d", logger.FileLine(FileWithLineNum()), collName, us.Milliseconds())
	}

	return nil
}

// InsertRows 插入数据 rows的每个item是可json化对象 返回插入的数据id列表
func (c *Client) InsertRows(ctx context.Context, collName string, rows []any) ([]string, error) {
	st := time.Now()
	ret, err := c.cli.Insert(ctx, milvusclient.NewRowBasedInsertOption(collName, rows...))
	// 立即计时
	us := time.Since(st)
	// 取得id
	var ids []string
	if ret.IDs != nil {
		for i := 0; i < ret.IDs.Len(); i++ {
			if id1, e1 := ret.IDs.GetAsInt64(i); e1 == nil {
				ids = append(ids, strconv.FormatInt(id1, 10))
			} else if id2, e2 := ret.IDs.GetAsString(i); e2 == nil {
				ids = append(ids, id2)
			}
		}
	}

	if err == nil {
		if len(rows) != int(ret.InsertCount) || us > time.Second*3 {
			logger.CtxWarnf(ctx, "[milvus] insert %s success! rows: %d count: %d ustime: %d]", logger.FileLine(FileWithLineNum()), collName, len(rows), ret.InsertCount, us.Milliseconds())
		} else {
			logger.CtxDebugf(ctx, "[milvus] insert %s success! rows: %d count: %d ustime: %d]", logger.FileLine(FileWithLineNum()), collName, len(rows), ret.InsertCount, us.Milliseconds())
		}
	} else {
		logger.CtxErrorf(ctx, "[milvus] insert %s error! rows: %d count: %d ustime: %d error: %v", logger.FileLine(FileWithLineNum()), collName, len(rows), ret.InsertCount, us.Milliseconds(), err)
	}

	return ids, err
}

// SearchReq 搜索条件
type SearchReq struct {
	TopK         int               `json:"topK"`         // 可空 为空时取默认值 10
	Expr         string            `json:"expr"`         // 可空
	VectorFloats [][]float32       `json:"vectorFloats"` // 按向量搜索 VectorFloats 和 VectorTexts 必填一个
	VectorTexts  []string          `json:"vectorTexts"`  // 按文本搜索 VectorFloats 和 VectorTexts 必填一个
	OutFields    map[string]string `json:"OutFields"`    // 可空 默认返回id score 输出 字段名：类型支持(int64 string bool float64)
	ANNSField    string            `json:"annsField"`    // 指定向量字段名 可空 取默认值 embedding
}

func (s *SearchReq) String() string {
	return fmt.Sprintf("topK: %d anns: %s out: %v vText: %v vF: %d expr: %s", s.TopK, s.ANNSField, s.OutFields, s.VectorTexts, len(s.VectorFloats), s.Expr)
}

type SearchRetItem map[string]any

func (item SearchRetItem) GetId() string {
	if id, ok := item[InnerId]; ok {
		return fmt.Sprint(id)
	}
	return ""
}

func (item SearchRetItem) GetScore() float32 {
	if val, ok1 := item[InnerScore]; ok1 {
		if score, ok2 := val.(float32); ok2 {
			return float32(score)
		}

		if score, ok2 := val.(float64); ok2 {
			return float32(score)
		}

		if score, ok2 := val.(int64); ok2 {
			return float32(score)
		}

		if score, ok2 := val.(int32); ok2 {
			return float32(score)
		}
		panic("score type error")
	} else {
		panic("no inne score")
	}
	return 0
}

func (item SearchRetItem) Get(name string) any {
	if val, ok1 := item[name]; ok1 {
		return val
	}
	return nil
}

// Search 返回结构为 [{"id": "1", "score": float32(0.9), "outfiled1":"约定的type数据"}]
// id和score是内置变量
func (c *Client) Search(ctx context.Context, collName string, req SearchReq) ([]SearchRetItem, error) {
	if req.ANNSField == "" {
		req.ANNSField = DefaultANNSField
	}

	// 如果检索范围为零值 则使用默认值5
	topK := req.TopK
	if topK == 0 {
		topK = DefaultTopK
	}

	// 准备向量
	var vectors []entity.Vector
	if len(req.VectorFloats) > 0 {
		for i := range req.VectorFloats {
			vectors = append(vectors, entity.FloatVector(req.VectorFloats[i]))
		}
	} else if len(req.VectorTexts) > 0 {
		for i := range req.VectorTexts {
			vectors = append(vectors, entity.Text(req.VectorTexts[i]))
		}
	} else {
		return nil, fmt.Errorf("[milvus] search vector empty")
	}

	searchOpt := milvusclient.NewSearchOption(collName, topK, vectors).
		WithSearchParam("nprobe", "16")

	// 输出字段
	if len(req.OutFields) > 0 {
		var outFields []string
		for k, _ := range req.OutFields {
			outFields = append(outFields, k)
		}
		searchOpt = searchOpt.WithOutputFields(outFields...)
	}
	// 指定过滤条件
	if req.Expr != "" {
		searchOpt = searchOpt.WithFilter(req.Expr)
	}
	// 指定向量字段名
	if req.ANNSField != "" {
		searchOpt = searchOpt.WithANNSField(req.ANNSField)
	}

	st := time.Now()
	ret, err := c.cli.Search(ctx, searchOpt)
	us := time.Since(st)
	if err != nil {
		logger.CtxErrorf(ctx, "[milvus] search error: %v usetime: %d collName: %s req: %s", logger.FileLine(FileWithLineNum()), err, us.Milliseconds(), collName, req.String())
		return nil, err
	}

	var results []SearchRetItem

	for _, sr := range ret {

		for i := 0; i < sr.ResultCount; i++ {
			var obj = make(SearchRetItem)
			// 内置id字段
			if id1, e1 := sr.IDs.GetAsInt64(i); e1 == nil {
				obj[InnerId] = fmt.Sprint(id1)
			} else if id2, e2 := sr.IDs.GetAsString(i); e2 == nil {
				obj[InnerId] = id2
			} else {
				id3, e3 := sr.IDs.Get(i)
				logger.CtxErrorf(ctx, "[milvus] search 获取ID异常(%v): %v", logger.FileLine(FileWithLineNum()), e3, id3)
				return nil, fmt.Errorf("[milvus] search 获取ID异常")
			}
			// 内置评分字段
			obj[InnerScore] = sr.Scores[i]
			// 获取自定义输出字段
			for fieldName, v := range req.OutFields {
				fieldVal, e3 := getFieldData(sr.GetColumn(fieldName), i, v)
				if e3 != nil {
					logger.CtxErrorf(ctx, "[milvus] search 获取字段异常 field: %s idx: %d err: %v", logger.FileLine(FileWithLineNum()), fieldName, i, e3)
					return nil, e3
				}
				obj[fieldName] = fieldVal
			}
			results = append(results, obj)
		}

	}

	if us > 5*time.Second {
		logger.CtxWarnf(ctx, "[milvus] search %s success! usetime: %d req: %s rows: %d", logger.FileLine(FileWithLineNum()), collName, us.Milliseconds(), req.String(), len(results))
	} else {
		logger.CtxDebugf(ctx, "[milvus] search %s success! usetime: %d req: %s rows: %d", logger.FileLine(FileWithLineNum()), collName, us.Milliseconds(), req.String(), len(results))
	}

	return results, nil
}

// 按类型获取数据
func getFieldData(col column.Column, idx int, typ string) (any, error) {
	switch typ {
	case "int64":
		return col.GetAsInt64(idx)
	case "string":
		return col.GetAsString(idx)
	case "float64":
		return col.GetAsDouble(idx)
	case "bool":
		return col.GetAsBool(idx)
	default:
		return nil, fmt.Errorf("[milvus] search 不支持的字段类型: %s", typ)
	}
}

type DeleteReq struct {
	Expr string
	Ids  []string
}

func (d DeleteReq) String() string {
	return fmt.Sprintf("expr: %s ids: %v", d.Expr, d.Ids)
}

// Delete 返回删除的行数
func (c *Client) Delete(ctx context.Context, collName string, req DeleteReq) (int64, error) {
	if len(req.Ids) <= 0 && req.Expr == "" {
		return 0, fmt.Errorf("BUG!!! [milvus] delete error: condition is empty")
	}

	deleteOpt := milvusclient.NewDeleteOption(collName)
	if req.Expr != "" {
		deleteOpt = deleteOpt.WithExpr(req.Expr)
	}

	if len(req.Ids) > 0 {
		ids := slice.Map(req.Ids, func(i int, id string) int64 {
			n, _ := strconv.ParseInt(id, 10, 64)
			return n
		})
		deleteOpt = deleteOpt.WithInt64IDs(InnerId, ids) // 目前都是int64类型
	}

	st := time.Now()
	ret, err := c.cli.Delete(ctx, deleteOpt)
	us := time.Since(st)
	if err != nil {
		logger.CtxErrorf(ctx, "[milvus] delete error: %v usetime: %d collName: %s req: %s rows: %d", logger.FileLine(FileWithLineNum()), err, us.Milliseconds(), collName, req.String(), ret.DeleteCount)
		return ret.DeleteCount, err
	}

	if us > 5*time.Second {
		logger.CtxWarnf(ctx, "[milvus] delete %s success! usetime: %d req: %s rows: %d", logger.FileLine(FileWithLineNum()), collName, us.Milliseconds(), req.String(), ret.DeleteCount)
	} else {
		logger.CtxDebugf(ctx, "[milvus] delete %s success! usetime: %d req: %s rows: %d", logger.FileLine(FileWithLineNum()), collName, us.Milliseconds(), req.String(), ret.DeleteCount)
	}

	return ret.DeleteCount, err

}

func (c *Client) LoadCollection(ctx context.Context, collName string) error {
	st := time.Now()
	loadTask, err := c.cli.LoadCollection(ctx, milvusclient.NewLoadCollectionOption(collName))
	us := time.Since(st)
	if err != nil {
		logger.CtxErrorf(ctx, "[milvus] loadd collection create task error: %v usetime: %d collName: %s", logger.FileLine(FileWithLineNum()), err, us.Milliseconds(), collName)
		return err
	}
	err = loadTask.Await(ctx)
	us = time.Since(st)
	if err != nil {
		logger.CtxErrorf(ctx, "[milvus] load collection task await error: %v usetime: %d collName: %s", logger.FileLine(FileWithLineNum()), err, us.Milliseconds(), collName)
		return err
	}
	return nil
}

func (c *Client) CreateCollection(ctx context.Context, collName string, schema *entity.Schema, indexOption milvusclient.CreateIndexOption) error {
	st := time.Now()
	err := c.cli.CreateCollection(ctx, milvusclient.NewCreateCollectionOption(collName, schema).WithIndexOptions(indexOption))
	us := time.Since(st)
	if err != nil {
		logger.CtxErrorf(ctx, "[milvus] create collection error: %v usetime: %d collName: %s", logger.FileLine(FileWithLineNum()), err, us.Milliseconds(), collName)
		return err
	}
	logger.CtxInfof(ctx, "[milvus] create collection success! usetime: %d collName: %s", logger.FileLine(FileWithLineNum()), us.Milliseconds(), collName)
	return nil
}

func (c *Client) CreateIndex(ctx context.Context, collName, fieldName string, indexOpt index.Index) error {
	st := time.Now()
	task, err := c.cli.CreateIndex(ctx, milvusclient.NewCreateIndexOption(collName, "vector", indexOpt))
	us := time.Since(st)
	if err != nil {
		logger.CtxErrorf(ctx, "[milvus] create index task error: %v usetime: %d collName: %s", logger.FileLine(FileWithLineNum()), err, us.Milliseconds(), collName)
		return err
	}
	logger.CtxInfof(ctx, "[milvus] create index task success! usetime: %d collName: %s", logger.FileLine(FileWithLineNum()), us.Milliseconds(), collName)

	err = task.Await(ctx)
	us = time.Since(st)
	if err != nil {
		logger.CtxErrorf(ctx, "[milvus] create index await error: %v usetime: %d collName: %s", logger.FileLine(FileWithLineNum()), err, us.Milliseconds(), collName)
		return err
	}
	logger.CtxInfof(ctx, "[milvus] create index await success! usetime: %d collName: %s", logger.FileLine(FileWithLineNum()), us.Milliseconds(), collName)

	return nil
}
