package com.zzk.mcp.controller;

import com.zzk.mcp.model.UserProfile;
import com.zzk.mcp.service.UserProfileService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/profile")
public class UserProfileController {

    @Autowired
    private UserProfileService userProfileService;

    /**
     * 获取指定用户的画像
     */
    @GetMapping("/{userId}")
    public UserProfile getProfile(@PathVariable String userId) {
        return userProfileService.getUserProfile(userId);
    }

    /**
     * 触发画像分析（通常在对话结束后调用，或者异步调用）
     * @param userId 用户ID
     * @param content 对话内容
     */
    @PostMapping("/analyze")
    public UserProfile analyzeProfile(@RequestParam String userId, 
                                      @RequestBody String content) {
        return userProfileService.updateUserProfile(userId, content);
    }
}
