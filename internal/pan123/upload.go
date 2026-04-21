package pan123

import (
	"encoding/json"
	"fmt"
)

func (c *Client) CreateFile(parentFileID, filename, etag string, size int64) (*CreateFileRespData, error) {
	reqBody := CreateFileReq{
		ParentFileID: parentFileID,
		Filename:     filename,
		Etag:         etag,
		Size:         size,
		Type:         1,
	}

	var resp BaseResp
	if err := c.DoJSONRequest("POST", APIBaseURL+"/upload/v1/oss/file/create", reqBody, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("123pan API error (create file), code: %d, msg: %s", resp.Code, resp.Message)
	}

	var data CreateFileRespData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("decode create_file data error: %w", err)
	}
	return &data, nil
}

func (c *Client) GetUploadURL(preuploadID string, sliceNo int) (*GetUploadURLRespData, error) {
	reqBody := GetUploadURLReq{
		PreuploadID: preuploadID,
		SliceNo:     sliceNo,
	}

	var resp BaseResp
	if err := c.DoJSONRequest("POST", APIBaseURL+"/upload/v1/oss/file/get_upload_url", reqBody, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("123pan API error (get upload url), code: %d, msg: %s", resp.Code, resp.Message)
	}

	var data GetUploadURLRespData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("decode get_upload_url data error: %w", err)
	}
	return &data, nil
}

func (c *Client) UploadComplete(preuploadID string) (*UploadCompleteRespData, error) {
	reqBody := UploadCompleteReq{PreuploadID: preuploadID}

	var resp BaseResp
	if err := c.DoJSONRequest("POST", APIBaseURL+"/upload/v1/oss/file/upload_complete", reqBody, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("123pan API error (upload complete), code: %d, msg: %s", resp.Code, resp.Message)
	}

	var data UploadCompleteRespData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("decode upload_complete data error: %w", err)
	}
	return &data, nil
}

func (c *Client) CheckAsyncResult(preuploadID string) (*UploadAsyncResultRespData, error) {
	reqBody := UploadAsyncResultReq{PreuploadID: preuploadID}

	var resp BaseResp
	if err := c.DoJSONRequest("POST", APIBaseURL+"/upload/v1/oss/file/upload_async_result", reqBody, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 0 {
		return nil, fmt.Errorf("123pan API error (async result query), code: %d, msg: %s", resp.Code, resp.Message)
	}

	var data UploadAsyncResultRespData
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("decode async_result data error: %w", err)
	}
	return &data, nil
}
