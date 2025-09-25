package com.zzk.mcp;

import org.mybatis.spring.annotation.MapperScan;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
@MapperScan(basePackages = {"com.zzk.mcp.**.mapper"})
public class AuroaMcpServerApplication {

    public static void main(String[] args) {
        SpringApplication.run(AuroaMcpServerApplication.class, args);
    }

}
