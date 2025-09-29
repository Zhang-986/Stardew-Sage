package com.zzk.mcp.controller;

import com.alibaba.fastjson2.JSON;
import com.zzk.mcp.model.ColumnSimpleMeta;
import com.zzk.mcp.model.RagAddVo;
import com.zzk.mcp.service.ResourceService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/table")
public class RagLoadController {

    @Autowired
    private ResourceService resourceService;
    @GetMapping("/info")
    public Map<String, Object> getTableInfo(@RequestParam String tableName) {
        return resourceService.getRAGDataInfo(tableName);
    }

    @PostMapping("/uploadToRAG")
    public String uploadToRAG(@RequestBody RagAddVo ragAddVo) {
        List<ColumnSimpleMeta> columnSimpleMetas = ragAddVo.getColumnSimpleMetas();
        String data = JSON.toJSONString(columnSimpleMetas);
        String tableName = ragAddVo.getTableName();
        return resourceService.uploadToRAG(data,tableName);
    }
}
