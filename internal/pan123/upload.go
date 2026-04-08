package pan123

import (
	"encoding/json"
	"fmt"
)

// CreateFile 步骤1: 创建文件/发卷预申报。123pan 会下发预上传 ID（preuploadID）或判定直接秒传。
func (c *Client) CreateFile(parentFileID, filename, etag string, size int64) (*CreateFileRespData, error) {
	reqBody := CreateFileReq{
		ParentFileID: parentFileID,
		Filename:     filename,
		Etag:         etag,
		Size:         size,
		Type:         1,
	}

	var resp BaseResp
	// 上传接口大部分挂在 UploadBaseURL + /upload/v1/xxx
	err := c.DoJSONRequest("POST", UploadBaseURL+"/upload/v1/oss/file/create", reqBody, &resp)
	if err != nil {
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

// GetUploadURL 步骤2: 凭借传入的 preuploadID 及 切片编号(对于整体直接设为1即可) 换取可以真正上传二进制的预签名 URL。
func (c *Client) GetUploadURL(preuploadID string, sliceNo int) (*GetUploadURLRespData, error) {
	reqBody := GetUploadURLReq{
		PreuploadID: preuploadID,
		SliceNo:     sliceNo,
	}

	var resp BaseResp
	err := c.DoJSONRequest("POST", UploadBaseURL+"/upload/v1/oss/file/get_upload_url", reqBody, &resp)
	if err != nil {
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

// UploadComplete 步骤4 (由于步骤3是调用泛化的 DoRawPUT 进行的纯数据上传): 当步骤3全部走完时，调用此 API 宣告物理上传完毕
func (c *Client) UploadComplete(preuploadID string) (*UploadCompleteRespData, error) {
	reqBody := UploadCompleteReq{
		PreuploadID: preuploadID,
	}

	var resp BaseResp
	err := c.DoJSONRequest("POST", UploadBaseURL+"/upload/v1/oss/file/upload_complete", reqBody, &resp)
	if err != nil {
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

// CheckAsyncResult 步骤5 (如需): 若上一步返回的数据指出 async == true 且 completed == false，那么需轮询此 API
func (c *Client) CheckAsyncResult(preuploadID string) (*UploadAsyncResultRespData, error) {
	reqBody := UploadAsyncResultReq{
		PreuploadID: preuploadID,
	}

	var resp BaseResp
	err := c.DoJSONRequest("POST", UploadBaseURL+"/upload/v1/oss/file/upload_async_result", reqBody, &resp)
	if err != nil {
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
