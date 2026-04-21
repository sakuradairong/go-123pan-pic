package pan123

import (
	"encoding/json"
	"fmt"
)

func (c *Client) GetFileList(parentFileID string, limit int, lastFileID string) ([]FileItem, string, error) {
	reqBody := FileListReq{
		ParentFileID: parentFileID,
		Limit:        limit,
		LastFileID:   lastFileID,
		Type:         1,
	}

	var resp BaseResp
	if err := c.DoJSONRequest("POST", APIBaseURL+"/api/v1/oss/file/list", reqBody, &resp); err != nil {
		return nil, "", err
	}
	if resp.Code != 0 {
		return nil, "", fmt.Errorf("123pan API error (list files), code: %d, msg: %s", resp.Code, resp.Message)
	}

	var data FileListRespData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, "", fmt.Errorf("decode file list data error: %w", err)
	}

	return data.FileList, data.LastFileID, nil
}

func (c *Client) DeleteFiles(fileIDs []string) error {
	reqBody := DeleteFileReq{FileIDs: fileIDs}

	var resp BaseResp
	if err := c.DoJSONRequest("POST", APIBaseURL+"/api/v1/oss/file/delete", reqBody, &resp); err != nil {
		return err
	}
	if resp.Code != 0 {
		return fmt.Errorf("123pan API error (delete file), code: %d, msg: %s", resp.Code, resp.Message)
	}

	return nil
}
