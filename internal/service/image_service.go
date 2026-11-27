package service

import (
	"fmt"
	"net/url"
	"strings"

	"imagehost/internal/pan123"
)

type ImageService struct {
	client       *pan123.Client
	parentFileID string
	customDomain string
}

func NewImageService(client *pan123.Client, parentFileID, customDomain string) *ImageService {
	return &ImageService{
		client:       client,
		parentFileID: parentFileID,
		customDomain: strings.TrimSuffix(customDomain, "/"),
	}
}

func (s *ImageService) GetImageItems() ([]pan123.FileItem, error) {
	var items []pan123.FileItem
	seen := make(map[string]bool)
	for cursor := ""; ; {
		page, next, err := s.client.GetFileList(s.parentFileID, 100, cursor)
		if err != nil {
			return nil, fmt.Errorf("获取列表失败: %w", err)
		}
		items = append(items, page...)
		if next == "" || next == "-1" {
			break
		}
		if seen[next] {
			return nil, fmt.Errorf("获取列表失败: 重复的分页游标 %q", next)
		}
		seen[next] = true
		cursor = next
	}

	for i := range items {
		if items[i].UserSelfURL == "" && s.customDomain != "" {
			items[i].UserSelfURL = s.customDomain + "/" + url.PathEscape(items[i].Filename)
		}
	}

	return items, nil
}

func (s *ImageService) DeleteImages(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return s.client.DeleteFiles(ids)
}
