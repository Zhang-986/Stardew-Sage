package com.zzk.mcp; /**
 * @author Andy
 * @date 2025-01-07
 */


import com.baomidou.mybatisplus.annotation.IdType;
import com.baomidou.mybatisplus.core.mapper.BaseMapper;
import com.baomidou.mybatisplus.generator.FastAutoGenerator;
import com.baomidou.mybatisplus.generator.config.TemplateType;
import com.baomidou.mybatisplus.generator.config.rules.DateType;
import com.baomidou.mybatisplus.generator.config.rules.NamingStrategy;
import org.apache.ibatis.annotations.Mapper;

import java.sql.Connection;
import java.sql.DatabaseMetaData;
import java.sql.ResultSet;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

public class MyBatisPlusGenerator {

    public static void main(String[] args) {
        // 自动获取以stardew开头的表名
        List<String> tableNames = getStardewTables();

        if (tableNames.isEmpty()) {
            System.out.println("没有找到以stardew开头的表");
            return;
        }

        System.out.println("找到以下表: " + String.join(", ", tableNames));

        // 创建代码生成工具类
        FastAutoGenerator generator = create(tableNames);
        // 执行生成代码
        generator.execute();
    }

    /**
     * 获取所有以stardew开头的表名
     */
    private static List<String> getStardewTables() {
        List<String> tableNames = new ArrayList<>();
        String url = "jdbc:mysql://117.72.112.62:3306/ry-vue?useSSL=false&serverTimezone=UTC&allowPublicKeyRetrieval=true";
        String username = "root";
        String password = "200522";

        try (Connection connection = java.sql.DriverManager.getConnection(url, username, password)) {
            DatabaseMetaData metaData = connection.getMetaData();
            // 获取所有表，使用"stardew%"作为模式匹配
            ResultSet tables = metaData.getTables(null, null, "stardew%", new String[]{"TABLE"});

            while (tables.next()) {
                String tableName = tables.getString("TABLE_NAME");
                tableNames.add(tableName);
            }
        } catch (Exception e) {
            e.printStackTrace();
        }

        return tableNames;
    }

    private static FastAutoGenerator create(List<String> tableNames) {
        // 数据库连接地址，
        String url = "jdbc:mysql://117.72.112.62:3306/ry-vue?useSSL=false&serverTimezone=UTC&allowPublicKeyRetrieval=true";
        // 数据库用户名
        String name = "root";
        // 数据库密码
        String password = "200522";
        FastAutoGenerator generator = FastAutoGenerator.create(url, name, password)
                // 全局配置
                .globalConfig(builder -> {
                    // 获取生成的代码路径,这里没有写死，是运行时获取的，这样可以防止不同的开发人员来回修改生成路径的问题。
                    String outputDir = System.getProperty("user.dir") + "/aurora-mcp/src/main/java";
                    builder.outputDir(outputDir)
                            .dateType(DateType.ONLY_DATE)
                            // 生成的类注释中的作者名称，为了统一表示，这里写死了
                            .author("ZZK");
                })
                // 生成的代码包路径配置
                .packageConfig(builder -> {
                    // 生成的代包公共路径
                    builder.parent("com.zzk.mcp");
                    // 生成的mapper xml的存放目录，是在parent路径下面的
                    builder.xml("mapper")
                            // 生成的实例类目录
                            .entity("model")
                            // 生成的service目录
                            .service("service")
                            // 生成的ampper目录
                            .mapper("mapper")
                            .controller("controller");

                }).strategyConfig(builder -> {
                    // 添加要生成的的数据库表
                    builder.addInclude(tableNames)
                            // 启用大写模式
                            .enableCapitalMode()
                            .controllerBuilder()
                            .enableHyphenStyle()
                            .enableRestStyle()
                            .formatFileName("%sController");
                    // 配置生成的实体类策略，不生成serialVersionID
                    builder.entityBuilder().disableSerialVersionUID()
                            // 如果数据库表名带下划线，按驼峰命名法
                            .columnNaming(NamingStrategy.underline_to_camel)
                            // 使用lombok
                            .enableLombok()
                            .enableFileOverride()
                            // 标记实例类的主键生成方式，如果插入时没有指定，刚自动分配一个，默认是雪花算法
                            .idType(IdType.NONE)
                            .logicDeletePropertyName("deleted")
                            // 实例类每次生成的时候，覆盖旧的实体类
                            .enableFileOverride()
                            // 指定生成的实体类名称
                            .convertFileName(entityName -> entityName + "Entity")
                            // 指定生成的service接口名称
                            .serviceBuilder().convertServiceFileName(entityName -> "IDao" + entityName + "Service")
                            // 指定生成的serviceImpl的名称
                            .convertServiceImplFileName(entityName -> "Dao" + entityName + "Service");
                    builder.mapperBuilder()
                            .enableFileOverride()
                            // 指定生成的mapper父类
                            .superClass(BaseMapper.class);
                }).templateConfig(builder -> {
                    // 不生成Controller
                    builder.disable(TemplateType.CONTROLLER);
                    builder.disable(TemplateType.SERVICE);
                    builder.disable(TemplateType.SERVICE_IMPL);
                });
        return generator;
    }

}

