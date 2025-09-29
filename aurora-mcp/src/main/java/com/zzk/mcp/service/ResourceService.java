package com.zzk.mcp.service;

import com.alibaba.fastjson2.JSON;
import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.core.toolkit.CollectionUtils;
import com.zzk.mcp.mapper.DatabaseMapper;
import com.zzk.mcp.mapper.StardewAnimalMapper;
import com.zzk.mcp.model.ColumnSimpleMeta;
import com.zzk.mcp.model.StardewAnimalEntity;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.ai.document.Document;
import org.springframework.ai.vectorstore.VectorStore;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Slf4j
@RequiredArgsConstructor
@Service
public class ResourceService {

    @Value("${spring.datasource.database}")
    private String dataName;

    private final VectorStore vectorStore;

    private final DatabaseMapper databaseMapper;

    private final StardewAnimalMapper stardewAnimalMapper;

    public List<StardewAnimalEntity> getAnimalInfo() {
        LambdaQueryWrapper<StardewAnimalEntity> stardewAnimalEntityLambdaQueryWrapper = new LambdaQueryWrapper<>();
        stardewAnimalEntityLambdaQueryWrapper.select(StardewAnimalEntity::getNameCh,
                StardewAnimalEntity::getBuyPrice,
                StardewAnimalEntity::getSellPrice,
                StardewAnimalEntity::getNeedBuild,
                StardewAnimalEntity::getAnimalOutput,
                StardewAnimalEntity::getRemark);
        return stardewAnimalMapper.selectList(stardewAnimalEntityLambdaQueryWrapper);
    }


    public List<String> getAllTables() {
        return databaseMapper.getTablesByDatabase(dataName);
    }


    public String getTableInfo(String tableName) {
        List<String> tablesByDatabase = databaseMapper.getTablesByDatabase(dataName);
        boolean contains = tablesByDatabase.contains(tableName);
        if (contains) {
            log.info("查看表中信息：{}", tableName);
            List<Map<String, Object>> tableInfo = databaseMapper.getTableInfo(tableName);
            String jsonString = JSON.toJSONString(tableInfo);
            log.info("查看表中信息完成{}",jsonString);
            return jsonString;
        }
        return null;
    }

    public Map<String,Object> getRAGDataInfo(String tableName){
        Map<String, Object> result = new HashMap<>();
        result.put("column_data", databaseMapper.getColumnSimpleMetas(tableName));
        result.put("sample_data", databaseMapper.getSampleData(tableName));
        return result;
    }
    @Transactional(rollbackFor = Exception.class)
    public String uploadToRAG(String data, String tableName) {
        List<ColumnSimpleMeta> columnSimpleMetas = JSON.parseArray(data, ColumnSimpleMeta.class);
        if (CollectionUtils.isNotEmpty(columnSimpleMetas)) {
            List<String> columnList = columnSimpleMetas.stream().map(ColumnSimpleMeta::getColumnName).toList();
            try {
                // 1. 查询数据
                List<Map<String, Object>> tableInfo = databaseMapper.getRAGDataInfo(columnList, tableName);

                if (CollectionUtils.isEmpty(tableInfo)) {
                    System.out.println("表 " + tableName + " 无数据");
                    return "OK";
                }

                System.out.println("查询到数据行数: " + tableInfo.size());

                // 2. 分批处理
                int batchSize = 50;
                int totalBatches = (int) Math.ceil((double) tableInfo.size() / batchSize);
                int totalProcessed = 0; // 记录已处理总数

                for (int i = 0; i < tableInfo.size(); i += batchSize) {
                    int end = Math.min(i + batchSize, tableInfo.size());
                    List<Map<String, Object>> batch = tableInfo.subList(i, end);

                    List<Document> documentList = new ArrayList<>();

                    // 为批次内的每个文档生成唯一ID
                    for (int j = 0; j < batch.size(); j++) {
                        Map<String, Object> item = batch.get(j);
                        int globalIndex = i + j; // 全局唯一索引：0, 1, 2, ..., 120
                        Document doc = createDocument(item, tableName, globalIndex);
                        documentList.add(doc);
                    }

                    // 3. 插入向量数据库
                    vectorStore.add(documentList);
                    totalProcessed += documentList.size();


                    // 4. 短暂延迟
                    Thread.sleep(50);
                }

                System.out.println("表 " + tableName + " 数据处理完成，总计 " + totalProcessed + " 条记录");

            } catch (Exception e) {
                System.err.println("处理表 " + tableName + " 时出错: " + e.getMessage());
                e.printStackTrace();
                return "ERROR: " + e.getMessage();
            }
        }
        return "OK";
    }

    private Document createDocument(Map<String, Object> item, String tableName, int globalIndex) {
        return Document.builder()
                .text(JSON.toJSONString(item))
                .metadata(item)
                .id(tableName + ":" + globalIndex) // 每个文档都有唯一ID
                .build();
    }
}
