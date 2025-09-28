package com.zzk.mcp.mapper;

import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;

import java.util.List;

public interface DatabaseMapper {


    @Select("SELECT table_name FROM information_schema.tables WHERE table_schema = #{databaseName}")
    List<String> getTablesByDatabase(@Param("databaseName") String databaseName);



    @Select("SELECT * FROM ${tableName}")
    List<Object> getTableInfo(@Param("tableName") String tableName);


}