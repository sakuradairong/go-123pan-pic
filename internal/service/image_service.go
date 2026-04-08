package service

import (
	"fmt"
	"strings"

	"imagehost/internal/config"
	"imagehost/internal/pan123"
)

type ImageService struct {
	client *pan123.Client
}

func NewImageService(client *pan123.Client) *ImageService {
	return &ImageService{client: client}
}

// GetImageItems 获取图床列表格式化信息
func (s *ImageService) GetImageItems() ([]pan123.FileItem, error) {
	// 获取前 100 张即可
	items, _, err := s.client.GetFileList(config.GlobalConfig.ParentFileID, 100, "")
	if err != nil {
		return nil, fmt.Errorf("获取列表异常: %w", err)
	}

	domain := strings.TrimRight(config.GlobalConfig.CustomDomain, "/")

	// 格式化清理，如果有设置自定义解析，则将直链全部附魔为指定域名
	for i := range items {
		// 如果 123pan 源生返回了 userSelfURL，可以优先。如果没有且我们配了 custom_domain，尝试通过拼接提供后备访问链
		if items[i].UserSelfURL == "" && domain != "" {
			// 一些简易替换，具体规则需按你在 123pan 的直链白名单空间里设置来调整
			items[i].UserSelfURL = fmt.Sprintf("%s/%s", domain, items[i].Filename)
		}
	}

	return items, nil
}

// DeleteImages 删除单张或多张
func (s *ImageService) DeleteImages(ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return s.client.DeleteFiles(ids)
}
