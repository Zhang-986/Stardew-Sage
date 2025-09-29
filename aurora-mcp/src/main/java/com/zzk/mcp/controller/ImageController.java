package com.zzk.mcp.controller;

import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.multipart.MultipartFile;

@RestController
@RequestMapping("/api/image")
public class ImageController {
    @PostMapping("/analyze")
    public String analyzeImage(
            @RequestParam("file") MultipartFile file,
            @RequestParam(value = "question", required = false) String question){
        return "OK";
    }
}
