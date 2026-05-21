// package es 提供 Elasticsearch 客户端的通用封装，search-svc 使用。
package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
)

// Client 封装 ES 客户端和相关配置。
type Client struct {
	cli   *elasticsearch.Client
	index string // ES 索引名称
}

// VideoDoc 存储在 ES 中的视频文档结构。
type VideoDoc struct {
	VideoId     int64  `json:"video_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	UserId      int64  `json:"user_id"`
	Username    string `json:"username"`
	CoverUrl    string `json:"cover_url"`
	PlayCount   int64  `json:"play_count"`
	LikeCount   int64  `json:"like_count"`
	Duration    int32  `json:"duration"`
	CreatedAt   int64  `json:"created_at"`
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

// SearchVideos 全文搜索视频（multi_match title + description）。
func (c *Client) SearchVideos(ctx context.Context, keyword string, category string, page, size int) (*SearchResult, error) {
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

	var result struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source VideoDoc `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	sr := &SearchResult{Total: result.Hits.Total.Value}
	for _, h := range result.Hits.Hits {
		doc := h.Source
		sr.Hits = append(sr.Hits, &doc)
	}
	return sr, nil
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

// EnsureIndex 确保索引存在，不存在则创建。
func (c *Client) EnsureIndex(ctx context.Context) error {
	res, err := c.cli.Indices.Exists([]string{c.index})
	if err != nil {
		return err
	}
	res.Body.Close()

	if res.StatusCode == 404 {
		res, err = c.cli.Indices.Create(c.index)
		if err != nil {
			return err
		}
		res.Body.Close()
	}
	return nil
}
