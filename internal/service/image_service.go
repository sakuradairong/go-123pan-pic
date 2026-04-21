package service

import (
	"fmt"
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
		customDomain: strings.TrimRight(customDomain, "/"),
	}
}

func (s *ImageService) GetImageItems() ([]pan123.FileItem, error) {
	items, _, err := s.client.GetFileList(s.parentFileID, 100, "")
	if err != nil {
		return nil, fmt.Errorf("获取列表失败: %w", err)
	}

	for i := range items {
		if items[i].UserSelfURL == "" && s.customDomain != "" {
			items[i].UserSelfURL = fmt.Sprintf("%s/%s", s.customDomain, items[i].Filename)
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
