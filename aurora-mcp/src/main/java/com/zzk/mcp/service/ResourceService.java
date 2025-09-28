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
import java.util.Map;

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


    public List<Map<String, String>> getTableInfo(String tableName) {
        List<String> tablesByDatabase = databaseMapper.getTablesByDatabase(dataName);
        boolean contains = tablesByDatabase.contains(tableName);
        if (contains) {
            List<Map<String, String>> tableInfo = databaseMapper.getTableInfo(tableName, dataName);
            log.debug("tableInfo数据是{}",JSON.toJSONString(tableInfo));
            return tableInfo;
        }
        return null;
    }

    public String getDataInfo(String tableName,String conlums) {
        log.info("conlums内容是{}",conlums);
        List<String> list = JSON.parseArray(conlums, String.class);
        List<Object> dataList = databaseMapper.getDataInfoDetail(tableName,list);
        log.info("dataLIst数据是{}",JSON.toJSONString(dataList));
        return JSON.toJSONString(dataList);
    }
}
