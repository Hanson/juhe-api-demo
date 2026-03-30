<?php
/**
 * 企微自动回复示例
 *
 * 企微回调格式：
 * {
 *   "guid": "xxx",
 *   "notify_type": 11010,  // 新消息
 *   "data": {
 *     "content": "消息内容",
 *     "from_id": "发送者ID",
 *     "room_id": "群ID（群消息时）",
 *     "msg_type": 1  // 1=文本消息
 *   }
 * }
 */

// 加载配置
require_once __DIR__ . '/config.php';
require_once __DIR__ . '/api_client.php';

/**
 * 处理企微文本消息
 *
 * @param array $data 回调中的 data 字段
 * @param int $notifyType 回调类型
 * @return bool 是否处理了消息
 */
function handleQiWeiTextMessage(array $data, int $notifyType): bool
{
    // 只处理新消息回调
    if ($notifyType !== 11010) {
        return false;
    }

    // 只处理文本消息（msg_type = 1）
    $msgType = $data['msg_type'] ?? 0;
    if ($msgType !== 1) {
        return false;
    }

    // 获取消息内容
    $content = $data['content'] ?? '';

    // 企微自动回复逻辑：用户发 "123" → 回复 "456"
    if ($content === '123') {
        // 获取发送者 ID（企微用 from_id）
        $fromId = $data['from_id'] ?? '';

        if (empty($fromId)) {
            error_log(date('Y-m-d H:i:s') . " [企微] 无法获取发送者ID\n");
            return true;
        }

        // 判断是群消息还是私聊
        $roomId = $data['room_id'] ?? '';
        $isRoom = !empty($roomId) && $roomId !== '0';

        error_log(date('Y-m-d H:i:s') . " [企微] 收到消息: {$content}, 从: {$fromId}, 群: {$roomId}\n");

        // 企微回复：私聊直接回复 from_id，群聊回复 room_id
        $replyTo = $isRoom ? $roomId : $fromId;

        // 创建 API 客户端并发送回复
        $client = new ApiClient(API_URL, GUID, APP_KEY, APP_SECRET, PROTOCOL);

        try {
            $result = $client->sendText($replyTo, '456');
            error_log(date('Y-m-d H:i:s') . " [企微] 回复发送成功\n");
        } catch (Exception $e) {
            error_log(date('Y-m-d H:i:s') . " [企微] 回复发送失败: " . $e->getMessage() . "\n");
        }

        return true;
    }

    return false;
}
