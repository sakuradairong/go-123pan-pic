package pan123

import (
	"encoding/json"
	"fmt"
)

// GetFileList 获取某个目录下面的（图片）文件列表
func (c *Client) GetFileList(parentFileID string, limit int, lastFileID string) ([]FileItem, string, error) {
	reqBody := FileListReq{
		ParentFileID: parentFileID,
		Limit:        limit,
		LastFileID:   lastFileID,
		Type:         1,
	}

	var resp BaseResp
	// 注意这里文档使用的是 APIBaseURL + /api/v1/oss/file/list
	err := c.DoJSONRequest("POST", APIBaseURL+"/api/v1/oss/file/list", reqBody, &resp)
	if err != nil {
		return nil, "", err
	}

	if resp.Code != 0 {
		return nil, "", fmt.Errorf("123pan API error (list files), code: %d, msg: %s", resp.Code, resp.Message)
	}

	var data FileListRespData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, "", fmt.Errorf("fail to decode file list data: %w", err)
	}

	return data.FileList, data.LastFileID, nil
}

// DeleteFiles 根据 FileIDs 数组批量/单点删除远程文件
func (c *Client) DeleteFiles(fileIDs []string) error {
	reqBody := DeleteFileReq{
		FileIDs: fileIDs,
	}

	var resp BaseResp
	// 删除接口使用的是 /api/v1/oss/file/delete
	err := c.DoJSONRequest("POST", APIBaseURL+"/api/v1/oss/file/delete", reqBody, &resp)
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return fmt.Errorf("123pan API error (delete file), code: %d, msg: %s", resp.Code, resp.Message)
	}

	return nil
}
