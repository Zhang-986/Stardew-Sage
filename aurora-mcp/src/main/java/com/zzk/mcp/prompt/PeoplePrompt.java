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
}
