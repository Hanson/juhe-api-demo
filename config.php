<?php
/**
 * 配置文件
 *
 * 从环境变量或 .env 文件加载配置
 */

// 尝试加载 .env 文件
$envFile = __DIR__ . '/.env';
if (file_exists($envFile)) {
    $lines = file($envFile, FILE_IGNORE_NEW_LINES | FILE_SKIP_EMPTY_LINES);
    foreach ($lines as $line) {
        // 跳过注释行
        if (strpos(trim($line), '#') === 0) {
            continue;
        }
        // 解析 KEY=VALUE 格式
        if (strpos($line, '=') !== false) {
            list($key, $value) = explode('=', $line, 2);
            $key = trim($key);
            $value = trim($value);
            // 设置环境变量（如果尚未设置）
            if (!isset($_ENV[$key]) && !isset($_SERVER[$key])) {
                $_ENV[$key] = $value;
                $_SERVER[$key] = $value;
            }
        }
    }
}

// ========== 配置项 ==========

// 协议类型: qiwei(企微/企业微信) 或 weixin(个微/个人微信)
define('PROTOCOL', $_ENV['PROTOCOL'] ?? $_SERVER['PROTOCOL'] ?? 'qiwei');

// 后端开放平台 API 配置
define('API_URL', $_ENV['API_URL'] ?? $_SERVER['API_URL'] ?? 'http://your-api-server:8080');
define('GUID', $_ENV['GUID'] ?? $_SERVER['GUID'] ?? '');

// 开放平台认证配置
define('APP_KEY', $_ENV['APP_KEY'] ?? $_SERVER['APP_KEY'] ?? '');
define('APP_SECRET', $_ENV['APP_SECRET'] ?? $_SERVER['APP_SECRET'] ?? '');

// 回调服务配置
define('CALLBACK_PORT', $_ENV['CALLBACK_PORT'] ?? $_SERVER['CALLBACK_PORT'] ?? '9000');
define('CALLBACK_PATH', $_ENV['CALLBACK_PATH'] ?? $_SERVER['CALLBACK_PATH'] ?? '/callback');

// 回调 URL（公网可访问的地址）
define('CALLBACK_URL', $_ENV['CALLBACK_URL'] ?? $_SERVER['CALLBACK_URL'] ?? 'http://your-public-ip:9000/callback');

// 回调密钥（用于验证回调来源，可选）
define('CALLBACK_SECRET', $_ENV['CALLBACK_SECRET'] ?? $_SERVER['CALLBACK_SECRET'] ?? '');

// ========== 显示配置 ==========

echo "当前配置:\n";
echo "========================================\n";
echo "协议类型: " . (PROTOCOL === 'qiwei' ? '企微（企业微信）' : '个微（个人微信）') . "\n";
echo "API 地址: " . API_URL . "\n";
echo "实例 GUID: " . (GUID ? GUID : '未配置') . "\n";
echo "回调端口: " . CALLBACK_PORT . "\n";
echo "回调路径: " . CALLBACK_PATH . "\n";
echo "回调地址: " . CALLBACK_URL . "\n";
echo "APP_KEY: " . (APP_KEY ? '已配置' : '未配置') . "\n";
echo "APP_SECRET: " . (APP_SECRET ? '已配置' : '未配置') . "\n";
echo "========================================\n\n";
