package service

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"strings"
	"time"

	"imagehost/internal/config"
	"imagehost/internal/pan123"
)

type UploadService struct {
	client *pan123.Client
}

func NewUploadService(client *pan123.Client) *UploadService {
	return &UploadService{client: client}
}

// UploadFile 处理单一文件的五步上传编排
func (s *UploadService) UploadFile(fileHeader *multipart.FileHeader) (pan123.FileItem, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return pan123.FileItem{}, fmt.Errorf("打开分段文件失败: %w", err)
	}
	defer file.Close()

	// 【致命漏洞修复】使用流式处理替换无痛但高危的被动全量加载 (ReadAll)
	h := md5.New()
	if _, err := io.Copy(h, file); err != nil {
		return pan123.FileItem{}, fmt.Errorf("动态计算文件 MD5 失败: %w", err)
	}
	etag := fmt.Sprintf("%x", h.Sum(nil))
	size := fileHeader.Size

	// 重置读取游标，准备让底层的 123pan.Client 直传该 file IO 节点
	file.Seek(0, io.SeekStart)
	fileName := fileHeader.Filename

	var fileID string

	// 步骤 1: 创建文件申报
	createResp, err := s.client.CreateFile(config.GlobalConfig.ParentFileID, fileName, etag, size)
	if err != nil {
		return pan123.FileItem{}, fmt.Errorf("123pan 发送申报异常: %w", err)
	}

	if createResp.Reuse {
		// 秒传命中
		fileID = createResp.FileID
		log.Printf("[上传] 文件 %s 命中服务器秒传!", fileName)
	} else {
		// 没有秒传，需要物理 PUT 上传切片
		preuploadID := createResp.PreuploadID

		// 步骤 2: 获取签发 URL
		urlResp, err := s.client.GetUploadURL(preuploadID, 1)
		if err != nil {
			return pan123.FileItem{}, fmt.Errorf("123pan 获取直传地址失败: %w", err)
		}

		log.Printf("[上传] 开始推送 %s 到 123pan (大小: %d bytes)...", fileName, size)
		// 步骤 3: 真正的二进制推送 (此处直接透传管道 file，坚决不在 RAM 内过境缓存！)
		err = s.client.DoRawPUT(urlResp.PresignedURL, file, size)
		if err != nil {
			return pan123.FileItem{}, fmt.Errorf("123pan 数据切片推送失败: %w", err)
		}

		// 步骤 4: 推送完毕申报
		completeResp, err := s.client.UploadComplete(preuploadID)
		if err != nil {
			return pan123.FileItem{}, fmt.Errorf("123pan 汇报上传完结失败: %w", err)
		}

		// 判决处理策略
		if completeResp.Completed {
			fileID = completeResp.FileID
		} else if completeResp.Async {
			// 步骤 5: 异步轮询 (最大容忍 20s)
			log.Printf("[上传] 123pan 正在后端异步处理合并 %s...", fileName)
			for i := 0; i < 20; i++ {
				time.Sleep(1 * time.Second)
				asyncResp, err := s.client.CheckAsyncResult(preuploadID)
				if err != nil {
					log.Printf("轮询报错 (将继续尝试): %v", err)
					continue
				}
				if asyncResp.Completed {
					fileID = asyncResp.FileID
					break
				}
			}
			if fileID == "" {
				return pan123.FileItem{}, fmt.Errorf("123pan 异步处理落盘超时")
			}
		} else {
			return pan123.FileItem{}, fmt.Errorf("123pan 返回的完结状态不能被理解: %v", completeResp)
		}
	}

	log.Printf("[上传] 文件 %s 整体保存完毕, ID: %s", fileName, fileID)

	// 根据图床需求，组装格式化的信息，此时可返回。
	// 这里通过自定义协议拼接URL假若没有详情API能立即获取。（或者使用 FileID 查询）
	downloadURL := ""
	if config.GlobalConfig.CustomDomain != "" {
		// 假设其静态资源访问域名形式为 domain/fileID/filename 或者是普通形式。
		// 大部分直链空间只需： customDomain / fileName / file_id
		// 作为通用兼容处理，如果直链域名未配特定的子径，先假设能直接根据路径拼接，具体的按实际调整。
		domain := strings.TrimRight(config.GlobalConfig.CustomDomain, "/")
		downloadURL = fmt.Sprintf("%s/%s", domain, fileName)
	}

	return pan123.FileItem{
		FileID:      fileID,
		Filename:    fileName,
		Size:        size,
		Etag:        etag,
		DownloadURL: downloadURL,
		UserSelfURL: downloadURL, // 我们以此值为向外暴露的主要直链URL
	}, nil
}
