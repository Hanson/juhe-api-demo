<?php
/**
 * 个微自动回复示例
 *
 * 个微回调格式：
 * {
 *   "guid": "xxx",
 *   "notify_type": 1010,  // 新消息
 *   "data": {
 *     "content": "消息内容",
 *     "from_username": "wxid_xxx",  // 发送者ID
 *     "room_username": "xxx@chatroom",  // 群ID（群消息时）
 *     "msg_type": 1  // 1=文本消息
 *   }
 * }
 */

// 加载配置
require_once __DIR__ . '/config.php';
require_once __DIR__ . '/api_client.php';

/**
 * 处理个微文本消息
 *
 * @param array $data 回调中的 data 字段
 * @param int $notifyType 回调类型
 * @return bool 是否处理了消息
 */
function handleWeixinTextMessage(array $data, int $notifyType): bool
{
    // 只处理新消息回调
    if ($notifyType !== 1010) {
        return false;
    }

    // 只处理文本消息（msg_type = 1）
    $msgType = $data['msg_type'] ?? 0;
    if ($msgType !== 1) {
        return false;
    }

    // 获取消息内容
    $content = $data['content'] ?? '';

    // 个微自动回复逻辑：用户发 "123" → 回复 "456"
    if ($content === '123') {
        // 获取发送者 ID（个微用 from_username，格式为 wxid_xxx）
        $fromId = $data['from_username'] ?? '';

        if (empty($fromId)) {
            error_log(date('Y-m-d H:i:s') . " [个微] 无法获取发送者ID\n");
            return true;
        }

        // 判断是群消息还是私聊（群消息有 room_username）
        $roomId = $data['room_username'] ?? '';
        $isRoom = !empty($roomId) && strpos($roomId, '@chatroom') !== false;

        error_log(date('Y-m-d H:i:s') . " [个微] 收到消息: {$content}, 从: {$fromId}, 群: {$roomId}\n");

        // 个微回复：私聊直接回复 from_username，群聊回复 room_username
        $replyTo = $isRoom ? $roomId : $fromId;

        // 创建 API 客户端并发送回复
        $client = new ApiClient(API_URL, GUID, APP_KEY, APP_SECRET, PROTOCOL);

        try {
            $result = $client->sendText($replyTo, '456');
            error_log(date('Y-m-d H:i:s') . " [个微] 回复发送成功\n");
        } catch (Exception $e) {
            error_log(date('Y-m-d H:i:s') . " [个微] 回复发送失败: " . $e->getMessage() . "\n");
        }

        return true;
    }

    return false;
}
