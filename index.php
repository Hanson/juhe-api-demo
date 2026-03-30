<?php
/**
 * 聚合聊天 - 回调服务
 *
 * 根据 notify_type 自动判断协议类型并处理
 */

// 企微 notify_type 列表
define('QIWEI_NOTIFY_TYPES', [
    11001, 11002, 11003, 11004, 11005, 11006, 11007, 11008, 11009,
    11010, 11011, 11012, 11013,
    2131, 2132,
    1001, 1023, 1037, 2118,
    1002, 1003, 1004, 1005, 1006,
]);

// 个微 notify_type 列表
define('WEIXIN_NOTIFY_TYPES', [
    1001, 1002, 1003, 1004, 1005, 1006, 1007, 1008, 1009, 1010, 1011,
    1039, 1040, 1041, 1042, 1043, 1044, 1045, 1046, 1047, 1048,
    1200, 1201, 1202, 1203, 2006,
]);

// 加载配置
require_once __DIR__ . '/config.php';
require_once __DIR__ . '/reply_qiwei.php';
require_once __DIR__ . '/reply_weixin.php';

// 设置时区
date_default_timezone_set('Asia/Shanghai');

// 记录请求
error_log(date('Y-m-d H:i:s') . " [收到请求] " . $_SERVER['REQUEST_METHOD'] . " " . $_SERVER['REQUEST_URI'] . "\n");

// 获取原始请求体
$rawInput = file_get_contents('php://input');
error_log(date('Y-m-d H:i:s') . " [Raw Body] " . $rawInput . "\n");

// 解析 JSON
$data = json_decode($rawInput, true);

if ($data === null) {
    error_log(date('Y-m-d H:i:s') . " [错误] JSON 解析失败\n");
    http_response_code(200);
    header('Content-Type: application/json');
    echo json_encode(['code' => 400, 'message' => 'invalid json']);
    exit;
}

// 解析回调数据
$guid = $data['guid'] ?? '';
$notifyType = $data['notify_type'] ?? 0;
$msgData = $data['data'] ?? [];

error_log(date('Y-m-d H:i:s') . " [回调] guid={$guid}, notify_type={$notifyType}\n");

// 根据 notify_type 判断协议类型
$isQiWei = in_array($notifyType, QIWEI_NOTIFY_TYPES);
$isWeixin = in_array($notifyType, WEIXIN_NOTIFY_TYPES);

if ($isQiWei) {
    // 企微回调处理
    switch ($notifyType) {
        case 11010: // NotifyTypeNewMsg - 新消息
            handleQiWeiTextMessage($msgData, $notifyType);
            break;
        case 11003:
            error_log(date('Y-m-d H:i:s') . " [企微] 用户登录事件\n");
            break;
        case 11004:
            error_log(date('Y-m-d H:i:s') . " [企微] 用户登出事件\n");
            break;
        case 11002:
            error_log(date('Y-m-d H:i:s') . " [企微] 登录二维码变化\n");
            break;
        case 2131:
            error_log(date('Y-m-d H:i:s') . " [企微] 好友变更\n");
            break;
        case 2132:
            error_log(date('Y-m-d H:i:s') . " [企微] 好友申请\n");
            break;
        case 1002:
            error_log(date('Y-m-d H:i:s') . " [企微] 群成员加入\n");
            break;
        case 1003:
            error_log(date('Y-m-d H:i:s') . " [企微] 群成员退出\n");
            break;
        case 1004:
            error_log(date('Y-m-d H:i:s') . " [企微] 群踢人\n");
            break;
        case 1005:
            error_log(date('Y-m-d H:i:s') . " [企微] 群退群\n");
            break;
        case 1006:
            error_log(date('Y-m-d H:i:s') . " [企微] 群创建\n");
            break;
        default:
            error_log(date('Y-m-d H:i:s') . " [企微] 未处理的 notify_type: {$notifyType}\n");
    }
} elseif ($isWeixin) {
    // 个微回调处理
    switch ($notifyType) {
        case 1010: // WxNotifyTypeNewMsg - 新消息
            handleWeixinTextMessage($msgData, $notifyType);
            break;
        case 1003:
            error_log(date('Y-m-d H:i:s') . " [个微] 用户登录事件\n");
            break;
        case 1004:
            error_log(date('Y-m-d H:i:s') . " [个微] 用户登出事件\n");
            break;
        case 1002:
            error_log(date('Y-m-d H:i:s') . " [个微] 登录二维码变化\n");
            break;
        case 1040:
            error_log(date('Y-m-d H:i:s') . " [个微] 好友添加\n");
            break;
        case 1041:
            error_log(date('Y-m-d H:i:s') . " [个微] 好友删除\n");
            break;
        case 1046:
            error_log(date('Y-m-d H:i:s') . " [个微] 群成员加入\n");
            break;
        case 1048:
            error_log(date('Y-m-d H:i:s') . " [个微] 群成员退出\n");
            break;
        default:
            error_log(date('Y-m-d H:i:s') . " [个微] 未处理的 notify_type: {$notifyType}\n");
    }
} else {
    error_log(date('Y-m-d H:i:s') . " [未知] 无法识别的 notify_type: {$notifyType}\n");
}

// 返回成功响应
http_response_code(200);
header('Content-Type: application/json');
echo json_encode(['code' => 0, 'message' => 'ok']);
