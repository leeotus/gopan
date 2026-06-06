// package es 提供 Elasticsearch 客户端的通用封装，search-svc 使用。
package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
)

// Client 封装 ES 客户端和相关配置。
type Client struct {
	cli   *elasticsearch.Client
	index string // ES 索引名称
}

// VideoDoc 存储在 ES 中的视频文档结构。
type VideoDoc struct {
	VideoId     int64     `json:"video_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	UserId      int64     `json:"user_id"`
	Username    string    `json:"username"`
	CoverUrl    string    `json:"cover_url"`
	PlayCount   int64     `json:"play_count"`
	LikeCount   int64     `json:"like_count"`
	Duration    int32     `json:"duration"`
	CreatedAt   int64     `json:"created_at"`
	VideoVector []float32 `json:"video_vector,omitempty"` // 512 维特征向量，用于多模态语义检索
}

func NewClient(addresses []string, index, username, password string) (*Client, error) {
	cfg := elasticsearch.Config{
		Addresses: addresses,
	}
	if username != "" {
		cfg.Username = username
		cfg.Password = password
	}
	cli, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("es connect: %w", err)
	}

	// 检查连接
	res, err := cli.Ping()
	if err != nil {
		return nil, fmt.Errorf("es ping: %w", err)
	}
	res.Body.Close()

	return &Client{cli: cli, index: index}, nil
}

// IndexVideo 创建或更新一条视频索引。
func (c *Client) IndexVideo(ctx context.Context, doc *VideoDoc) error {
	// 【AI 智能注入】如果文档尚未包含特征向量且 AI 服务在线，自动提取“Title + Description”联合语义，注入 512 维稠密特征向量
	if len(doc.VideoVector) == 0 {
		textToEmbed := doc.Title
		if doc.Description != "" {
			textToEmbed = fmt.Sprintf("%s %s", doc.Title, doc.Description)
		}
		vec, err := getEmbeddingVector(ctx, textToEmbed)
		if err == nil && len(vec) == 512 {
			doc.VideoVector = vec
			fmt.Printf("[AI Auto-Index] Generated 512-dim embedding from title & desc for video: %d\n", doc.VideoId)
		} else {
			fmt.Printf("[AI Auto-Index Warning] Skip vectorizing video: %d, err: %v\n", doc.VideoId, err)
		}
	}

	body, _ := json.Marshal(doc)
	res, err := c.cli.Index(
		c.index,
		bytes.NewReader(body),
		c.cli.Index.WithContext(ctx),
		c.cli.Index.WithDocumentID(fmt.Sprintf("%d", doc.VideoId)),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return fmt.Errorf("es index error: %s", res.String())
	}
	return nil
}

// SearchResult ES 搜索结果。
type SearchResult struct {
	Total int64
	Hits  []*VideoDoc
}

// SearchVideos 混合/降级检索大网关：
// 1. 尝试通过 HTTP 调用本地的 semantic-ai 向量解码服务将 text 变为 512 浮点表示。
// 2. 如果成功拿到向量，则走 ES 8.x 原生高阶 K-NN 语义匹配机制。
// 3. 如果 AI 服务不通/故障等，无锁优雅降级回传统的文本分词 BM25 全文检索 (multi_match)，保障系统绝不断服。
func (c *Client) SearchVideos(ctx context.Context, keyword string, category string, page, size int) (*SearchResult, error) {
	// 尝试向本地 AI 节点提交特征请求
	vec, aiErr := getEmbeddingVector(ctx, keyword)
	if aiErr == nil && len(vec) == 512 {
		fmt.Printf("[AI Search OK] Performing Vector k-NN Search for: '%s'\n", keyword)
		res, err := c.searchVideosByKNN(ctx, vec, category, page, size)
		if err == nil {
			return res, nil
		}
		// KNN 检索出错，降级到传统 BM25 词频检索（保留原始关键词）
		fmt.Printf("[KNN Fallback] KNN search failed (%v), falling back to lexical search for: '%s'\n", err, keyword)
	}

	// 触发安全降级
	fmt.Printf("[AI Search Fallback] AI Offline (%v). Performing classical BM25 Lucene Search for: '%s'\n", aiErr, keyword)
	return c.searchVideosByLexical(ctx, keyword, category, page, size)
}

// searchVideosByKNN 进行向量相空间语义近邻匹配
func (c *Client) searchVideosByKNN(ctx context.Context, vector []float32, category string, page, size int) (*SearchResult, error) {
	from := (page - 1) * size
	if from < 0 {
		from = 0
	}

	// ES 8.x 官方原生 k-NN 近邻检索模型配置
	query := map[string]any{
		"knn": map[string]any{
			"field":          "video_vector",
			"query_vector":   vector,
			"k":              size,
			"num_candidates": 50,
		},
		"from": from,
		"size": size,
	}

	// 支持混合式分类硬筛选
	if category != "" {
		query["knn"].(map[string]any)["filter"] = map[string]any{
			"term": map[string]any{"category": category},
		}
	}

	body, _ := json.Marshal(query)
	res, err := c.cli.Search(
		c.cli.Search.WithContext(ctx),
		c.cli.Search.WithIndex(c.index),
		c.cli.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		res.Body.Close()
		return nil, fmt.Errorf("es knn search error: %s", string(bodyBytes))
	}

	return parseSearchResponse(res.Body)
}

// searchVideosByLexical 传统倒排词频检索
func (c *Client) searchVideosByLexical(ctx context.Context, keyword string, category string, page, size int) (*SearchResult, error) {
	from := (page - 1) * size
	if from < 0 {
		from = 0
	}

	query := map[string]any{
		"query": map[string]any{
			"bool": map[string]any{
				"must": []map[string]any{
					{
						"multi_match": map[string]any{
							"query":  keyword,
							"fields": []string{"title^3", "description"},
						},
					},
				},
			},
		},
		"from": from,
		"size": size,
		"sort": []map[string]any{
			{"_score": map[string]string{"order": "desc"}},
		},
	}

	if category != "" {
		must := query["query"].(map[string]any)["bool"].(map[string]any)["must"].([]map[string]any)
		must = append(must, map[string]any{
			"term": map[string]any{"category": category},
		})
		query["query"].(map[string]any)["bool"].(map[string]any)["must"] = must
	}

	body, _ := json.Marshal(query)
	res, err := c.cli.Search(
		c.cli.Search.WithContext(ctx),
		c.cli.Search.WithIndex(c.index),
		c.cli.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("es search error: %s", res.String())
	}

	return parseSearchResponse(res.Body)
}

// RemoveVideo 从 ES 索引中删除一条视频文档。
func (c *Client) RemoveVideo(ctx context.Context, videoId int64) error {
	res, err := c.cli.Delete(
		c.index,
		fmt.Sprintf("%d", videoId),
		c.cli.Delete.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

// EnsureIndex 确保索引存在，不存在则初始化并挂载 DENSE_VECTOR 密集向量图。
// 如果索引已存在但缺少 video_vector 的 dense_vector 映射，则自动补全。
func (c *Client) EnsureIndex(ctx context.Context) error {
	res, err := c.cli.Indices.Exists([]string{c.index})
	if err != nil {
		return err
	}
	res.Body.Close()

	if res.StatusCode == 404 {
		// 【终极重构】创建索引时直接预制 DENSE_VECTOR 特性。512通道，Cosine 计算，自动生成内聚索引图，无缝兼容 OpenAI 与 CLIP 特征。
		mapping := `{
			"mappings": {
				"properties": {
					"video_id":     {"type": "long"},
					"title":        {"type": "text", "analyzer": "standard"},
					"description":  {"type": "text", "analyzer": "standard"},
					"category":     {"type": "keyword"},
					"user_id":      {"type": "long"},
					"username":     {"type": "keyword"},
					"cover_url":    {"type": "keyword"},
					"play_count":   {"type": "long"},
					"like_count":   {"type": "long"},
					"duration":     {"type": "integer"},
					"created_at":   {"type": "long"},
					"video_vector": {
						"type": "dense_vector",
						"dims": 512,
						"index": true,
						"similarity": "cosine"
					}
				}
			}
		}`
		res, err = c.cli.Indices.Create(
			c.index,
			c.cli.Indices.Create.WithContext(ctx),
			c.cli.Indices.Create.WithBody(bytes.NewReader([]byte(mapping))),
		)
		if err != nil {
			return err
		}
		res.Body.Close()
		fmt.Println("[ES Schema Patch] Index 'gopan_videos' successfully created with 512-dim Dense Vector capability!")
	} else {
		// 索引已存在，检查并补全 video_vector 的 dense_vector 映射
		// 如果索引创建时没有 dense_vector 映射，KNN 查询会失败
		getMapping, err := c.cli.Indices.GetMapping(
			c.cli.Indices.GetMapping.WithIndex(c.index),
		)
		if err != nil {
			return err
		}
		defer getMapping.Body.Close()

		var mappingResult map[string]any
		if err := json.NewDecoder(getMapping.Body).Decode(&mappingResult); err != nil {
			return err
		}

		// 检查 video_vector 字段是否已存在且为 dense_vector 类型
		indexMapping, ok := mappingResult[c.index].(map[string]any)
		if !ok {
			return nil
		}
		mappings, ok := indexMapping["mappings"].(map[string]any)
		if !ok {
			return nil
		}
		properties, ok := mappings["properties"].(map[string]any)
		if !ok {
			return nil
		}
		videoVectorField, exists := properties["video_vector"].(map[string]any)
		fieldType, _ := videoVectorField["type"].(string)

		if !exists || fieldType != "dense_vector" {
			// 补全 video_vector 的 dense_vector 映射
			putMapping := `{
				"properties": {
					"video_vector": {
						"type": "dense_vector",
						"dims": 512,
						"index": true,
						"similarity": "cosine"
					}
				}
			}`
			putRes, err := c.cli.Indices.PutMapping(
				[]string{c.index},
				bytes.NewReader([]byte(putMapping)),
				c.cli.Indices.PutMapping.WithContext(ctx),
			)
			if err != nil {
				return err
			}
			putRes.Body.Close()
			if putRes.IsError() {
				return fmt.Errorf("es put mapping error: %s", putRes.String())
			}
			fmt.Println("[ES Schema Patch] Added 'video_vector' dense_vector mapping to existing index 'gopan_videos'!")
		}
	}
	return nil
}

// getEmbeddingVector 发送 HTTP 转换向量 (对 9900/GPU 和 9901/CPU 做弹性降级双层发现)
func getEmbeddingVector(ctx context.Context, keyword string) ([]float32, error) {
	aiURL := os.Getenv("AI_SERVICE_URL")
	if aiURL == "" {
		aiURL = "http://127.0.0.1:9900" // 默认优先走搭载 GPU 的 9900
	}

	payload, _ := json.Marshal(map[string]string{"text": keyword})
	client := &http.Client{Timeout: 3 * time.Second}

	// 1. 发射第一次尝试请求
	req, err := http.NewRequestWithContext(ctx, "POST", aiURL+"/embed/text", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)

	// 2. 如果发生网络错，且用的是默认 9900，优雅转向 9901（CPU 本地测试端）
	if err != nil && aiURL == "http://127.0.0.1:9900" {
		req2, retryErr := http.NewRequestWithContext(ctx, "POST", "http://127.0.0.1:9901/embed/text", bytes.NewReader(payload))
		if retryErr == nil {
			req2.Header.Set("Content-Type", "application/json")
			resp, err = client.Do(req2)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("ai-service offline: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ai-service return status %d", resp.StatusCode)
	}

	var parsed struct {
		Dimension int       `json:"dimension"`
		Vector    []float32 `json:"vector"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	return parsed.Vector, nil
}

// 解析统一查询返回值
func parseSearchResponse(body io.ReadCloser) (*SearchResult, error) {
	var result struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Score  float64  `json:"_score"`
				Source VideoDoc `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return nil, err
	}

	sr := &SearchResult{Total: result.Hits.Total.Value}
	for _, h := range result.Hits.Hits {
		doc := h.Source
		sr.Hits = append(sr.Hits, &doc)
	}
	return sr, nil
}
