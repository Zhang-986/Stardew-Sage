package com.zzk.mcp.model;

import lombok.Data;
import java.util.List;
import java.util.Map;

/**
 * 用户画像模型
 */
@Data
public class UserProfile {
    private String userId;
    private List<String> tags;          // 用户标签 (e.g., "Java开发者", "喜欢科幻")
    private String summary;             // 用户简介/摘要
    private Map<String, Object> preferences; // 具体偏好 (e.g., {"language": "zh", "editor": "vscode"})
    private String lastUpdate;          // 最后更新时间
}
