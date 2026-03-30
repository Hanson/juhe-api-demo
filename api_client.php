<?php
/**
 * API 客户端
 *
 * 通过 /open/GuidRequest 接口调用聚合聊天后端开放平台
 */

class ApiClient
{
    private string $apiUrl;
    private string $guid;
    private string $appKey;
    private string $appSecret;
    private string $protocol;

    public function __construct(string $apiUrl, string $guid, string $appKey, string $appSecret, string $protocol = 'qiwei')
    {
        $this->apiUrl = rtrim($apiUrl, '/');
        $this->guid = $guid;
        $this->appKey = $appKey;
        $this->appSecret = $appSecret;
        $this->protocol = $protocol;
    }

    /**
     * 发送请求到开放平台 /open/GuidRequest
     */
    public function request(string $path, array $data): array
    {
        $url = $this->apiUrl . '/open/GuidRequest';

        $payload = [
            'app_key' => $this->appKey,
            'app_secret' => $this->appSecret,
            'path' => $path,
            'data' => array_merge(['guid' => $this->guid], $data),
        ];

        $ch = curl_init();
        curl_setopt($ch, CURLOPT_URL, $url);
        curl_setopt($ch, CURLOPT_POST, true);
        curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode($payload));
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_HTTPHEADER, ['Content-Type: application/json']);
        curl_setopt($ch, CURLOPT_TIMEOUT, 30);

        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        $error = curl_error($ch);
        curl_close($ch);

        if ($error) {
            throw new Exception("API 请求失败: {$error}");
        }

        if ($httpCode !== 200) {
            throw new Exception("API 返回错误码: {$httpCode}");
        }

        $result = json_decode($response, true);
        if (isset($result['code']) && $result['code'] !== 0) {
            throw new Exception("API 业务错误: " . ($result['message'] ?? '未知错误'));
        }

        return $result;
    }

    /**
     * 发送文本消息
     *
     * @param string $toId 接收者ID
     * @param string $content 消息内容
     * @return array
     */
    public function sendText(string $toId, string $content): array
    {
        if ($this->protocol === 'qiwei') {
            // 企微：使用 conversation_id
            return $this->request('/msg/send_text', [
                'conversation_id' => $toId,
                'content' => $content,
            ]);
        } else {
            // 个微：使用 to_username
            return $this->request('/msg/send_text', [
                'to_username' => $toId,
                'content' => $content,
            ]);
        }
    }

    /**
     * 设置回调地址
     */
    public function setCallback(string $callbackUrl): array
    {
        return $this->request('/callback/set', [
            'callback_url' => $callbackUrl,
        ]);
    }
}
