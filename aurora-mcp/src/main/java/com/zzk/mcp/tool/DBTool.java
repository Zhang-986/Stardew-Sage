package com.zzk.mcp.tool;

import com.zzk.mcp.service.ResourceService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.ai.tool.annotation.Tool;
import org.springframework.ai.tool.annotation.ToolParam;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;

import java.util.List;

@Slf4j
@Component
public class DBTool {
    @Autowired
    private ResourceService resourceService;
    @Tool(name = "get_all_tables",description = "查看当前库中所有表")
    public List<String> getAllTables(){
        return resourceService.getAllTables();
    }

    @Tool(name = "get_table_info",description = "根据表查看具体信息")
    public String getTableInfo(@ToolParam(description = "表名")String tableName){
        return resourceService.getTableInfo(tableName);
    }


}
