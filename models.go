package main

// ==================== 消息回调相关 ====================

// CallbackMessage 回调消息结构
type CallbackMessage struct {
	MsgId       string `json:"msg_id"`        // 消息ID
	MsgType     int    `json:"msg_type"`      // 消息类型
	FromId      string `json:"from_id"`       // 发送者ID
	ToId        string `json:"to_id"`         // 接收者ID
	Content     string `json:"content"`       // 消息内容
	RoomId      string `json:"room_id"`       // 群ID（群消息时）
	CreateTime  int64  `json:"create_time"`   // 消息创建时间
	GUID        string `json:"guid"`          // 实例GUID

	// 图片/文件消息相关
	FileId      string `json:"file_id"`       // 文件ID
	AesKey      string `json:"aes_key"`       // AES密钥
	FileSize    int64  `json:"file_size"`     // 文件大小
	FileName    string `json:"file_name"`     // 文件名
	FileType    int    `json:"file_type"`     // 文件类型

	// 其他字段
	AtUserList  []string `json:"at_user_list"` // @用户列表
	IsAtMe      bool     `json:"is_at_me"`     // 是否@自己
}

// 消息类型常量 - 个微（微信）
const (
	// 个微消息类型
	WxMsgTypeText     = 1   // 文本消息
	WxMsgTypeImage    = 3   // 图片消息
	WxMsgTypeVoice    = 34  // 语音消息
	WxMsgTypeFriend   = 37  // 加好友请求
	WxMsgTypeCard     = 42  // 名片消息
	WxMsgTypeVideo    = 43  // 视频消息
	WxMsgTypeEmoji    = 47  // 表情消息
	WxMsgTypeLocation = 48  // 位置消息
	WxMsgTypeApp      = 49  // 应用类型消息（文件、链接、小程序等）
	WxMsgTypeSystem   = 10000 // 系统消息
	WxMsgTypeRevoke   = 10002 // 撤回消息

	// WxMsgTypeApp 子类型（通过 msg.app_type 区分）
	WxAppTypeFile    = 6  // 文件
	WxAppTypeLink    = 5  // 链接
	WxAppTypeWeApp   = 33 // 小程序
	WxAppTypeFinder  = 51 // 视频号
	WxAppTypeRefer   = 57 // 引用消息
)

// 消息类型常量 - 企微（企业微信）
// 注意：企微消息类型与个微不同，具体值请参考企微文档
// 企微使用不同的消息类型编码，回调处理时需要根据协议类型区分
const (
	// 企微消息类型（示例，具体值需要参考企微文档）
	QiWeiMsgTypeText   = 1   // 文本消息
	QiWeiMsgTypeImage  = 2   // 图片消息（企微图片是2）
	QiWeiMsgTypeVoice  = 34  // 语音消息
	QiWeiMsgTypeVideo  = 43  // 视频消息
	QiWeiMsgTypeFile   = 490 // 文件消息
	QiWeiMsgTypeCard   = 42  // 名片消息
	QiWeiMsgTypeLink   = 49  // 链接消息
	QiWeiMsgTypeWeApp  = 49  // 小程序消息
)

// GetMsgTypeInfo 获取消息类型描述
func GetMsgTypeInfo(msgType int, protocol ProtocolType) string {
	if protocol == ProtocolWeiXin {
		switch msgType {
		case WxMsgTypeText:
			return "文本"
		case WxMsgTypeImage:
			return "图片"
		case WxMsgTypeVoice:
			return "语音"
		case WxMsgTypeVideo:
			return "视频"
		case WxMsgTypeCard:
			return "名片"
		case WxMsgTypeEmoji:
			return "表情"
		case WxMsgTypeLocation:
			return "位置"
		case WxMsgTypeApp:
			return "应用(文件/链接/小程序)"
		case WxMsgTypeSystem:
			return "系统消息"
		case WxMsgTypeRevoke:
			return "撤回消息"
		}
	} else {
		switch msgType {
		case QiWeiMsgTypeText:
			return "文本"
		case QiWeiMsgTypeImage:
			return "图片"
		case QiWeiMsgTypeVoice:
			return "语音"
		case QiWeiMsgTypeVideo:
			return "视频"
		case QiWeiMsgTypeFile:
			return "文件"
		case QiWeiMsgTypeCard:
			return "名片"
		case QiWeiMsgTypeLink:
			return "链接/小程序"
		}
	}
	return "未知"
}

// 兼容旧常量（已废弃，建议使用 WxMsgType* 或 QiWeiMsgType*）
const (
	MsgTypeText     = WxMsgTypeText
	MsgTypeImage    = WxMsgTypeImage
	MsgTypeVoice    = WxMsgTypeVoice
	MsgTypeVideo    = WxMsgTypeVideo
	MsgTypeCard     = WxMsgTypeCard
	MsgTypeEmoji    = WxMsgTypeEmoji
	MsgTypeLocation = WxMsgTypeLocation
	MsgTypeFile     = QiWeiMsgTypeFile
	MsgTypeLink     = WxMsgTypeApp
	MsgTypeWeApp    = WxMsgTypeApp
)

// ==================== 发送消息相关 ====================

// SendTextRequest 发送文本消息请求
type SendTextRequest struct {
	GUID      string `json:"guid"`
	ToId      string `json:"to_id"`       // 接收者ID（用户ID或群ID）
	Content   string `json:"content"`     // 消息内容
	AtUserIds string `json:"at_user_ids"` // @用户ID列表（群消息时使用，逗号分隔）
}

// SendImageRequest 发送图片消息请求
type SendImageRequest struct {
	GUID   string `json:"guid"`
	ToId   string `json:"to_id"`     // 接收者ID
	FileId string `json:"file_id"`   // 文件ID（需要先上传获取）
	AesKey string `json:"aes_key"`   // AES密钥
	Size   int64  `json:"size"`      // 文件大小
}

// SendWeAppRequest 发送小程序消息请求
type SendWeAppRequest struct {
	GUID         string `json:"guid"`
	ToId         string `json:"to_id"`          // 接收者ID
	AppId        string `json:"app_id"`         // 小程序AppID
	Title        string `json:"title"`          // 小程序标题
	Desc         string `json:"desc"`           // 小程序描述
	PagePath     string `json:"page_path"`      // 小程序页面路径
	IconUrl      string `json:"icon_url"`       // 小程序图标URL
	ShareUrl     string `json:"share_url"`      // 分享URL（可选）
	UserName     string `json:"user_name"`      // 小程序原始ID（可选）
}

// SendFileRequest 发送文件消息请求
type SendFileRequest struct {
	GUID     string `json:"guid"`
	ToId     string `json:"to_id"`      // 接收者ID
	FileId   string `json:"file_id"`    // 文件ID
	AesKey   string `json:"aes_key"`    // AES密钥
	FileName string `json:"file_name"`  // 文件名
	FileSize int64  `json:"file_size"`  // 文件大小
}

// ==================== 回调设置相关 ====================

// NotifyUrlRequest 设置回调URL请求
type NotifyUrlRequest struct {
	GUID      string `json:"guid"`
	NotifyUrl string `json:"notify_url"` // 回调URL
}

// ==================== 文件上传相关 ====================

// BigUploadRequest 大文件上传请求
type BigUploadRequest struct {
	GUID     string `json:"guid"`
	FileUrl  string `json:"file_url"`  // 文件URL（公网可访问）
	FileName string `json:"file_name"` // 文件名
	FileSize int64  `json:"file_size"` // 文件大小
}

// C2CUploadRequest C2C文件上传请求
type C2CUploadRequest struct {
	GUID     string `json:"guid"`
	FileUrl  string `json:"file_url"`  // 文件URL
	FileName string `json:"file_name"` // 文件名
	ToId     string `json:"to_id"`     // 接收者ID
}

// ==================== CDN 相关 ====================

// CDNInfo CDN信息
type CDNInfo struct {
	CDNDns        string `json:"cdn_dns"`         // CDN DNS
	Vid           string `json:"vid"`             // 用户ID
	CorpId        string `json:"corp_id"`         // 企业ID
	ClientVersion string `json:"client_version"`  // 客户端版本
}

// ==================== 用户信息相关 ====================

// UserInfo 用户信息
type UserInfo struct {
	Vid        string `json:"vid"`         // 用户ID
	Name       string `json:"name"`        // 昵称
	Avatar     string `json:"avatar"`      // 头像URL
	CorpId     string `json:"corp_id"`     // 企业ID
	CorpName   string `json:"corp_name"`   // 企业名称
}
