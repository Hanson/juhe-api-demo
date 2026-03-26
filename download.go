package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tidwall/gjson"
)

// CdnInfo CDN 信息（通过 /cdn/get_cdn_info 获取）
type CdnInfo struct {
	CdnDns        string `json:"cdn_dns"`        // CDN 服务器地址（加密字符串）
	Vid           string `json:"vid"`            // 用户 vid
	CorpId        string `json:"corp_id"`        // 企业ID
	ClientVersion string `json:"client_version"` // 客户端版本
}

// DownloadClient 下载客户端（用于私有化 CDN）
type DownloadClient struct {
	apiClient     *Client       // API 客户端（用于调用 /cdn/get_cdn_info）
	httpClient    *http.Client  // HTTP 客户端
	cdnInfo       *CdnInfo      // 缓存的 CDN 信息
	cdnInfoMutex  sync.RWMutex  // CDN 信息锁
	cdnInfoExpiry time.Time     // CDN 信息过期时间
	cdnURLQiWei   string        // 企微 CDN 服务地址
	cdnURLWeiXin  string        // 个微 CDN 服务地址
	protocol      ProtocolType
}

// NewDownloadClient 创建下载客户端
func NewDownloadClient(apiClient *Client, cdnURLQiWei, cdnURLWeiXin string, protocol ProtocolType) *DownloadClient {
	return &DownloadClient{
		apiClient:    apiClient,
		cdnURLQiWei:  cdnURLQiWei,
		cdnURLWeiXin: cdnURLWeiXin,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
		protocol: protocol,
	}
}

// GetCdnInfo 获取 CDN 信息（带缓存）
func (d *DownloadClient) GetCdnInfo() (*CdnInfo, error) {
	d.cdnInfoMutex.RLock()
	// 如果缓存存在且未过期（3小时有效期）
	if d.cdnInfo != nil && time.Now().Before(d.cdnInfoExpiry) {
		defer d.cdnInfoMutex.RUnlock()
		return d.cdnInfo, nil
	}
	d.cdnInfoMutex.RUnlock()

	d.cdnInfoMutex.Lock()
	defer d.cdnInfoMutex.Unlock()

	// 双重检查
	if d.cdnInfo != nil && time.Now().Before(d.cdnInfoExpiry) {
		return d.cdnInfo, nil
	}

	// 调用 /cdn/get_cdn_info 获取 CDN 信息
	resp, err := d.apiClient.GuidRequest(nil, "/cdn/get_cdn_info", map[string]interface{}{})
	if err != nil {
		return nil, fmt.Errorf("获取CDN信息失败: %w", err)
	}

	// 解析响应
	errCode := gjson.Get(resp, "error_code").Int()
	if errCode != 0 {
		errMsg := gjson.Get(resp, "error_message").String()
		return nil, fmt.Errorf("获取CDN信息失败 [%d]: %s", errCode, errMsg)
	}

	data := gjson.Get(resp, "data")
	cdnInfo := &CdnInfo{
		CdnDns:        data.Get("cdn_dns").String(),
		Vid:           data.Get("vid").String(),
		CorpId:        data.Get("corp_id").String(),
		ClientVersion: data.Get("client_version").String(),
	}

	// 缓存 CDN 信息，3小时过期
	d.cdnInfo = cdnInfo
	d.cdnInfoExpiry = time.Now().Add(3 * time.Hour)

	log.Printf("[CDN] 获取CDN信息成功: vid=%s, corp_id=%s", cdnInfo.Vid, cdnInfo.CorpId)
	return cdnInfo, nil
}

// ==================== 企微 CDN 大文件下载 ====================

// BigDownloadRequest 企微大文件下载请求
type BigDownloadRequest struct {
	CdnDns        string `json:"cdn_dns"`
	Vid           string `json:"vid"`
	CorpId        string `json:"corp_id"`
	ClientVersion string `json:"client_version"`
	FileId        string `json:"file_id"`
	AesKey        string `json:"aes_key"`
	Type          int    `json:"type"` // 2=图片, 4=视频, 6=文件
}

// BigDownloadResponse 企微大文件下载响应
type BigDownloadResponse struct {
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Data         struct {
		DownloadUrl string `json:"download_url"` // 下载URL
		FileSize    int64  `json:"file_size"`    // 文件大小
		FileName    string `json:"file_name"`    // 文件名
	} `json:"data"`
}

// DownloadBigFile 下载大文件（企微 CDN）
func (d *DownloadClient) DownloadBigFile(fileId, aesKey string, fileType int, savePath string) error {
	cdnInfo, err := d.GetCdnInfo()
	if err != nil {
		return err
	}

	reqBody := BigDownloadRequest{
		CdnDns:        cdnInfo.CdnDns,
		Vid:           cdnInfo.Vid,
		CorpId:        cdnInfo.CorpId,
		ClientVersion: cdnInfo.ClientVersion,
		FileId:        fileId,
		AesKey:        aesKey,
		Type:          fileType,
	}

	url := d.cdnURLQiWei + "/cloud/big_download"
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	resp, err := d.httpClient.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	var result BigDownloadResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if result.ErrorCode != 0 {
		return fmt.Errorf("下载失败 [%d]: %s", result.ErrorCode, result.ErrorMessage)
	}

	return d.downloadFile(result.Data.DownloadUrl, savePath)
}

// ==================== 企微 CDN C2C 文件下载 ====================

// C2CDownloadRequest 企微C2C文件下载请求
type C2CDownloadRequest struct {
	CdnDns        string `json:"cdn_dns"`
	Vid           string `json:"vid"`
	CorpId        string `json:"corp_id"`
	ClientVersion string `json:"client_version"`
	FileId        string `json:"file_id"`
	AesKey        string `json:"aes_key"`
	FileType      int    `json:"file_type"`
}

// C2CDownloadResponse 企微C2C文件下载响应
type C2CDownloadResponse struct {
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Data         struct {
		DownloadUrl string `json:"download_url"`
		FileSize    int64  `json:"file_size"`
	} `json:"data"`
}

// DownloadC2CFile 下载C2C文件（企微 CDN）
func (d *DownloadClient) DownloadC2CFile(fileId, aesKey string, fileType int, savePath string) error {
	cdnInfo, err := d.GetCdnInfo()
	if err != nil {
		return err
	}

	reqBody := C2CDownloadRequest{
		CdnDns:        cdnInfo.CdnDns,
		Vid:           cdnInfo.Vid,
		CorpId:        cdnInfo.CorpId,
		ClientVersion: cdnInfo.ClientVersion,
		FileId:        fileId,
		AesKey:        aesKey,
		FileType:      fileType,
	}

	url := d.cdnURLQiWei + "/cloud/c2c_download"
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	resp, err := d.httpClient.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("C2C下载失败，状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	var result C2CDownloadResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if result.ErrorCode != 0 {
		return fmt.Errorf("C2C下载失败 [%d]: %s", result.ErrorCode, result.ErrorMessage)
	}

	return d.downloadFile(result.Data.DownloadUrl, savePath)
}

// ==================== 个微 CDN C2C 文件下载 ====================

// WxDownloadRequest 个微C2C文件下载请求
type WxDownloadRequest struct {
	CdnDns        string `json:"cdn_dns"`
	Vid           string `json:"vid"`
	CorpId        string `json:"corp_id"`
	ClientVersion string `json:"client_version"`
	FileId        string `json:"file_id"`
	AesKey        string `json:"aes_key"`
}

// WxDownloadResponse 个微C2C文件下载响应
type WxDownloadResponse struct {
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Data         struct {
		DownloadUrl string `json:"download_url"`
		FileSize    int64  `json:"file_size"`
	} `json:"data"`
}

// DownloadWxFile 下载微信文件（个微 CDN）
func (d *DownloadClient) DownloadWxFile(fileId, aesKey string, savePath string) error {
	cdnInfo, err := d.GetCdnInfo()
	if err != nil {
		return err
	}

	reqBody := WxDownloadRequest{
		CdnDns:        cdnInfo.CdnDns,
		Vid:           cdnInfo.Vid,
		CorpId:        cdnInfo.CorpId,
		ClientVersion: cdnInfo.ClientVersion,
		FileId:        fileId,
		AesKey:        aesKey,
	}

	url := d.cdnURLWeiXin + "/cloud/c2c_download"
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	resp, err := d.httpClient.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	var result WxDownloadResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	if result.ErrorCode != 0 {
		return fmt.Errorf("下载失败 [%d]: %s", result.ErrorCode, result.ErrorMessage)
	}

	return d.downloadFile(result.Data.DownloadUrl, savePath)
}

// ==================== 统一下载方法 ====================

// isFileIdUrl 检查 FileId 是否为 URL（已经是可直接下载的链接）
func isFileIdUrl(fileId string) bool {
	return strings.HasPrefix(fileId, "http://") || strings.HasPrefix(fileId, "https://")
}

// isBigFile 判断是否需要使用大文件下载接口（根据文件大小）
func isBigFile(fileSize int64) bool {
	// 大于 25MB 使用大文件下载
	return fileSize > 25*1024*1024
}

// DownloadFile 统一文件下载方法（根据协议类型和 file_id 格式自动选择正确的 API）
// fileType: 2=图片, 3=小程序封面, 4=视频, 5=文件/语音
func (d *DownloadClient) DownloadFile(fileId, aesKey string, fileType int, savePath string) error {
	// 检查 FileId 是否已经是 URL
	if isFileIdUrl(fileId) {
		return d.downloadFile(fileId, savePath)
	}

	if d.protocol == ProtocolWeiXin {
		// 个微使用微信 CDN 下载
		return d.DownloadWxFile(fileId, aesKey, savePath)
	} else {
		// 企微：根据文件类型和大小选择下载接口
		// 大文件（视频、大文件）使用 big_download
		if fileType == 4 || fileType == 5 {
			return d.DownloadBigFile(fileId, aesKey, fileType, savePath)
		}
		// 其他文件使用 c2c_download
		return d.DownloadC2CFile(fileId, aesKey, fileType, savePath)
	}
}

// ==================== 辅助方法 ====================

// downloadFile 下载文件到本地
func (d *DownloadClient) downloadFile(url, savePath string) error {
	// 确保目录存在
	dir := filepath.Dir(savePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	resp, err := d.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("下载文件失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	// 创建文件
	file, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	// 写入文件
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	log.Printf("[Download] 文件已保存: %s", savePath)
	return nil
}
