package com.zzk.mcp.model;

import lombok.Data;

import java.util.List;

@Data
public class RagAddVo {
    private String tableName;
    private List<ColumnSimpleMeta> columnSimpleMetas;
}
