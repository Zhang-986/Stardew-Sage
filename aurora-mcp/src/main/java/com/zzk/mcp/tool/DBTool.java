package com.zzk.mcp.tool;

import com.zzk.mcp.service.ResourceService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.ai.tool.annotation.Tool;
import org.springframework.ai.tool.annotation.ToolParam;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.Map;

@Slf4j
@Component
public class DBTool {
    @Autowired
    private ResourceService resourceService;
    @Tool(name = "getAllTables",description = "查看当前库中所有表")
    public List<String> getAllTables(){
        return resourceService.getAllTables();
    }

    @Tool(name = "getTableInfo",description = "查看当前表中的备注信息")
    public List<Map<String,String>> getTableInfo(@ToolParam(description = "表名")String tableName){
        return resourceService.getTableInfo(tableName);
    }

    @Tool(name = "getDataInfo",description = "获取数据信息")
    public String getDataInfo(@ToolParam(description = "表名")String tableName,
                              @ToolParam(description = "查询的列信息")String conlums){
        return resourceService.getDataInfo(tableName,conlums);
    }
}
