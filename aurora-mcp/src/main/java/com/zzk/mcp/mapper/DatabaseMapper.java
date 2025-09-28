package com.zzk.mcp.mapper;

import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;
import java.util.Map;

public interface DatabaseMapper {


    @Select("SELECT table_name FROM information_schema.tables WHERE table_schema = #{databaseName}")
    List<String> getTablesByDatabase(@Param("databaseName") String databaseName);



    @Select("SELECT COLUMN_NAME AS columnName, COLUMN_COMMENT AS columnComment " +
            "FROM INFORMATION_SCHEMA.COLUMNS " +
            "WHERE TABLE_SCHEMA = #{dataName} AND TABLE_NAME = #{tableName}")
    List<Map<String, String>> getTableInfo(@Param("tableName") String tableName,
                                           @Param("dataName") String dataName);



    List<Object> getDataInfoDetail(String tableName, List<String> conlums);
}