table和映射的关系
1. 如果调用了Table() 并同时使用了映射
    1. 如果映射定义了Table() 则使用预定义的table
    2. 如果映射没有定义Table() -> 使用预定义的table, 如果预定义的table也没有 
        则解析struct名
    最终查询：
    1. 如果调用了session.Table() 则使用table和alias 否则使用映射后的