package com.zzk.mcp.model;

import lombok.Data;

@Data  // Lombok 注解，自动生成 getter/setter
public class ColumnSimpleMeta {
    private String columnName;     // 列名
    private String columnComment;  // 列备注
}