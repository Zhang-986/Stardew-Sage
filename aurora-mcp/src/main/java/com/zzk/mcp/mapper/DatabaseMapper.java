package com.zzk.mcp.mapper;

import com.zzk.mcp.model.ColumnSimpleMeta;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;
import java.util.Map;

public interface DatabaseMapper {


    @Select("SELECT table_name FROM information_schema.tables WHERE table_schema = #{databaseName}")
    List<String> getTablesByDatabase(@Param("databaseName") String databaseName);



    @Select("SELECT * FROM ${tableName}")
    List<Map<String, Object>> getTableInfo(@Param("tableName") String tableName);



    @Select("SELECT * FROM ${tableName} LIMIT 2")
    List<Map<String, Object>> getSampleData(@Param("tableName") String tableName);

    @Select("""
        SELECT 
            COLUMN_NAME AS columnName,
            COLUMN_COMMENT AS columnComment
        FROM 
            information_schema.COLUMNS 
        WHERE 
            TABLE_SCHEMA = DATABASE() 
            AND TABLE_NAME = #{tableName}
        ORDER BY 
            ORDINAL_POSITION
        """)
    List<ColumnSimpleMeta> getColumnSimpleMetas(@Param("tableName") String tableName);

    @Select({
            "<script>",
            "SELECT ",
            "<if test='columnList != null and columnList.size() > 0'>",
            "<foreach item='column' collection='columnList' separator=','>",
            "${column}",
            "</foreach>",
            "</if>",
            "<if test='columnList == null or columnList.size() == 0'>",
            "*",
            "</if>",
            "FROM ${tableName}",
            "</script>"
    })
    List<Map<String, Object>> getRAGDataInfo(
            @Param("columnList") List<String> columnList,
            @Param("tableName") String tableName
    );
}