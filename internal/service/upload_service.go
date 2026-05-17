package service

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"strings"
	"time"

	"imagehost/internal/pan123"
)

type UploadService struct {
	client       *pan123.Client
	parentFileID string
	customDomain string
}

func NewUploadService(client *pan123.Client, parentFileID, customDomain string) *UploadService {
	return &UploadService{
		client:       client,
		parentFileID: parentFileID,
		customDomain: strings.TrimSuffix(customDomain, "/"),
	}
}

func (s *UploadService) UploadFile(fileHeader *multipart.FileHeader) (pan123.FileItem, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return pan123.FileItem{}, fmt.Errorf("打开分段文件失败: %w", err)
	}
	defer file.Close()

	h := md5.New()
	if _, err := io.Copy(h, file); err != nil {
		return pan123.FileItem{}, fmt.Errorf("计算文件 MD5 失败: %w", err)
	}
	etag := fmt.Sprintf("%x", h.Sum(nil))
	size := fileHeader.Size

	file.Seek(0, io.SeekStart)
	fileName := fileHeader.Filename

	var fileID string

	createResp, err := s.client.CreateFile(s.parentFileID, fileName, etag, size)
	if err != nil {
		return pan123.FileItem{}, fmt.Errorf("123pan 创建文件失败: %w", err)
	}

	if createResp.Reuse {
		fileID = createResp.FileID
		log.Printf("[上传] 文件 %s 命中秒传", fileName)
	} else {
		preuploadID := createResp.PreuploadID
		sliceSize := createResp.SliceSize

		if sliceSize == 0 || size <= sliceSize {
			// 单分片上传
			urlResp, err := s.client.GetUploadURL(preuploadID, 1)
			if err != nil {
				return pan123.FileItem{}, fmt.Errorf("123pan 获取上传地址失败: %w", err)
			}
			log.Printf("[上传] 推送 %s 到 123pan (%d bytes)...", fileName, size)
			if err = s.client.DoRawPUT(urlResp.PresignedURL, file, size); err != nil {
				return pan123.FileItem{}, fmt.Errorf("123pan 数据上传失败: %w", err)
			}
		} else {
			// 多分片上传
			numSlices := int((size + sliceSize - 1) / sliceSize)
			log.Printf("[上传] %s 分 %d 片上传播 (sliceSize=%d)", fileName, numSlices, sliceSize)
			for i := 0; i < numSlices; i++ {
				sliceNo := i + 1
				offset := int64(i) * sliceSize
				readSize := sliceSize
				if offset+readSize > size {
					readSize = size - offset
				}

				urlResp, err := s.client.GetUploadURL(preuploadID, sliceNo)
				if err != nil {
					return pan123.FileItem{}, fmt.Errorf("123pan 获取第%d片上传达址失败: %w", sliceNo, err)
				}

				sectionReader := io.NewSectionReader(file, offset, readSize)
				log.Printf("[上传] %s 第%d片 (%d bytes offset=%d)...", fileName, sliceNo, readSize, offset)
				if err = s.client.DoRawPUT(urlResp.PresignedURL, sectionReader, readSize); err != nil {
					return pan123.FileItem{}, fmt.Errorf("123pan 第%d片数据上传失败: %w", sliceNo, err)
				}
			}
		}

		completeResp, err := s.client.UploadComplete(preuploadID)
		if err != nil {
			return pan123.FileItem{}, fmt.Errorf("123pan 上传完结通知失败: %w", err)
		}

		if completeResp.Completed {
			fileID = completeResp.FileID
		} else if completeResp.Async {
			log.Printf("[上传] 123pan 异步处理 %s...", fileName)
			for i := 0; i < 20; i++ {
				time.Sleep(1 * time.Second)
				asyncResp, err := s.client.CheckAsyncResult(preuploadID)
				if err != nil {
					log.Printf("[上传] 轮询出错 (继续重试): %v", err)
					continue
				}
				if asyncResp.Completed {
					fileID = asyncResp.FileID
					break
				}
			}
			if fileID == "" {
				return pan123.FileItem{}, fmt.Errorf("123pan 异步处理超时")
			}
		} else {
			return pan123.FileItem{}, fmt.Errorf("123pan 返回未知完结状态: %v", completeResp)
		}
	}

	log.Printf("[上传] %s 完成, ID: %s", fileName, fileID)

	downloadURL := ""
	if s.customDomain != "" {
		downloadURL = fmt.Sprintf("%s/%s", s.customDomain, fileName)
	}

	return pan123.FileItem{
		FileID:      fileID,
		Filename:    fileName,
		Size:        size,
		Etag:        etag,
		DownloadURL: downloadURL,
		UserSelfURL: downloadURL,
	}, nil
}
