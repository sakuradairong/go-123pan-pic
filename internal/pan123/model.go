package pan123

import "encoding/json"

// BaseResp 123pan API 统一响应的包裹格式
type BaseResp struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"` // 延迟解析真实的业务数据
}

// ----------- 核心：图床上传五步相关 -----------

// CreateFileReq 步骤1：创建文件申报
type CreateFileReq struct {
	ParentFileID string `json:"parentFileID"` // 必填，填图床专属目录
	Filename     string `json:"filename"`
	Etag         string `json:"etag"`         // 必须填文件内容的真实 MD5
	Size         int64  `json:"size"`         // 文件大小
	Type         int    `json:"type"`         // 固定填 1
}

type CreateFileRespData struct {
	FileID      string `json:"fileID"`
	Reuse       bool   `json:"reuse"`       // 为 true 的话说明发生秒传，直接上传成功
	PreuploadID string `json:"preuploadID"` // 若需上传分配的 ID
	SliceSize   int64  `json:"sliceSize"`   // 要求分块的尺寸
}

// GetUploadURLReq 步骤2：获取真正可以上传分片的直链 URL
type GetUploadURLReq struct {
	PreuploadID string `json:"preuploadID"`
	SliceNo     int    `json:"sliceNo"` // 从 1 开始一直自增
}

type GetUploadURLRespData struct {
	PresignedURL string `json:"presignedURL"` // 这个 URL 需要通过纯 PUT 来访问
	IsMultipart  bool   `json:"isMultipart"`
}

// UploadCompleteReq 步骤4：文件所有切片 PUT 完成后，报告已传完
type UploadCompleteReq struct {
	PreuploadID string `json:"preuploadID"`
}

type UploadCompleteRespData struct {
	FileID    string `json:"fileID"`
	Async     bool   `json:"async"` // 如果是 true，说明它需要排队处理，我们得进行步骤5轮询
	Completed bool   `json:"completed"`
}

// UploadAsyncResultReq 步骤5：轮询是否最终真正落盘完毕
type UploadAsyncResultReq struct {
	PreuploadID string `json:"preuploadID"`
}

type UploadAsyncResultRespData struct {
	Completed bool   `json:"completed"`
	FileID    string `json:"fileID"`
}

// ----------- 图床查询与管理相关 -----------

// FileListReq 取文件列表
type FileListReq struct {
	ParentFileID string `json:"parentFileID"`          // 请求字段统一使用 parentFileID（大写 D）
	Limit        int    `json:"limit"`                 // 最大不能超过 100
	LastFileID   string `json:"lastFileID,omitempty"`  // 分页游标
	Type         int    `json:"type"`                  // 固定填 1
}

type FileItem struct {
	FileID      string `json:"fileId"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	Etag        string `json:"etag"`
	Status      int    `json:"status"`
	CreateAt    string `json:"createAt"`
	DownloadURL string `json:"downloadURL"` // 防盗链 URL 
	UserSelfURL string `json:"userSelfURL"` // 自定义域名下发 URL
}

type FileListRespData struct {
	LastFileID string     `json:"lastFileId"` // 若为 -1 代表没有下一页
	FileList   []FileItem `json:"fileList"`
}

// DeleteFileReq 删除指定的文件（可以是批量）
type DeleteFileReq struct {
	FileIDs []string `json:"fileIDs"` // 数组不能超过 100 长度限制
}

// ----------- 鉴权相关 -----------

// AccessTokenReq 获取凭证请求
type AccessTokenReq struct {
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
}

// AccessTokenRespData 获取凭证的回复 Data
type AccessTokenRespData struct {
	AccessToken string `json:"accessToken"`
	ExpiredAt   string `json:"expiredAt"`
}
