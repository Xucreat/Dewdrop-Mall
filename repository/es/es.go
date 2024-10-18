package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
)

type ElasticClient struct {
	client *elasticsearch.Client
}

// NewElasticClient 创建 ES 客户端实例
func NewElasticClient(addresses []string) (*ElasticClient, error) {
	cfg := elasticsearch.Config{
		Addresses: addresses,
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating elasticsearch client: %v", err)
	}
	return &ElasticClient{client: client}, nil
}

// IndexDocument 将单条数据写入 Elasticsearch
func (ec *ElasticClient) IndexDocument(ctx context.Context, id string, data map[string]interface{}) error {
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshalling data: %v", err)
	}

	req := bytes.NewReader(body)
	res, err := ec.client.Index("my-index", req, ec.client.Index.WithDocumentID(id))
	if err != nil {
		return fmt.Errorf("error indexing document: %v", err)
	}
	defer res.Body.Close()

	log.Printf("Document indexed with ID: %s", id)
	return nil
}

// IndexBulkDocuments 批量写入 Elasticsearch
func (ec *ElasticClient) IndexBulkDocuments(ctx context.Context, index string, bulkData []map[string]interface{}) error {
	var buf bytes.Buffer

	for _, data := range bulkData {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_index" : "%s" } }%s`, index, "\n"))
		buf.Write(meta)

		dataBytes, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("error marshalling bulk data: %v", err)
		}
		buf.Write(dataBytes)
		buf.WriteString("\n")
	}
	/*通过ec.client.Bulk方法发起一个批量请求，使用bytes.NewReader(buf.Bytes())作为请求体的数据来源,
	WithIndex(index)选项指定了批量操作应该针对的索引.res将包含操作的结果:一个类型为*esapi.Response的响应对象*/
	// bytes.NewReader函数创建一个*bytes.Reader，它实现了io.Reader接口
	// bytes.NewReader(buf.Bytes()):将字节切片作为 读取源 传递给Bulk方法,
	// ec.client.Bulk.WithIndex(index):Bulk方法的一个选项设置,指定批量操作的目标索引
	res, err := ec.client.Bulk(bytes.NewReader(buf.Bytes()), ec.client.Bulk.WithIndex(index))
	if err != nil {
		return fmt.Errorf("error bulk indexing documents: %v", err)
	}
	defer res.Body.Close()

	log.Println("Bulk documents indexed successfully")
	return nil
}

// SearchDocumentsWithPagination 在 Elasticsearch 中执行搜索，支持分页
func (ec *ElasticClient) SearchDocumentsWithPagination(ctx context.Context, query string, from, size int) ([]map[string]interface{}, uint, error) {
	// 构造查询请求体，带分页参数
	searchQuery := fmt.Sprintf(`
	{
		"query": {
			"multi_match": {
				"query": "%s",
				"fields": ["name", "title", "info"]
			}
		},
		"from": %d,
		"size": %d
	}`, query, from, size)

	// 执行搜索请求
	res, err := ec.client.Search(
		ec.client.Search.WithContext(ctx),
		ec.client.Search.WithIndex("products"), // 假设商品数据存储在 "products" 索引中
		ec.client.Search.WithBody(strings.NewReader(searchQuery)),
		ec.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		log.Printf("Error executing search: %v", err)
		return nil, 0, err
	}
	defer res.Body.Close()

	// 解析结果
	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Printf("Error parsing the search response body: %v", err)
		return nil, 0, err
	}

	// 提取 hits 和 total 数量
	hits := r["hits"].(map[string]interface{})["hits"].([]interface{})
	total := uint(r["hits"].(map[string]interface{})["total"].(map[string]interface{})["value"].(float64))

	// 将 hits 解析为结果集
	var results []map[string]interface{}
	for _, hit := range hits {
		source := hit.(map[string]interface{})["_source"].(map[string]interface{})
		results = append(results, source)
	}

	return results, total, nil
}

// SearchDocument 方法用于在指定索引中搜索文档
func (e *ElasticClient) SearchDocument(index, query string) (map[string]interface{}, error) {
	// 构造查询请求体
	searchQuery := fmt.Sprintf(`{
        "query": {
            "match": {
                "name": "%s"
            }
        }
    }`, query)

	fmt.Printf("Search query: %s\n", searchQuery) // 打印查询

	// 执行查询
	res, err := e.client.Search(
		e.client.Search.WithContext(context.Background()),
		e.client.Search.WithIndex(index),
		e.client.Search.WithBody(strings.NewReader(searchQuery)),
		e.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("error executing search: %v", err)
	}
	defer res.Body.Close()

	// 打印 Elasticsearch 响应内容
	body, _ := io.ReadAll(res.Body)
	fmt.Printf("Elasticsearch response: %s\n", string(body))

	// 解析结果
	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("error parsing response body: %v", err)
	}

	// 检查 hits 和返回第一个找到的文档
	hits := r["hits"].(map[string]interface{})["hits"].([]interface{})
	if len(hits) == 0 {
		return nil, fmt.Errorf("no document found")
	}

	// 返回第一个文档的 _source 数据
	source := hits[0].(map[string]interface{})["_source"].(map[string]interface{})
	return source, nil
}
