package com.zzk.mcp.controller;

import com.alibaba.fastjson2.JSON;
import com.zzk.mcp.model.ColumnSimpleMeta;
import com.zzk.mcp.model.RagAddVo;
import com.zzk.mcp.service.ResourceService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.data.redis.connection.DataType;
import org.springframework.data.redis.core.Cursor;
import org.springframework.data.redis.core.RedisCallback;
import org.springframework.data.redis.core.RedisTemplate;
import org.springframework.data.redis.core.ScanOptions;
import org.springframework.web.bind.annotation.*;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/table")
public class RagLoadController {

    @Autowired
    private RedisTemplate<String,String> redisTemplate;

    @Autowired
    private ResourceService resourceService;

    @GetMapping()
    public List<String> getTable() {
        return resourceService.getAllTables();
    }

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

    /**
     * 按前缀分组统计default命名空间下的键数量
     * 例如：default:stardew_animal:*, default:stardew_build:* 等
     * 返回格式：{"stardew_animal": 15, "stardew_build": 29, "default": 44}
     */
    @GetMapping("/default/keys/grouped-count")
    public Map<String, Long> countDefaultKeysByGroup() {
        Map<String, Long> groupedResult = new HashMap<>();

        try {
            // 扫描所有default:开头的键
            String pattern = "default:*";
            Cursor<byte[]> cursor = redisTemplate.execute((RedisCallback<Cursor<byte[]>>) connection ->
                    connection.scan(ScanOptions.scanOptions().match(pattern).count(100).build())
            );

            if (cursor != null) {
                while (cursor.hasNext()) {
                    String key = new String(cursor.next());

                    // 提取分类前缀（如stardew_animal, stardew_build）
                    String category = extractKeyCategory(key);

                    // 每个键算作一个元素（因为您要统计键的数量）
                    groupedResult.merge(category, 1L, Long::sum);
                }
                cursor.close();
            }

        } catch (Exception e) {
            System.err.println("分组统计Redis键数量时发生错误: " + e.getMessage());
        }

        return groupedResult;
    }

    /**
     * 提取键的分类名称
     * default:stardew_animal:0 -> stardew_animal
     * default:stardew_animal:1 -> stardew_animal
     * default:stardew_build:22 -> stardew_build
     * default:other_key -> other_key
     */
    private String extractKeyCategory(String key) {
        if (key == null || !key.startsWith("default:")) {
            return "other";
        }

        // 移除default:前缀
        String withoutPrefix = key.substring(8); // "default:".length() = 8

        // 查找第二个冒号的位置（如果有的话）
        int secondColonIndex = withoutPrefix.indexOf(":");
        if (secondColonIndex > 0) {
            // 有第二个冒号，返回中间部分：stardew_animal
            return withoutPrefix.substring(0, secondColonIndex);
        } else {
            // 没有第二个冒号，返回剩余部分
            return withoutPrefix;
        }
    }



    @GetMapping("/stats/simple")
    public Map<String, Object> getTableStatsSimple(@RequestParam String tableName) {
        Map<String, Object> result = new HashMap<>();

        try {
            String pattern = "default:" + tableName + ":*";
            List<String> keys = new ArrayList<>();

            // 扫描键
            ScanOptions options = ScanOptions.scanOptions().match(pattern).count(100).build();
            Cursor<String> cursor = redisTemplate.scan(options);
            while (cursor.hasNext()) {
                keys.add(cursor.next());
            }
            cursor.close();

            // 直接用 JSON.GET 读取，不判断类型
            Object sampleData = null;
            if (!keys.isEmpty()) {
                sampleData = redisTemplate.execute((RedisCallback<String>) connection -> {
                    byte[] rawValue = (byte[]) connection.execute("JSON.GET", keys.get(0).getBytes());
                    return rawValue != null ? new String(rawValue) : null;
                });

                if (sampleData != null) {
                    try {
                        sampleData = JSON.parseObject(sampleData.toString());
                    } catch (Exception e) {
                        // 保留原始字符串
                    }
                }
            }

            result.put("tableName", tableName);
            result.put("keyCount", keys.size());
            result.put("sampleData", sampleData);

        } catch (Exception e) {
            result.put("error", e.getMessage());
            e.printStackTrace();
        }

        return result;
    }
}
