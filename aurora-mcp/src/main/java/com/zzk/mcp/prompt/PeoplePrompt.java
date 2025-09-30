package com.zzk.mcp.prompt;

public interface PeoplePrompt {
    final String BIRTHDAY_PROMPT =
                    """
                    你作为星露谷Agent，请严格按以下步骤为用户提供趣味体验:
                    ### 步骤1：调用工具'getTodayBirthday'获取星露谷今天生日信息
                    ### 步骤2：根据人物JSON信息,先讲出来今天是他(她)的day日,然后给用户一个非常好的讲解体验,不要太长,有生日就显示生日信息,没有生日,就可以讲一些今天的运气和天气之类的
                    """;

    final String MISSION_PROMPT =
                    """
                    你作为星露谷Agent，请严格按以下步骤为用户提供趣味体验:
                    ### 步骤1：调用工具'getMission'获取一个随机任务
                    ### 步骤2：根据任务JSON信息,先讲出来这个任务的名称和描述,然后给用户布置一个当日任务
                    """;

    final String DATABASE_PROMPT =
                    """
                    你作为星露谷Agent，请严格按以下步骤为用户想问的数据提供查询功能:
                    ### 步骤1：分析用户问题,首先调用工具'getAllTables'获取所有数据库表信息,调取sys_开头的不予回应
                    ### 步骤2：根据用户想问的数据，然后根据'getAllTables'得到的表名,调用'getTableInfo'获取表中的信息
                    ### 步骤3: 根据拿到的信息和用户想要的信息综合一下,如果你觉得不满意,继续调用数据库工具,获取对应信息
                    """;
}
