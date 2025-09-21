package com.aurora.dinner.domain;

import org.apache.commons.lang3.builder.ToStringBuilder;
import org.apache.commons.lang3.builder.ToStringStyle;
import com.aurora.common.annotation.Excel;
import com.aurora.common.core.domain.BaseEntity;

/**
 * 午市就餐-桌台对象 lunch_table
 * 
 * @author ruoyi
 * @date 2025-09-19
 */
public class LunchTable extends BaseEntity
{
    private static final long serialVersionUID = 1L;

    /** 桌台号 */
    @Excel(name = "桌台号")
    private Long tableId;

    /** 桌台名称 */
    @Excel(name = "桌台名称")
    private String tableName;

    /** 容纳人数 */
    @Excel(name = "容纳人数")
    private Long capacity;

    /** 桌台状态 (0=空闲, 1=就座) */
    @Excel(name = "桌台状态 (0=空闲, 1=就座)")
    private String status;

    public void setTableId(Long tableId) 
    {
        this.tableId = tableId;
    }

    public Long getTableId() 
    {
        return tableId;
    }

    public void setTableName(String tableName) 
    {
        this.tableName = tableName;
    }

    public String getTableName() 
    {
        return tableName;
    }

    public void setCapacity(Long capacity) 
    {
        this.capacity = capacity;
    }

    public Long getCapacity() 
    {
        return capacity;
    }

    public void setStatus(String status) 
    {
        this.status = status;
    }

    public String getStatus() 
    {
        return status;
    }

    @Override
    public String toString() {
        return new ToStringBuilder(this,ToStringStyle.MULTI_LINE_STYLE)
            .append("tableId", getTableId())
            .append("tableName", getTableName())
            .append("capacity", getCapacity())
            .append("status", getStatus())
            .append("createBy", getCreateBy())
            .append("createTime", getCreateTime())
            .append("updateBy", getUpdateBy())
            .append("updateTime", getUpdateTime())
            .append("remark", getRemark())
            .toString();
    }
}
