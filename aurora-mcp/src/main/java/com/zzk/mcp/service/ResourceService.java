package com.zzk.mcp.service;

import com.alibaba.fastjson2.JSON;
import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.zzk.mcp.mapper.DatabaseMapper;
import com.zzk.mcp.mapper.StardewAnimalMapper;
import com.zzk.mcp.model.StardewAnimalEntity;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.util.List;

@Slf4j
@RequiredArgsConstructor
@Service
public class ResourceService {

    @Value("${spring.datasource.database}")
    private String dataName;

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
            List<Object> tableInfo = databaseMapper.getTableInfo(tableName);
            String jsonString = JSON.toJSONString(tableInfo);
            log.info("查看表中信息完成{}",jsonString);
            return jsonString;
        }
        return null;
    }
}
